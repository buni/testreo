package wallet

import (
	"context"
	"fmt"

	"github.com/buni/wallet/internal/api/app/contract"
	"github.com/buni/wallet/internal/api/app/request"
	"github.com/buni/wallet/internal/api/app/response"
	"github.com/buni/wallet/internal/pkg/handler"
	"github.com/buni/wallet/internal/pkg/render"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc contract.WalletService
}

func NewHandler(svc contract.WalletService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Create(ctx context.Context, req *request.CreateWallet) (*response.Wallet, error) {
	wallet, err := h.svc.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	walletResp, err := render.NewResponse[response.Wallet](wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to render wallet response: %w", err)
	}

	return walletResp, nil
}

func (h *Handler) Get(ctx context.Context, req *request.GetWallet) (*response.Wallet, error) {
	wallet, err := h.svc.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	walletResp, err := render.NewResponse[response.Wallet](wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to render wallet response: %w", err)
	}

	return walletResp, nil
}

func (h *Handler) DebitTransfer(ctx context.Context, req *request.DebitTransfer) (*response.WalletEvent, error) {
	event, err := h.svc.DebitTransfer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to debit transfer: %w", err)
	}

	eventResp, err := render.NewResponse[response.WalletEvent](event)
	if err != nil {
		return nil, fmt.Errorf("failed to render wallet event response: %w", err)
	}

	return eventResp, nil
}

func (h *Handler) CreditTransfer(ctx context.Context, req *request.CreditTransfer) (*response.WalletEvent, error) {
	event, err := h.svc.CreditTransfer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to credit transfer: %w", err)
	}

	eventResp, err := render.NewResponse[response.WalletEvent](event)
	if err != nil {
		return nil, fmt.Errorf("failed to render wallet event response: %w", err)
	}

	return eventResp, nil
}

func (h *Handler) CompleteTransfer(ctx context.Context, req *request.CompleteTransfer) (*response.WalletEvent, error) {
	event, err := h.svc.CompleteTransfer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete transfer: %w", err)
	}

	eventResp, err := render.NewResponse[response.WalletEvent](event)
	if err != nil {
		return nil, fmt.Errorf("failed to render wallet event response: %w", err)
	}

	return eventResp, nil
}

func (h *Handler) RevertTransfer(ctx context.Context, req *request.RevertTransfer) (*response.WalletEvent, error) {
	event, err := h.svc.RevertTransfer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to revert transfer: %w", err)
	}

	eventResp, err := render.NewResponse[response.WalletEvent](event)
	if err != nil {
		return nil, fmt.Errorf("failed to render wallet event response: %w", err)
	}

	return eventResp, nil
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/wallets", func(r chi.Router) {
		r.Route("/{walletID}", func(r chi.Router) {
			r.Get("/", handler.WrapDefaultBasic(h.Get))
			r.Post("/", handler.WrapDefaultBasic(h.Create))
			r.Route("/transfers", func(r chi.Router) {
				r.Post("/debit", handler.WrapDefaultBasic(h.DebitTransfer))
				r.Post("/credit", handler.WrapDefaultBasic(h.CreditTransfer))
				r.Route("/{transferID}", func(r chi.Router) {
					r.Post("/complete", handler.WrapDefaultBasic(h.CompleteTransfer))
					r.Post("/revert", handler.WrapDefaultBasic(h.RevertTransfer))
				})
			})
		})
	})
}
