package wallet_test

import (
	"context"
	"testing"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProcessEvents(t *testing.T) {
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
					ID:         "debit0",
					WalletID:   "wallet1",
					EventType:  entity.EventTypeDebitTransfer,
					TransferID: "debit",
					Amount:     decimal.NewFromInt(50),
					Status:     entity.TransferStatusFailed,
				},
				{
					ID:         "debit0",
					WalletID:   "wallet1",
					EventType:  entity.EventTypeCreditTransfer,
					TransferID: "credit123",
					Amount:     decimal.NewFromInt(50),
					Status:     entity.TransferStatusFailed,
				},
				{
					ID:         "debit1",
					EventType:  entity.EventTypeDebitTransfer,
					WalletID:   "wallet1",
					TransferID: "debit1",
					Amount:     decimal.NewFromInt(100),
					Status:     entity.TransferStatusCompleted,
				},
				{
					ID:         "debit2",
					WalletID:   "wallet1",
					EventType:  entity.EventTypeDebitTransfer,
					TransferID: "debit2",
					Amount:     decimal.NewFromInt(50),
					Status:     entity.TransferStatusCompleted,
				},
			},
			expected: entity.WalletProjection{
				WalletID:    "wallet1",
				Balance:     decimal.NewFromInt(150),
				LastEventID: "debit2",
			},
			expectedErr: nil,
		},
		{
			name: "valid credit transfer event",
			events: []entity.WalletEvent{
				{
					EventType: entity.EventTypeDebitTransfer,
					WalletID:  "wallet1",
					Amount:    decimal.NewFromInt(100),
					Status:    entity.TransferStatusCompleted,
				},
				{
					WalletID:  "wallet1",
					EventType: entity.EventTypeCreditTransfer,
					Amount:    decimal.NewFromInt(50),
					Status:    entity.TransferStatusCompleted,
				},
			},
			expected: entity.WalletProjection{
				WalletID:    "wallet1",
				Balance:     decimal.NewFromInt(50),
				LastEventID: "",
			},
			expectedErr: nil,
		},
		{
			name: "valid pending debit transfer event",
			events: []entity.WalletEvent{
				{
					EventType: entity.EventTypeDebitTransfer,
					WalletID:  "wallet1",
					Amount:    decimal.NewFromInt(100),
					Status:    entity.TransferStatusCompleted,
				},
				{
					WalletID:  "wallet1",
					EventType: entity.EventTypeDebitTransfer,
					Amount:    decimal.NewFromInt(50),
					Status:    entity.TransferStatusPending,
				},
			},
			expected: entity.WalletProjection{
				WalletID:     "wallet1",
				Balance:      decimal.NewFromInt(100),
				PendingDebit: decimal.NewFromInt(50),
			},
			expectedErr: nil,
		},
		{
			name: "debit and credit transfer with failed transaction",
			events: []entity.WalletEvent{
				{
					EventType: entity.EventTypeDebitTransfer,
					WalletID:  "wallet1",
					Status:    entity.TransferStatusCompleted,
					Amount:    decimal.NewFromInt(100),
				},
				{
					WalletID:   "wallet1",
					TransferID: "debit1",
					Amount:     decimal.NewFromInt(50),
					EventType:  entity.EventTypeDebitTransfer,
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					TransferID: "debit123",
					Amount:     decimal.NewFromInt(50),
					EventType:  entity.EventTypeDebitTransfer,
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					TransferID: "credit1",
					Amount:     decimal.NewFromInt(20),
					EventType:  entity.EventTypeCreditTransfer,
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "credit1",
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "credit1",
					Status:     entity.TransferStatusFailed,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "debit123",
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "debit123",
					Status:     entity.TransferStatusFailed,
				},

				{
					WalletID:   "wallet1",
					TransferID: "debit1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					Status:     entity.TransferStatusCompleted,
				},
			},
			expected: entity.WalletProjection{
				WalletID:      "wallet1",
				Balance:       decimal.NewFromInt(150),
				PendingDebit:  decimal.NewFromInt(1).Sub(decimal.NewFromInt(1)),
				PendingCredit: decimal.NewFromInt(1).Sub(decimal.NewFromInt(1)),
			},
			expectedErr: nil,
		},
		{
			name: "duplicate transfer ids for credit and debit",
			events: []entity.WalletEvent{
				{
					TransferID: "123",
					EventType:  entity.EventTypeCreditTransfer,
					Amount:     decimal.NewFromInt(30),
					Status:     entity.TransferStatusPending,
				},
				{
					TransferID: "123",
					EventType:  entity.EventTypeDebitTransfer,
					Amount:     decimal.NewFromInt(50),
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					TransferID: "123",
					EventType:  entity.EventTypeUpdateTransferStatus,
					Status:     entity.TransferStatusCompleted,
				},
			},
			expected: entity.WalletProjection{
				WalletID:      "wallet1",
				Balance:       decimal.NewFromInt(-30),
				PendingCredit: decimal.NewFromInt(1).Sub(decimal.NewFromInt(1)),
			},
			expectedErr: nil,
		},
		{
			name: "debit, credit, and pending transactions",
			events: []entity.WalletEvent{
				{
					EventType: entity.EventTypeDebitTransfer,
					WalletID:  "wallet1",
					Amount:    decimal.NewFromInt(100),
					Status:    entity.TransferStatusCompleted,
				},
				{
					EventType: entity.EventTypeDebitTransfer,
					Amount:    decimal.NewFromInt(50),
					Status:    entity.TransferStatusPending,
				},
				{
					EventType:  entity.EventTypeCreditTransfer,
					TransferID: "credit",
					Amount:     decimal.NewFromInt(30),
					Status:     entity.TransferStatusPending,
				},
				{
					EventType:  entity.EventTypeCreditTransfer,
					TransferID: "credit1",
					Amount:     decimal.NewFromInt(0),
					Status:     entity.TransferStatusPending,
				},
				{
					EventType: entity.EventTypeDebitTransfer,
					Amount:    decimal.NewFromInt(20),
					Status:    entity.TransferStatusCompleted,
				},
				{
					WalletID:  "wallet1",
					EventType: entity.EventTypeCreditTransfer,
					Amount:    decimal.NewFromInt(10),
					Status:    entity.TransferStatusCompleted,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "credit1",
					Status:     entity.TransferStatusCompleted,
				},
			},
			expected: entity.WalletProjection{
				WalletID:      "wallet1",
				Balance:       decimal.NewFromInt(80),
				PendingDebit:  decimal.NewFromInt(50),
				PendingCredit: decimal.NewFromInt(30),
			},
			expectedErr: nil,
		},
		{
			name: "debit, credit, and pending transactions",
			events: []entity.WalletEvent{
				{
					EventType:  entity.EventTypeDebitTransfer,
					WalletID:   "wallet1",
					TransferID: "12345576223435543334434",
					Amount:     decimal.NewFromFloat(155),
					Status:     entity.TransferStatusPending,
				},
				{
					EventType:  entity.EventTypeDebitTransfer,
					TransferID: "123455762234355433344343",
					Amount:     decimal.NewFromFloat(155),
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "123455762234355433344343",
					Status:     entity.TransferStatusCompleted,
				},
				{
					EventType:  entity.EventTypeCreditTransfer,
					TransferID: "12345576223435543334434",
					Amount:     decimal.NewFromFloat(155),
					Status:     entity.TransferStatusPending,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "12345576223435543334434",
					Status:     entity.TransferStatusCompleted,
				},
				{
					WalletID:   "wallet1",
					EventType:  entity.EventTypeUpdateTransferStatus,
					TransferID: "1234557622343554333443434",
					Status:     entity.TransferStatusCompleted,
				},
			},
			expected: entity.WalletProjection{
				WalletID:     "wallet1",
				Balance:      decimal.NewFromInt(310),
				PendingDebit: decimal.NewFromInt(1).Sub(decimal.NewFromInt(1)),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projection := &entity.WalletProjection{}
			err := wallet.ProcessEvents(context.Background(), projection, tt.events)

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
