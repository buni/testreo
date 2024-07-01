package contract

import (
	"context"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/app/request"
)

type WalletRepository interface {
	Get(ctx context.Context, id string) (entity.Wallet, error)
	Create(ctx context.Context, wallet entity.Wallet) (entity.Wallet, error)
}

type WalletEventRepository interface {
	Create(ctx context.Context, event entity.WalletEvent) (entity.WalletEvent, error)
	ListByWalletID(ctx context.Context, walletID string) ([]entity.WalletEvent, error)
}

type WalletProjectionRepository interface {
	Get(ctx context.Context, walletID string) (entity.WalletProjection, error)
	Create(ctx context.Context, projection entity.WalletProjection) (entity.WalletProjection, error)
	Update(ctx context.Context, projection entity.WalletProjection) (entity.WalletProjection, error)
}

type WalletService interface {
	Create(ctx context.Context, req *request.CreateWallet) (entity.Wallet, error)
	Get(ctx context.Context, req *request.GetWallet) (entity.WalletBalanceProjection, error)
	DebitTransfer(ctx context.Context, req *request.DebitTransfer) (entity.WalletEvent, error)
	CreditTransfer(ctx context.Context, req *request.CreditTransfer) (entity.WalletEvent, error)
	CompleteTransfer(ctx context.Context, req *request.CompleteTransfer) (entity.WalletEvent, error)
	RevertTransfer(ctx context.Context, req *request.RevertTransfer) (entity.WalletEvent, error)
}

type WalletEventPublisher interface {
	PublishCreated(ctx context.Context, event entity.WalletEvent) error
}
