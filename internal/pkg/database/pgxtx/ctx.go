//nolint:ireturn
package pgxtx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

func ToContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func FromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey{}).(pgx.Tx) //nolint:revive
	return tx
}

func GetTxOrCreate(ctx context.Context, db *pgxpool.Pool, opts pgx.TxOptions) (context.Context, pgx.Tx, bool, error) {
	tx := FromContext(ctx)
	if tx != nil {
		return ctx, tx, false, nil
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, tx, false, fmt.Errorf("failed to create transaction: %w", err)
	}

	return ToContext(ctx, tx), tx, true, nil
}
