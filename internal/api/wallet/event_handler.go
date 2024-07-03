package wallet

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/buni/wallet/internal/api/app/contract"
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/pkg/database"
	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/buni/wallet/internal/pkg/sloglog"
)

type EventCreatedHandler struct {
	svc contract.WalletService
	txm database.TransactionManager
}

func NewWalletEventCreatedHandler(
	svc contract.WalletService,
	txm database.TransactionManager,
) *EventCreatedHandler {
	return &EventCreatedHandler{
		txm: txm,
		svc: svc,
	}
}

func (h *EventCreatedHandler) HandlerName() string {
	return "WalletEventCreatedHandler"
}

func (h *EventCreatedHandler) Topic() string {
	return entity.WalletEventsTopic + "." + entity.WalletEventsCreated
}

func (h *EventCreatedHandler) SubscriberOptions() []pubsub.SubscriberOption {
	return []pubsub.SubscriberOption{}
}

func (h *EventCreatedHandler) Handle(ctx context.Context, event *entity.WalletEvent, _ pubsub.SubscriberMessage) (err error) {
	logger := sloglog.FromContext(ctx)

	logger.InfoContext(ctx, "received event", slog.Any("event", event))

	err = h.txm.Run(ctx, func(ctx context.Context) error {
		_, err = h.svc.RebuildWalletProjection(ctx, event)
		if err != nil {
			return fmt.Errorf("failed to rebuild wallet projection: %w", err)
		}
		return nil
	})
	if err != nil {
		return
	}

	return nil
}
