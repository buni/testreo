//nolint:ireturn
package pgxtx

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Query interface {
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
}

type QueryRow interface {
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TxWrapper struct {
	*pgxpool.Pool
	txOptions pgx.TxOptions
}

// NewTxWrapper returns a new TxWrapper.
func NewTxWrapper(db *pgxpool.Pool, txOptions pgx.TxOptions) *TxWrapper {
	return &TxWrapper{
		Pool:      db,
		txOptions: txOptions,
	}
}

// Exec...
func (db *TxWrapper) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	ctx, tx, ok, err := GetTxOrCreate(ctx, db.Pool, db.txOptions)
	if err != nil {
		return pgconn.CommandTag{}, fmt.Errorf("failed to get or create tx: %w", err)
	}

	if ok {
		defer func() {
			commitErr := tx.Commit(ctx)
			if commitErr != nil {
				err = errors.Join(err, commitErr)
			}
		}()
	}

	return tx.Exec(ctx, query, args...) //nolint:wrapcheck
}

// Query ...
func (db *TxWrapper) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	var tx Query
	tx = db.Pool

	ctxTx := FromContext(ctx)
	if ctxTx != nil {
		tx = ctxTx
	}

	return tx.Query(ctx, query, args...) //nolint:wrapcheck
}

// QueryRow ...
func (db *TxWrapper) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	var tx QueryRow
	tx = db.Pool

	ctxTx := FromContext(ctx)
	if ctxTx != nil {
		tx = ctxTx
	}

	return tx.QueryRow(ctx, query, args...) //nolint:wrapcheck
}
