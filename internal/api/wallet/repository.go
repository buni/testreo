package wallet

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/buni/wallet/internal/api/app/contract"
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/pkg/database/pgxtx"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/iZettle/structextract"
)

const (
	db = "db"
)

var _ contract.WalletRepository = (*Repository)(nil)

type Repository struct {
	pgxpool *pgxtx.TxWrapper
	table   string
}

func NewRepository(pgxpool *pgxtx.TxWrapper) *Repository {
	return &Repository{
		pgxpool: pgxpool,
		table:   "wallets",
	}
}

func (r *Repository) Create(ctx context.Context, wallet entity.Wallet) (entity.Wallet, error) {
	fvMap, err := structextract.New(&wallet).FieldValueFromTagMap(db)
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to extract field value map: %w", err)
	}

	query, args, err := sq.Insert(r.table).SetMap(fvMap).Suffix("RETURNING created_at").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to build insert query: %w", err)
	}

	err = pgxscan.Get(ctx, r.pgxpool, &wallet, query, args...)
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return wallet, nil
}

func (r *Repository) Get(ctx context.Context, id string) (result entity.Wallet, err error) {
	columns, err := structextract.New(&entity.Wallet{}).NamesFromTag(db)
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to extract columns: %w", err)
	}

	query, args, err := sq.Select(columns...).From(r.table).PlaceholderFormat(sq.Dollar).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to build select query: %w", err)
	}

	err = pgxscan.Get(ctx, r.pgxpool, &result, query, args...)
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to execute select query: %w", err)
	}

	return result, nil
}

var _ contract.WalletEventRepository = (*EventRepository)(nil)

type EventRepository struct {
	pgxpool *pgxtx.TxWrapper
	table   string
}

func NewEventRepository(pgxpool *pgxtx.TxWrapper) *EventRepository {
	return &EventRepository{
		pgxpool: pgxpool,
		table:   "wallet_events",
	}
}

func (r *EventRepository) Create(ctx context.Context, event entity.WalletEvent) (entity.WalletEvent, error) {
	fvMap, err := structextract.New(&event).FieldValueFromTagMap(db)
	if err != nil {
		return entity.WalletEvent{}, fmt.Errorf("failed to extract field value map: %w", err)
	}

	query, args, err := sq.Insert(r.table).SetMap(fvMap).Suffix("RETURNING created_at").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return entity.WalletEvent{}, fmt.Errorf("failed to build insert query: %w", err)
	}

	err = pgxscan.Get(ctx, r.pgxpool, &event, query, args...)
	if err != nil {
		return entity.WalletEvent{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return event, nil
}

func (r *EventRepository) ListByWalletID(ctx context.Context, walletID string) (result []entity.WalletEvent, err error) {
	columns, err := structextract.New(&entity.WalletEvent{}).NamesFromTag(db)
	if err != nil {
		return nil, fmt.Errorf("failed to extract columns: %w", err)
	}

	query, args, err := sq.Select(columns...).From(r.table).PlaceholderFormat(sq.Dollar).Where(sq.Eq{"wallet_id": walletID}).OrderBy("id ASC").ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	err = pgxscan.Select(ctx, r.pgxpool, &result, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}

	return result, nil
}

var _ contract.WalletProjectionRepository = (*ProjectionRepository)(nil)

type ProjectionRepository struct {
	pgxpool *pgxtx.TxWrapper
	table   string
}

func NewProjectionRepository(pgxpool *pgxtx.TxWrapper) *ProjectionRepository {
	return &ProjectionRepository{
		pgxpool: pgxpool,
		table:   "wallet_projections",
	}
}

func (r *ProjectionRepository) Create(ctx context.Context, projection entity.WalletProjection) (entity.WalletProjection, error) {
	fvMap, err := structextract.New(&projection).FieldValueFromTagMap(db)
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to extract field value map: %w", err)
	}

	query, args, err := sq.Insert(r.table).SetMap(fvMap).Suffix("RETURNING created_at").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to build insert query: %w", err)
	}

	err = pgxscan.Get(ctx, r.pgxpool, &projection, query, args...)
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return projection, nil
}

func (r *ProjectionRepository) Get(ctx context.Context, walletID string) (result entity.WalletProjection, err error) {
	columns, err := structextract.New(&entity.WalletProjection{}).NamesFromTag(db)
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to extract columns: %w", err)
	}

	query, args, err := sq.Select(columns...).From(r.table).PlaceholderFormat(sq.Dollar).Where(sq.Eq{"wallet_id": walletID}).ToSql()
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to build select query: %w", err)
	}

	err = pgxscan.Get(ctx, r.pgxpool, &result, query, args...)
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to execute select query: %w", err)
	}

	return result, nil
}

func (r *ProjectionRepository) Update(ctx context.Context, projection entity.WalletProjection) (entity.WalletProjection, error) {
	query, args, err := sq.Update(r.table).SetMap(map[string]any{
		"balance":        projection.Balance,
		"pending_debit":  projection.PendingDebit,
		"pending_credit": projection.PendingCredit,
		"last_event_id":  projection.LastEventID,
		"updated_at":     projection.UpdatedAt,
	}).Suffix("RETURNING updated_at").PlaceholderFormat(sq.Dollar).Where(sq.Eq{"wallet_id": projection.WalletID}).ToSql()
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to build update query: %w", err)
	}

	err = pgxscan.Get(ctx, r.pgxpool, &projection, query, args...)
	if err != nil {
		return entity.WalletProjection{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return projection, nil
}
