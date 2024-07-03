package outbox

import (
	"context"
	"fmt"

	"github.com/buni/wallet/internal/pkg/database"
	"github.com/buni/wallet/internal/pkg/pubsub"
)

var _ pubsub.Publisher = (*Publisher[string])(nil)

type Publisher[OptionsType any] struct {
	repo          Repository
	txm           database.TransactionManager
	publisherType string
}

func NewPublisher[OptionsType any](
	repo Repository,
	txm database.TransactionManager,
	publisherType string,
) *Publisher[OptionsType] {
	return &Publisher[OptionsType]{
		repo:          repo,
		txm:           txm,
		publisherType: publisherType,
	}
}

func (p *Publisher[OptionsType]) Publish(ctx context.Context, msg *pubsub.Message, _ ...pubsub.PublishOption) (err error) {
	outboxMsg, err := NewMessage(msg, p.publisherType, MessageStatusQueued)
	if err != nil {
		return fmt.Errorf("failed to create outbox message: %w", err)
	}

	err = p.txm.Run(ctx, func(ctx context.Context) error {
		err = p.repo.Create(ctx, outboxMsg)
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to run transaction: %w", err)
	}

	return nil
}
