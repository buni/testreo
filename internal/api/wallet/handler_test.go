package wallet_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	contract_mock "github.com/buni/wallet/internal/api/app/contract/mock"
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/app/request"
	"github.com/buni/wallet/internal/api/app/response"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/handler"
	"github.com/buni/wallet/internal/pkg/render"
	"github.com/buni/wallet/internal/pkg/testing/testutils"
	"github.com/go-chi/chi/v5"
	"github.com/kinbiko/jsonassert"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type WalletHandlerTestSuite struct {
	suite.Suite
	svcMock *contract_mock.MockWalletService
	handler *wallet.Handler
	ctx     context.Context
	ctrl    *gomock.Controller
}

func (s *WalletHandlerTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())
	s.svcMock = contract_mock.NewMockWalletService(s.ctrl)
	s.handler = wallet.NewHandler(s.svcMock)
}

func (s *WalletHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *WalletHandlerTestSuite) statusCompare(gotCode, expectedCode int, gotBody string, expectedBody any) {
	s.Equal(expectedCode, gotCode)
	if expectedBody != nil && expectedBody != "" {
		jsonassert.New(s.T()).Assertf(gotBody, testutils.ToJSON(s.T(), expectedBody))
	}
}

func (s *WalletHandlerTestSuite) buildContext(walletID, transferID string) context.Context {
	chiContext := chi.NewRouteContext()
	chiContext.URLParams.Add("walletID", walletID)
	chiContext.URLParams.Add("transferID", transferID)

	return context.WithValue(s.ctx, chi.RouteCtxKey, chiContext)
}

func (s *WalletHandlerTestSuite) TestCreateSuccess() {
	req := &request.CreateWallet{
		ReferenceID: "ref1",
	}
	expectedBody := response.Wallet{
		ID:          "id1",
		ReferenceID: "ref1",
	}

	s.ctx = s.buildContext("", "")

	s.svcMock.EXPECT().Create(s.ctx, req).Return(entity.Wallet{
		ID:          expectedBody.ID,
		ReferenceID: expectedBody.ReferenceID,
	}, nil)

	recorder := httptest.NewRecorder()

	handler.WrapDefault(s.handler.Create).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusCreated, recorder.Body.String(), expectedBody)
}

func (s *WalletHandlerTestSuite) TestCreateFailure() {
	req := &request.CreateWallet{
		ReferenceID: "ref1",
	}

	s.ctx = s.buildContext("", "")

	s.svcMock.EXPECT().Create(s.ctx, req).Return(entity.Wallet{}, context.DeadlineExceeded)

	recorder := httptest.NewRecorder()

	handler.WrapDefault(s.handler.Create).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusInternalServerError, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.InternalServerError,
			Message: "internal server error",
		},
	})
}

func (s *WalletHandlerTestSuite) TestGetSuccess() {
	req := &request.GetWallet{
		WalletID: "id1",
	}
	expectedBody := response.Wallet{
		ID:          "id1",
		ReferenceID: "ref1",
		Balance:     decimal.NewFromInt(100),
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().Get(s.ctx, req).Return(entity.WalletBalanceProjection{
		Wallet: entity.Wallet{
			ID:          expectedBody.ID,
			ReferenceID: expectedBody.ReferenceID,
		},
		WalletProjection: entity.WalletProjection{
			Balance: decimal.NewFromInt(100),
		},
	}, nil)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.Get).ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusOK, recorder.Body.String(), expectedBody)
}

func (s *WalletHandlerTestSuite) TestGetFailure() {
	req := &request.GetWallet{
		WalletID: "id1",
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().Get(s.ctx, req).Return(entity.WalletBalanceProjection{}, context.DeadlineExceeded)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.Get).ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusInternalServerError, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.InternalServerError,
			Message: "internal server error",
		},
	})
}

func (s *WalletHandlerTestSuite) TestGetNotFound() {
	req := &request.GetWallet{
		WalletID: "id1",
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().Get(s.ctx, req).Return(entity.WalletBalanceProjection{}, entity.ErrEntityNotFound)

	recorder := httptest.NewRecorder()
	handler.WrapDefaultBasic(s.handler.Get).ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil).WithContext(s.ctx))

	s.statusCompare(recorder.Code, http.StatusNotFound, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.NotFoundError,
			Message: "not found",
		},
	})
}

func (s *WalletHandlerTestSuite) TestDebitTransferSuccess() {
	req := &request.DebitTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}
	expectedBody := response.WalletEvent{
		ID:          "event1",
		ReferenceID: "ref1",
		WalletID:    "id1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.ctx = s.buildContext(req.WalletID, req.TransferID)

	s.svcMock.EXPECT().DebitTransfer(s.ctx, req).Return(entity.WalletEvent{
		ID:          expectedBody.ID,
		ReferenceID: expectedBody.ReferenceID,
		WalletID:    expectedBody.WalletID,
		Amount:      expectedBody.Amount,
		Status:      expectedBody.Status,
	}, nil)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.DebitTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusOK, recorder.Body.String(), expectedBody)
}

func (s *WalletHandlerTestSuite) TestDebitTransferFailure() {
	req := &request.DebitTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().DebitTransfer(s.ctx, req).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.DebitTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusInternalServerError, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.InternalServerError,
			Message: "internal server error",
		},
	})
}

