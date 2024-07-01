package wallet

import (
	"context"
	"fmt"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/pkg/pubsub"
)

type Publisher struct {
	publisher pubsub.Publisher
}

func NewPublisher(publisher pubsub.Publisher) *Publisher {
	return &Publisher{
		publisher: publisher,
	}
}

func (p *Publisher) PublishCreated(ctx context.Context, event entity.WalletEvent) error {
	msg, err := pubsub.NewJSONMessage(event, nil)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	msg.Key = entity.WalletEventsCreated
	msg.Topic = entity.WalletEventsTopic

	err = p.publisher.Publish(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
