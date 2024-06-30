package pgxtx

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	db        *pgxpool.Pool
	txOptions pgx.TxOptions
}

func NewTransactionManager(db *pgxpool.Pool, txOptions pgx.TxOptions) *TransactionManager {
	return &TransactionManager{db: db, txOptions: txOptions}
}

func (txm *TransactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	ctx, tx, created, err := GetTxOrCreate(ctx, txm.db, txm.txOptions)
	if err != nil {
		return fmt.Errorf("failed to get or create Tx : %w", err)
	}

	err = fn(ctx)
	if err != nil {
		if created {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				return fmt.Errorf("failed to rollback transaction: %w", errors.Join(err, rollbackErr))
			}
		}

		return fmt.Errorf("failed to run transaction: %w", err)
	}

	if created {
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}
