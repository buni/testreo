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

type WalletEventCreatedHandler struct { //nolint:revive
	svc contract.WalletService
	txm database.TransactionManager
}

func NewWalletEventCreatedHandler(
	svc contract.WalletService,
	txm database.TransactionManager,
) *WalletEventCreatedHandler {
	return &WalletEventCreatedHandler{
		txm: txm,
		svc: svc,
	}
}

func (h *WalletEventCreatedHandler) HandlerName() string {
	return "WalletEventCreatedHandler"
}

func (h *WalletEventCreatedHandler) Topic() string {
	return entity.WalletEventsTopic + "." + entity.WalletEventsCreated
}

func (h *WalletEventCreatedHandler) SubscriberOptions() []pubsub.SubscriberOption {
	return []pubsub.SubscriberOption{}
}

func (h *WalletEventCreatedHandler) Handle(ctx context.Context, event *entity.WalletEvent, _ pubsub.SubscriberMessage) error {
	logger := sloglog.FromContext(ctx)

	logger.InfoContext(ctx, "received event", slog.Any("event", event))
	err := h.txm.Run(ctx, func(ctx context.Context) error {
		_, err := h.svc.RebuildWalletProjection(ctx, event.WalletID)
		if err != nil {
			return fmt.Errorf("failed to rebuild wallet projection: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to handle message: %w", err)
	}

	return nil
}
