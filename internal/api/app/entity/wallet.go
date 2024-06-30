package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	WalletEventVersionInvalid = iota
	WalletEventVersionOne
)

const (
	EventTypeInvalid WalletEventType = iota
	EventTypeDebitTransfer
	EventTypeCreditTransfer
	EventTypeUpdateTransferStatus
)

const (
	TransferStatusInvalid TransferStatus = iota
	TransferStatusPending
	TransferStatusCompleted
	TransferStatusFailed
)

//go:generate enumer -type=WalletEventType,TransferStatus -transform=snake -output=wallet_enum.go -json -sql -text
type WalletEventType uint

type TransferStatus uint

type WalletEvent struct {
	ID             string          `db:"id"`
	Sequence       int64           `db:"sequence"`
	Version        int             `db:"version"`
	TransferID     string          `db:"transfer_id"`
	ReferenceID    string          `db:"reference_id"`
	WalletID       string          `db:"wallet_id"`
	Amount         decimal.Decimal `db:"amount"`
	EventType      WalletEventType `db:"mutation_type"`
	TransferStatus TransferStatus  `db:"transfer_status"`
	CreatedAt      time.Time       `db:"created_at"`
}

type Wallet struct {
	ID          string    `db:"id"`
	ReferenceID string    `db:"reference_id"`
	Name        string    `db:"name"`
	Currency    string    `db:"currency"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type WalletProjection struct {
	WalletID      string          `db:"wallet_id"`
	Balance       decimal.Decimal `db:"balance" updatable:"true"`
	PendingDebit  decimal.Decimal `db:"pending_debit" updatable:"true"`
	PendingCredit decimal.Decimal `db:"pending_credit" updatable:"true"`
	LastEventID   string          `db:"last_event_id" updatable:"true"`
	LastSequence  int64           `db:"last_sequence" updatable:"true"`
}
