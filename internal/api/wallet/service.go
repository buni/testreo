package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/buni/wallet/internal/api/app/contract"
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/app/request"
	"github.com/buni/wallet/internal/pkg/database"
	"github.com/buni/wallet/internal/pkg/sloglog"
	"github.com/shopspring/decimal"
)

var _ contract.WalletService = (*Service)(nil)

type Service struct {
	repo           contract.WalletRepository
	projectionRepo contract.WalletProjectionRepository
	eventRepo      contract.WalletEventRepository
	publisher      contract.WalletEventPublisher
	txm            database.TransactionManager
}

func NewService(
	repo contract.WalletRepository,
	projectionRepo contract.WalletProjectionRepository,
	eventRepo contract.WalletEventRepository,
	publisher contract.WalletEventPublisher,
	txm database.TransactionManager,
) *Service {
	return &Service{
		repo:           repo,
		projectionRepo: projectionRepo,
		eventRepo:      eventRepo,
		publisher:      publisher,
		txm:            txm,
	}
}

func (s *Service) Create(ctx context.Context, req *request.CreateWallet) (result entity.Wallet, err error) {
	result, err = entity.NewWallet(req.ReferenceID)
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to create wallet entity: %w", err)
	}

	err = s.txm.Run(ctx, func(ctx context.Context) error {
		result, err = s.repo.Create(ctx, result)
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		projection := entity.NewWalletProjection(result.ID, result.ID, decimal.NewFromInt(0), decimal.NewFromInt(0), decimal.NewFromInt(0))

		_, err = s.projectionRepo.Create(ctx, projection)
		if err != nil {
			return fmt.Errorf("failed to create wallet projection: %w", err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return result, nil
}

func (s *Service) Get(ctx context.Context, req *request.GetWallet) (result entity.WalletBalanceProjection, err error) {
	err = s.txm.Run(ctx, func(ctx context.Context) error {
		wallet, err := s.repo.Get(ctx, req.WalletID)
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		projection, err := s.projectionRepo.Get(ctx, req.WalletID) // we can get both in a single query, but it makes sense to treat the wallet and projection as being stored in a separate database
		if err != nil {
			return fmt.Errorf("failed to get wallet projection: %w", err)
		}

		result = entity.WalletBalanceProjection{
			Wallet:           wallet,
			WalletProjection: projection,
		}

		return nil
	})
	if err != nil {
		return
	}

	return result, nil
}

func (s *Service) DebitTransfer(ctx context.Context, req *request.DebitTransfer) (result entity.WalletEvent, err error) {
	if req.Amount.IsNegative() {
		return entity.WalletEvent{}, entity.ErrNegativeAmount
	}

	event, err := entity.NewWalletEvent(
		req.TransferID,
		req.ReferenceID,
		req.WalletID,
		req.Amount,
		entity.EventTypeDebitTransfer,
		req.Status,
	)
	if err != nil {
		return result, fmt.Errorf("failed to create wallet event: %w", err)
	}

	err = s.txm.Run(ctx, func(ctx context.Context) error {
		_, err := s.repo.Get(ctx, req.WalletID) // make sure the wallet exists
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		result, err = s.eventRepo.Create(ctx, event) // when doing a debit transfer we don't need to rebuild the state as we are only adding to the balance
		if err != nil {
			return fmt.Errorf("failed to create wallet event: %w", err)
		}

		err = s.publisher.PublishCreated(ctx, result)
		if err != nil {
			return fmt.Errorf("failed to publish wallet event: %w", err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return result, nil
}

func (s *Service) CreditTransfer(ctx context.Context, req *request.CreditTransfer) (result entity.WalletEvent, err error) {
	if req.Amount.IsNegative() {
		return entity.WalletEvent{}, entity.ErrNegativeAmount
	}

	event, err := entity.NewWalletEvent(
		req.TransferID,
		req.ReferenceID,
		req.WalletID,
		req.Amount,
		entity.EventTypeCreditTransfer,
		req.Status,
	)
	if err != nil {
		return result, fmt.Errorf("failed to create wallet event: %w", err)
	}

	err = s.txm.Run(ctx, func(ctx context.Context) error {
		_, err = s.repo.Get(ctx, req.WalletID) // make sure the wallet exists
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		events, err := s.eventRepo.ListByWalletID(ctx, req.WalletID) // get all events for the wallet and rebuild the state
		if err != nil {
			return fmt.Errorf("failed to list wallet events: %w", err)
		}

		projection := &entity.WalletProjection{}

		err = ProcessEvents(ctx, projection, events)
		if err != nil {
			return fmt.Errorf("failed to process wallet events: %w", err)
		}

		if projection.Balance.LessThan(req.Amount) {
			return entity.ErrInsufficientBalance
		}

		result, err = s.eventRepo.Create(ctx, event) // when doing a debit transfer we don't need to rebuild the state as we are only adding to the balance
		if err != nil {
			return fmt.Errorf("failed to create wallet event: %w", err)
		}

		err = s.publisher.PublishCreated(ctx, result)
		if err != nil {
			return fmt.Errorf("failed to publish wallet event: %w", err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return result, nil
}

func (s *Service) CompleteTransfer(ctx context.Context, req *request.CompleteTransfer) (result entity.WalletEvent, err error) { //nolint:dupl
	event, err := entity.NewWalletEvent(req.TransferID, req.ReferenceID, req.WalletID, decimal.NewFromInt(0), entity.EventTypeUpdateTransferStatus, entity.TransferStatusCompleted)
	if err != nil {
		return result, fmt.Errorf("failed to create wallet event: %w", err)
	}

	err = s.txm.Run(ctx, func(ctx context.Context) error {
		_, err = s.repo.Get(ctx, req.WalletID) // make sure the wallet exists
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		// since we don't really care if the transfer_id exists, we can just create the event and worst case it will just get ignored during state rebuild
		result, err = s.eventRepo.Create(ctx, event)
		if err != nil {
			return fmt.Errorf("failed to create wallet event: %w", err)
		}

		err = s.publisher.PublishCreated(ctx, result)
		if err != nil {
			return fmt.Errorf("failed to publish wallet event: %w", err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return result, nil
}

func (s *Service) RevertTransfer(ctx context.Context, req *request.RevertTransfer) (result entity.WalletEvent, err error) { //nolint:dupl
	event, err := entity.NewWalletEvent(req.TransferID, req.ReferenceID, req.WalletID, decimal.NewFromInt(0), entity.EventTypeUpdateTransferStatus, entity.TransferStatusFailed)
	if err != nil {
		return result, fmt.Errorf("failed to create wallet event: %w", err)
	}

	err = s.txm.Run(ctx, func(ctx context.Context) error {
		_, err = s.repo.Get(ctx, req.WalletID) // make sure the wallet exists
		if err != nil {
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		// since we don't really care if the transfer_id exists, we can just create the event and worst case it will just get ignored during state rebuild
		result, err = s.eventRepo.Create(ctx, event)
		if err != nil {
			return fmt.Errorf("failed to create wallet event: %w", err)
		}

		err = s.publisher.PublishCreated(ctx, result)
		if err != nil {
			return fmt.Errorf("failed to publish wallet event: %w", err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return result, nil
}

func (s *Service) RebuildWalletProjection(ctx context.Context, event *entity.WalletEvent) (result entity.WalletProjection, err error) {
	logger := sloglog.FromContext(ctx)
	err = s.txm.Run(ctx, func(ctx context.Context) error {
		projection, err := s.projectionRepo.Get(ctx, event.WalletID)
		if err != nil {
			return fmt.Errorf("failed to get wallet projection: %w", err)
		}

		if projection.LastEventID >= event.ID { // UUIDv7's are k-sortable and lexicographic, so we can just compare the last event id using comparison operators, since strings in go are compared lexicographically
			logger.InfoContext(ctx, "no new events to process skipping rebuild") // if we need to do a full rebuild for some reason and there are no new events a different mechanism function should be used
			return nil
		}

		events, err := s.eventRepo.ListByWalletID(ctx, event.WalletID)
		if err != nil {
			return fmt.Errorf("failed to list wallet events: %w", err)
		}

		err = ProcessEvents(ctx, &result, events)
		if err != nil {
			return fmt.Errorf("failed to process wallet events: %w", err)
		}

		result.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

		_, err = s.projectionRepo.Update(ctx, result)
		if err != nil {
			return fmt.Errorf("failed to update wallet projection: %w", err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return result, nil
}
