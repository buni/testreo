package database

import "context"

type TransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
