package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/buni/wallet/internal/pkg/database/pgxtx"
	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/iZettle/structextract"
)

var ErrDatabaseConnectionNotSupplied = errors.New("database connection not supplied")

const (
	db    = "db"
	table = "outbox_messages"

	MessageStatusQueued = "queued"
	MessageStatusSent   = "published"
)

type Repository interface {
	Create(ctx context.Context, msg Message) error
	Update(ctx context.Context, msg Message) error
	List(ctx context.Context, limit uint64, status, publisherType string, publishAt time.Time) ([]Message, error)
}

type Message struct {
	ID             string         `db:"id"`
	Payload        pubsub.Message `db:"payload"`
	PublisherType  string         `db:"publisher_type"`
	PublishOptions []byte         `db:"publisher_options"`
	Status         string         `db:"status"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func NewMessage(msg *pubsub.Message, publisherType string, status string) (Message, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Message{}, fmt.Errorf("failed to generate uuid: %w", err)
	}

	timeUTC := time.Now().UTC().Truncate(time.Microsecond) // truncate to microsecond to avoid rounding errors when inserting/returning from db

	return Message{
		ID:             id.String(),
		Payload:        *msg,
		PublisherType:  publisherType,
		PublishOptions: []byte("{}"),
		Status:         status,
		CreatedAt:      timeUTC,
		UpdatedAt:      timeUTC,
	}, nil
}

type PostgresRepository[Querier any] struct {
	pgxpool *pgxtx.TxWrapper
	querier Querier
	scanOne func(ctx context.Context, db Querier, dst any, query string, args ...any) error
	scanAll func(ctx context.Context, db Querier, dst any, query string, args ...any) error
	table   string
}

func NewPGxRepository(pgxpool *pgxtx.TxWrapper) *PostgresRepository[pgxscan.Querier] {
	return &PostgresRepository[pgxscan.Querier]{
		pgxpool: pgxpool,
		querier: pgxpool,
		scanOne: pgxscan.Get,
		scanAll: pgxscan.Select,
		table:   table,
	}
}

func (r *PostgresRepository[Querier]) Migrate(ctx context.Context) error {
	migration := `
	CREATE TABLE IF NOT EXISTS outbox_messages (
		id uuid NOT NULL PRIMARY KEY,
		payload jsonb NOT NULL,
		publisher_type text NOT NULL,
		publisher_options jsonb NOT NULL,
		status text NOT NULL,
		created_at timestamp NOT NULL,
		updated_at timestamp NOT NULL);
		CREATE INDEX IF NOT EXISTS idx_outbox_messages_status_publisher_type ON outbox_messages (status,publisher_type);
		`

	_, err := r.pgxpool.Exec(ctx, migration)
	if err != nil {
		return fmt.Errorf("failed to create outbox_messages table: %w", err)
	}

	return nil
}

func (r *PostgresRepository[Querier]) Create(ctx context.Context, msg Message) (err error) {
	fvMap, err := structextract.New(&msg).FieldValueFromTagMap("db")
	if err != nil {
		return fmt.Errorf("failed to extract field value map: %w", err)
	}

	query, args, err := sq.Insert(r.table).SetMap(fvMap).Suffix("RETURNING updated_at").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	err = r.scanOne(ctx, r.querier, &msg, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	return nil
}

func (r *PostgresRepository[Querier]) Update(ctx context.Context, msg Message) (err error) {
	fvMap, err := structextract.New(&msg).FieldValueFromTagMap(db)
	if err != nil {
		return fmt.Errorf("failed to extract field value map: %w", err)
	}

	query, args, err := sq.Update(r.table).SetMap(fvMap).Where(sq.Eq{"id": msg.ID}).Suffix("RETURNING updated_at").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	err = r.scanOne(ctx, r.querier, &msg, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	return nil
}

func (r *PostgresRepository[Querier]) List(ctx context.Context, limit uint64, status, publisherType string, publishAt time.Time) (result []Message, err error) {
	columns, err := structextract.New(&Message{}).NamesFromTag(db) //nolint:exhaustruct
	if err != nil {
		return nil, fmt.Errorf("failed to extract columns: %w", err)
	}

	query, args, err := sq.Select(columns...).From(r.table).Where(sq.And{sq.Eq{"status": status}, sq.Eq{"publisher_type": publisherType}}).Limit(limit).Suffix("FOR UPDATE SKIP LOCKED").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	err = r.scanAll(ctx, r.querier, &result, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}

	return result, nil
}
