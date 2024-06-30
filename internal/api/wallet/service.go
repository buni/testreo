package wallet

import (
	"context"
	"fmt"

	"github.com/buni/wallet/internal/api/app/contract"
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/app/request"
	"github.com/buni/wallet/internal/pkg/database"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo           contract.WalletRepository
	projectionRepo contract.WalletProjectionRepository
	eventRepo      contract.WalletEventRepository
	txm            database.TransactionManager
}

func NewService(
	repo contract.WalletRepository,
	projectionRepo contract.WalletProjectionRepository,
	eventRepo contract.WalletEventRepository,
	txm database.TransactionManager,
) *Service {
	return &Service{
		repo:           repo,
		projectionRepo: projectionRepo,
		eventRepo:      eventRepo,
		txm:            txm,
	}
}

func (s *Service) CreateWallet(ctx context.Context, req request.CreateWallet) (wallet entity.Wallet, err error) {
	wallet, err = entity.NewWallet(req.ReferenceID)
	if err != nil {
		return entity.Wallet{}, fmt.Errorf("failed to create wallet entity: %w", err)
	}

	err = s.txm.Run(ctx, func(ctx context.Context) error {
		wallet, err = s.repo.Create(ctx, wallet)
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		projection := entity.NewWalletProjection(wallet.ID, "", decimal.NewFromInt(0), decimal.NewFromInt(0), decimal.NewFromInt(0), 0)

		_, err = s.projectionRepo.Create(ctx, projection)
		if err != nil {
			return fmt.Errorf("failed to create wallet projection: %w", err)
		}

		return nil
	})
	if err != nil {
		return
	}

	return wallet, nil
}
