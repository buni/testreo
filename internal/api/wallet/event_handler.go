package wallet

import (
	"context"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/pkg/database"
	"github.com/buni/wallet/internal/pkg/pubsub"
)

type WalletEventCreatedHandler struct { //nolint:revive
	txm database.TransactionManager
}

func NewAuctionBidedHandler(
	txm database.TransactionManager,
) *WalletEventCreatedHandler {
	return &WalletEventCreatedHandler{
		txm: txm,
	}
}

func (h *WalletEventCreatedHandler) HandlerName() string {
	return "AuctionBidedHandler"
}

func (h *WalletEventCreatedHandler) Topic() string {
	return entity.WalletEventsTopic + "." + entity.WalletEventsCreated
}

func (h *WalletEventCreatedHandler) Handle(ctx context.Context, event *entity.WalletEvent, _ pubsub.SubscriberMessage) (err error) {
	return nil
}
