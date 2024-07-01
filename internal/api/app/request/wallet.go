package request

import (
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/shopspring/decimal"
)

type DebitTransfer struct {
	WalletID    string                `json:"-" in:"path=walletID"`
	ReferenceID string                `json:"reference_id"`
	TransferID  string                `json:"transfer_id" validate:"required"`
	Amount      decimal.Decimal       `json:"amount" validate:"required"`
	Status      entity.TransferStatus `json:"status" validate:"required"`
}

type CreditTransfer struct {
	WalletID    string                `json:"-" in:"path=walletID"`
	ReferenceID string                `json:"reference_id"`
	TransferID  string                `json:"transfer_id" validate:"required"`
	Amount      decimal.Decimal       `json:"amount" validate:"required"`
	Status      entity.TransferStatus `json:"status" validate:"required"`
}

type CompleteTransfer struct {
	WalletID    string `json:"-" in:"path=walletID"`
	TransferID  string `json:"-" in:"path=transferID"`
	ReferenceID string `json:"reference_id"`
}

type RevertTransfer struct {
	WalletID    string `json:"-" in:"path=walletID"`
	TransferID  string `json:"-" in:"path=transferID"`
	ReferenceID string `json:"reference_id"`
}

type GetWallet struct {
	WalletID string `json:"-" in:"path=walletID"`
}

type CreateWallet struct {
	ReferenceID string `json:"reference_id"`
}
