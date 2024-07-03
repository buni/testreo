package testutils

import "context"

// NoopTransactionManager implements database.TransactionManager.
type NoopTransactionManager struct{}

func (n NoopTransactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
