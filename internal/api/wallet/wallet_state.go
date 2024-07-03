package wallet

import (
	"context"
	"log/slog"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/pkg/sloglog"
)

func ProcessEvents(ctx context.Context, projection *entity.WalletProjection, events []entity.WalletEvent) error {
	logger := sloglog.FromContext(ctx)
	if len(events) == 0 {
		return nil
	}

	eventMapping := map[string]int{}

	for k, event := range events {
		if event.Version > entity.WalletEventVersionOne { // assumes we would do backwards compatibility, but for event versions we can't support in this build we return an error early
			return entity.ErrUnsupportedEventVersion
		}

		switch event.EventType {
		case entity.EventTypeDebitTransfer:
			if event.Status == entity.TransferStatusPending {
				if _, ok := eventMapping[event.TransferID]; ok {
					logger.WarnContext(ctx, "transfer id already exists", slog.String("transfer_id", event.TransferID), slog.Any("event", event))
					continue
				}
				eventMapping[event.TransferID] = k
				projection.PendingDebit = projection.PendingDebit.Add(event.Amount)
				continue
			}

			if event.Status == entity.TransferStatusFailed {
				continue // skip since we don't need to process failed transfers
			}

			if event.Status == entity.TransferStatusCompleted {
				projection.Balance = projection.Balance.Add(event.Amount)
			}

		case entity.EventTypeCreditTransfer:
			if event.Status == entity.TransferStatusPending {
				if _, ok := eventMapping[event.TransferID]; ok {
					logger.WarnContext(ctx, "transfer id already exists", slog.String("transfer_id", event.TransferID), slog.Any("event", event))
					continue
				}
				eventMapping[event.TransferID] = k
				projection.PendingCredit = projection.PendingCredit.Add(event.Amount)
			}

			if event.Status == entity.TransferStatusFailed {
				continue // skip since we don't need to process failed transfers
			}

			projection.Balance = projection.Balance.Sub(event.Amount) // as long as the transfer hasn't failed we remove the amount from the balance

		case entity.EventTypeUpdateTransferStatus:
			idx, ok := eventMapping[event.TransferID]
			if !ok {
				logger.WarnContext(ctx, "transfer id not found", slog.String("transfer_id", event.TransferID), slog.Any("event", event))
				continue // can't find the transfer id, skip since we mostly likely already processed this event, there is a chance that the event is out of order too and if that is a possibility we can add a safe guard
			}

			desiredEvent := events[idx]

			if desiredEvent.EventType == entity.EventTypeDebitTransfer {
				if event.Status == entity.TransferStatusPending {
					continue
				}

				if event.Status == entity.TransferStatusFailed {
					projection.PendingDebit = projection.PendingDebit.Sub(desiredEvent.Amount)
				}

				if event.Status == entity.TransferStatusCompleted {
					projection.PendingDebit = projection.PendingDebit.Sub(desiredEvent.Amount)
					projection.Balance = projection.Balance.Add(desiredEvent.Amount)
				}

				delete(eventMapping, event.TransferID) // remove the transfer id from the map since we have processed it and we want to avoid overriding the status of the transfer
			}

			if desiredEvent.EventType == entity.EventTypeCreditTransfer {
				if event.Status == entity.TransferStatusPending {
					continue
				}

				if event.Status == entity.TransferStatusFailed {
					projection.PendingCredit = projection.PendingCredit.Sub(desiredEvent.Amount)
					projection.Balance = projection.Balance.Add(desiredEvent.Amount) // revert the balance, since we already took it out in the above case
				}

				if event.Status == entity.TransferStatusCompleted {
					projection.PendingCredit = projection.PendingCredit.Sub(desiredEvent.Amount)
				}

				delete(eventMapping, event.TransferID)
			}

		case entity.EventTypeInvalid:
			return entity.ErrInvalidEventType
		default:
			return entity.ErrUnsupportedEventType
		}
	}

	lastIdx := len(events) - 1
	projection.WalletID = events[lastIdx].WalletID
	projection.LastEventID = events[lastIdx].ID

	return nil
}
