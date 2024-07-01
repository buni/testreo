package entity

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
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
	Version        int             `db:"version"`
	TransferID     string          `db:"transfer_id"`
	ReferenceID    string          `db:"reference_id"`
	WalletID       string          `db:"wallet_id"`
	Amount         decimal.Decimal `db:"amount"`
	EventType      WalletEventType `db:"mutation_type"`
	TransferStatus TransferStatus  `db:"transfer_status"`
	CreatedAt      time.Time       `db:"created_at"`
}

func NewWalletEvent(
	transferID, referenceID, walletID string,
	amount decimal.Decimal,
	eventType WalletEventType,
	transferStatus TransferStatus,
) (WalletEvent, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return WalletEvent{}, fmt.Errorf("failed to generate wallet event id: %w", err)
	}

	return WalletEvent{
		ID:             id.String(),
		Version:        WalletEventVersionOne,
		TransferID:     transferID,
		ReferenceID:    referenceID,
		WalletID:       walletID,
		Amount:         amount,
		EventType:      eventType,
		TransferStatus: transferStatus,
		CreatedAt:      time.Now().UTC().Truncate(time.Microsecond),
	}, nil
}

type Wallet struct {
	ID          string    `db:"id"`
	ReferenceID string    `db:"reference_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewWallet(referenceID string) (Wallet, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Wallet{}, fmt.Errorf("failed to generate wallet id: %w", err)
	}

	tt := time.Now().UTC().Truncate(time.Microsecond)

	return Wallet{
		ID:          id.String(),
		ReferenceID: referenceID,
		CreatedAt:   tt,
		UpdatedAt:   tt,
	}, nil
}

type WalletProjection struct {
	WalletID      string          `db:"wallet_id"`
	Balance       decimal.Decimal `db:"balance"`
	PendingDebit  decimal.Decimal `db:"pending_debit"`
	PendingCredit decimal.Decimal `db:"pending_credit"`
	LastEventID   string          `db:"last_event_id"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
}

type WalletBalanceProjection struct {
	Wallet           Wallet
	WalletProjection WalletProjection
}

func NewWalletProjection(walletID, lastEventID string, balance, pendingDebit, pendingCredit decimal.Decimal) WalletProjection {
	tt := time.Now().UTC().Truncate(time.Microsecond)

	return WalletProjection{
		WalletID:      walletID,
		Balance:       balance,
		PendingDebit:  pendingDebit,
		PendingCredit: pendingCredit,
		LastEventID:   lastEventID,
		CreatedAt:     tt,
		UpdatedAt:     tt,
	}
}