func (s *WalletHandlerTestSuite) TestDebitTransferNotFound() {
	req := &request.DebitTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().DebitTransfer(s.ctx, req).Return(entity.WalletEvent{}, entity.ErrEntityNotFound)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.DebitTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusNotFound, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.NotFoundError,
			Message: "not found",
		},
	})
}

func (s *WalletHandlerTestSuite) TestCreditTransferSuccess() {
	req := &request.CreditTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}
	expectedBody := response.WalletEvent{
		ID:          "event1",
		ReferenceID: "ref1",
		WalletID:    "id1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().CreditTransfer(s.ctx, req).Return(entity.WalletEvent{
		ID:          expectedBody.ID,
		ReferenceID: expectedBody.ReferenceID,
		WalletID:    expectedBody.WalletID,
		Amount:      expectedBody.Amount,
		Status:      expectedBody.Status,
	}, nil)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.CreditTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusOK, recorder.Body.String(), expectedBody)
}

func (s *WalletHandlerTestSuite) TestCreditTransferFailure() {
	req := &request.CreditTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().CreditTransfer(s.ctx, req).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.CreditTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusInternalServerError, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.InternalServerError,
			Message: "internal server error",
		},
	})
}

func (s *WalletHandlerTestSuite) TestCreditTransferNotFound() {
	req := &request.CreditTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.ctx = s.buildContext(req.WalletID, "")

	s.svcMock.EXPECT().CreditTransfer(s.ctx, req).Return(entity.WalletEvent{}, entity.ErrEntityNotFound)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.CreditTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusNotFound, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.NotFoundError,
			Message: "not found",
		},
	})
}

func (s *WalletHandlerTestSuite) TestRevertTransferSuccess() {
	req := &request.RevertTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
	}
	expectedBody := response.WalletEvent{
		ID:          "event1",
		ReferenceID: "ref1",
		WalletID:    "id1",
		Status:      entity.TransferStatusFailed,
	}

	s.ctx = s.buildContext(req.WalletID, req.TransferID)

	s.svcMock.EXPECT().RevertTransfer(s.ctx, req).Return(entity.WalletEvent{
		ID:          expectedBody.ID,
		ReferenceID: expectedBody.ReferenceID,
		WalletID:    expectedBody.WalletID,
		Status:      expectedBody.Status,
	}, nil)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.RevertTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusOK, recorder.Body.String(), expectedBody)
}

func (s *WalletHandlerTestSuite) TestRevertTransferFailure() {
	req := &request.RevertTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
	}

	s.ctx = s.buildContext(req.WalletID, req.TransferID)

	s.svcMock.EXPECT().RevertTransfer(s.ctx, req).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.RevertTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusInternalServerError, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.InternalServerError,
			Message: "internal server error",
		},
	})
}

func (s *WalletHandlerTestSuite) TestRevertTransferNotFound() {
	req := &request.RevertTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
	}

	s.ctx = s.buildContext(req.WalletID, req.TransferID)

	s.svcMock.EXPECT().RevertTransfer(s.ctx, req).Return(entity.WalletEvent{}, entity.ErrEntityNotFound)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.RevertTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusNotFound, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.NotFoundError,
			Message: "not found",
		},
	})
}

func (s *WalletHandlerTestSuite) TestCompleteTransferSuccess() {
	req := &request.CompleteTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
	}
	expectedBody := response.WalletEvent{
		ID:          "event1",
		ReferenceID: "ref1",
		WalletID:    "id1",
		Status:      entity.TransferStatusCompleted,
	}

	s.ctx = s.buildContext(req.WalletID, req.TransferID)

	s.svcMock.EXPECT().CompleteTransfer(s.ctx, req).Return(entity.WalletEvent{
		ID:          expectedBody.ID,
		ReferenceID: expectedBody.ReferenceID,
		WalletID:    expectedBody.WalletID,
		Status:      expectedBody.Status,
	}, nil)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.CompleteTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusOK, recorder.Body.String(), expectedBody)
}

func (s *WalletHandlerTestSuite) TestCompleteTransferFailure() {
	req := &request.CompleteTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
	}

	s.ctx = s.buildContext(req.WalletID, req.TransferID)

	s.svcMock.EXPECT().CompleteTransfer(s.ctx, req).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.CompleteTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusInternalServerError, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.InternalServerError,
			Message: "internal server error",
		},
	})
}

func (s *WalletHandlerTestSuite) TestCompleteTransferNotFound() {
	req := &request.CompleteTransfer{
		WalletID:    "id1",
		TransferID:  "transfer1",
		ReferenceID: "ref1",
	}

	s.ctx = s.buildContext(req.WalletID, req.TransferID)

	s.svcMock.EXPECT().CompleteTransfer(s.ctx, req).Return(entity.WalletEvent{}, entity.ErrEntityNotFound)

	recorder := httptest.NewRecorder()

	handler.WrapDefaultBasic(s.handler.CompleteTransfer).ServeHTTP(recorder, httptest.NewRequest("POST", "/", testutils.ToJSONReader(s.T(), req)).WithContext(s.ctx))
	s.statusCompare(recorder.Code, http.StatusNotFound, recorder.Body.String(), render.ErrorResponse{
		Error: &render.Error{
			Status:  render.NotFoundError,
			Message: "not found",
		},
	})
}

func TestWalletHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(WalletHandlerTestSuite))
}
