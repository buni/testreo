package response

import (
	"time"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID            string          `json:"id"`
	ReferenceID   string          `json:"reference_id"`
	Balance       decimal.Decimal `json:"balance"`
	PendingDebit  decimal.Decimal `json:"pending_debit"`
	PendingCredit decimal.Decimal `json:"pending_credit"`
}

type WalletEvent struct {
	ID          string                 `json:"id"`
	Version     int                    `json:"version"`
	TransferID  string                 `json:"transfer_id"`
	ReferenceID string                 `json:"reference_id"`
	WalletID    string                 `json:"wallet_id"`
	Amount      decimal.Decimal        `json:"amount"`
	EventType   entity.WalletEventType `json:"event_type"`
	Status      entity.TransferStatus  `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
}
