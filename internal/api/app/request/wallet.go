package request

type DebitWallet struct{}

type CreditWallet struct{}

type CompleteTransfer struct{}

type RevertTransfer struct{}

type GetWallet struct{}

type GetWalletTransactions struct{}

type CreateWallet struct {
	ReferenceID string `json:"reference_id"`
}
