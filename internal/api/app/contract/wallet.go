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
	GetByWalletID(ctx context.Context, walletID string) ([]entity.WalletEvent, error)
}

type WalletProjectionRepository interface {
	Get(ctx context.Context, walletID string) (entity.WalletProjection, error)
	Create(ctx context.Context, projection entity.WalletProjection) (entity.WalletProjection, error)
	Update(ctx context.Context, projection entity.WalletProjection) (entity.WalletProjection, error)
}

type WalletService interface {
	Create(ctx context.Context, req request.CreateWallet) (entity.Wallet, error)
	Get(ctx context.Context, req request.GetWallet) (entity.Wallet, error)
	GetTransactions(ctx context.Context, req request.GetWalletTransactions) ([]entity.WalletEvent, error)
	Debit(ctx context.Context, req request.DebitWallet) (entity.WalletEvent, error)
	Credit(ctx context.Context, req request.CreditWallet) (entity.WalletEvent, error)
	CompleteTransfer(ctx context.Context, req request.CompleteTransfer) (entity.WalletEvent, error)
	RevertTransfer(ctx context.Context, req request.RevertTransfer) (entity.WalletEvent, error)
}
