package wallet_test

import (
	"log"
	"testing"
	"time"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProcessEvents(t *testing.T) {
	tt := time.Now()

	tests := []struct {
		name        string
		events      []entity.WalletEvent
		expected    entity.WalletProjection
		expectedErr error
	}{
		{
			name:     "empty events slice",
			events:   []entity.WalletEvent{},
			expected: entity.WalletProjection{},
		},
		{
			name: "unsupported event version",
			events: []entity.WalletEvent{
				{
					EventType: entity.EventTypeDebitTransfer,
					Version:   entity.WalletEventVersionOne + 1,
				},
			},
			expected:    entity.WalletProjection{},
			expectedErr: entity.ErrUnsupportedEventVersion,
		},
		{
			name: "out of order update transfer status",
			events: []entity.WalletEvent{
				{
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "transfer1",
				},
			},
			expected:    entity.WalletProjection{},
			expectedErr: nil,
		},
		{
			name: "invalid event type",
			events: []entity.WalletEvent{
				{
					EventType: entity.EventTypeInvalid,
				},
			},
			expected:    entity.WalletProjection{},
			expectedErr: entity.ErrInvalidEventType,
		},
		{
			name: "valid debit transfer event",
			events: []entity.WalletEvent{
				{
					ID:             "debit1",
					EventType:      entity.EventTypeDebitTransfer,
					WalletID:       "wallet1",
					TransferID:     "debit1",
					Amount:         decimal.NewFromInt(100),
					TransferStatus: entity.TransferStatusCompleted,
					Sequence:       1,
				},
				{
					ID:             "debit2",
					WalletID:       "wallet1",
					EventType:      entity.EventTypeDebitTransfer,
					TransferID:     "debit2",
					Amount:         decimal.NewFromInt(50),
					TransferStatus: entity.TransferStatusCompleted,
					Sequence:       2,
				},
			},
			expected: entity.WalletProjection{
				WalletID:     "wallet1",
				Balance:      decimal.NewFromInt(150),
				LastEventID:  "debit2",
				LastSequence: 2,
			},
			expectedErr: nil,
		},
		{
			name: "valid credit transfer event",
			events: []entity.WalletEvent{
				{
					EventType:      entity.EventTypeDebitTransfer,
					WalletID:       "wallet1",
					Amount:         decimal.NewFromInt(100),
					TransferStatus: entity.TransferStatusCompleted,
					Sequence:       1,
				},
				{
					WalletID:       "wallet1",
					EventType:      entity.EventTypeCreditTransfer,
					Amount:         decimal.NewFromInt(50),
					TransferStatus: entity.TransferStatusCompleted,
					Sequence:       2,
				},
			},
			expected: entity.WalletProjection{
				WalletID:     "wallet1",
				Balance:      decimal.NewFromInt(50),
				LastEventID:  "",
				LastSequence: 2,
			},
			expectedErr: nil,
		},
		{
			name: "valid pending debit transfer event",
			events: []entity.WalletEvent{
				{
					EventType:      entity.EventTypeDebitTransfer,
					WalletID:       "wallet1",
					Amount:         decimal.NewFromInt(100),
					TransferStatus: entity.TransferStatusCompleted,
					Sequence:       1,
				},
				{
					WalletID:       "wallet1",
					EventType:      entity.EventTypeDebitTransfer,
					Amount:         decimal.NewFromInt(50),
					TransferStatus: entity.TransferStatusPending,
					Sequence:       2,
				},
			},
			expected: entity.WalletProjection{
				WalletID:     "wallet1",
				Balance:      decimal.NewFromInt(100),
				PendingDebit: decimal.NewFromInt(50),
				LastSequence: 2,
			},
			expectedErr: nil,
		},
		{
			name: "debit and credit transfer with failed transaction",
			events: []entity.WalletEvent{
				{
					EventType:      entity.EventTypeDebitTransfer,
					WalletID:       "wallet1",
					TransferStatus: entity.TransferStatusCompleted,
					Amount:         decimal.NewFromInt(100),
					Sequence:       1,
				},
				{
					WalletID:       "wallet1",
					Sequence:       2,
					TransferID:     "debit1",
					Amount:         decimal.NewFromInt(50),
					EventType:      entity.EventTypeDebitTransfer,
					TransferStatus: entity.TransferStatusPending,
				},
				{
					WalletID:       "wallet1",
					Sequence:       3,
					TransferID:     "credit1",
					Amount:         decimal.NewFromInt(20),
					EventType:      entity.EventTypeCreditTransfer,
					TransferStatus: entity.TransferStatusPending,
				},
				{
					WalletID:       "wallet1",
					EventType:      entity.EventTypeUpdateTransferStatus,
					TransferID:     "credit1",
					Sequence:       4,
					TransferStatus: entity.TransferStatusFailed,
				},
				{
					WalletID:       "wallet1",
					Sequence:       5,
					TransferID:     "debit1",
					EventType:      entity.EventTypeUpdateTransferStatus,
					TransferStatus: entity.TransferStatusCompleted,
				},
			},
			expected: entity.WalletProjection{
				WalletID:      "wallet1",
				Balance:       decimal.NewFromInt(150),
				PendingDebit:  decimal.NewFromInt(1).Sub(decimal.NewFromInt(1)),
				PendingCredit: decimal.NewFromInt(1).Sub(decimal.NewFromInt(1)),
				LastSequence:  5,
			},
			expectedErr: nil,
		},
		{
			name: "debit, credit, and pending transactions",
			events: []entity.WalletEvent{
				{
					EventType:      entity.EventTypeDebitTransfer,
					WalletID:       "wallet1",
					Amount:         decimal.NewFromInt(100),
					TransferStatus: entity.TransferStatusCompleted,
					Sequence:       1,
				},
				{
					EventType:      entity.EventTypeDebitTransfer,
					Amount:         decimal.NewFromInt(50),
					TransferStatus: entity.TransferStatusPending,
					Sequence:       2,
				},
				{
					EventType:      entity.EventTypeCreditTransfer,
					Amount:         decimal.NewFromInt(30),
					TransferStatus: entity.TransferStatusPending,
					Sequence:       3,
				},
				{
					EventType:      entity.EventTypeDebitTransfer,
					Amount:         decimal.NewFromInt(20),
					TransferStatus: entity.TransferStatusCompleted,
					Sequence:       4,
				},
				{
					WalletID:       "wallet1",
					EventType:      entity.EventTypeCreditTransfer,
					Amount:         decimal.NewFromInt(10),
					TransferStatus: entity.TransferStatusCompleted,

					Sequence:  5,
					CreatedAt: tt,
				},
			},
			expected: entity.WalletProjection{
				WalletID:      "wallet1",
				Balance:       decimal.NewFromInt(80),
				PendingDebit:  decimal.NewFromInt(50),
				PendingCredit: decimal.NewFromInt(30),
				LastSequence:  5,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projection := &entity.WalletProjection{}
			err := wallet.ProcessEvents2(projection, tt.events)
			log.Println(projection, err)
			if tt.expectedErr != nil {
				assert.Empty(t, projection)
				assert.ErrorIs(t, tt.expectedErr, err)
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, tt.expected, *projection)
		})
	}
}
