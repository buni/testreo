package wallet_test

import (
	"context"
	"testing"

	contract_mock "github.com/buni/wallet/internal/api/app/contract/mock"
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/app/request"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/testing/testutils"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type WalletServiceTestSuite struct {
	suite.Suite
	ctrl               *gomock.Controller
	repoMock           *contract_mock.MockWalletRepository
	eventRepoMock      *contract_mock.MockWalletEventRepository
	projectionRepoMock *contract_mock.MockWalletProjectionRepository
	publisherMock      *contract_mock.MockWalletEventPublisher
	svc                *wallet.Service
}

func (s *WalletServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.repoMock = contract_mock.NewMockWalletRepository(s.ctrl)
	s.eventRepoMock = contract_mock.NewMockWalletEventRepository(s.ctrl)
	s.projectionRepoMock = contract_mock.NewMockWalletProjectionRepository(s.ctrl)
	s.publisherMock = contract_mock.NewMockWalletEventPublisher(s.ctrl)
	s.svc = wallet.NewService(s.repoMock, s.projectionRepoMock, s.eventRepoMock, s.publisherMock, testutils.NoopTransactionManager{})
}

func (s *WalletServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *WalletServiceTestSuite) TestCreateSuccess() {
	req := &request.CreateWallet{
		ReferenceID: "ref-id",
	}

	wallet := entity.Wallet{
		ReferenceID: req.ReferenceID,
	}

	s.repoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(wallet, cmpopts.IgnoreFields(entity.Wallet{}, "ID", "CreatedAt", "UpdatedAt"))).Return(entity.Wallet{
		ReferenceID: req.ReferenceID,
	}, nil)
	s.projectionRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(entity.WalletProjection{}, nil)

	wallet, err := s.svc.Create(context.Background(), req)
	s.NoError(err)
	s.NotEmpty(wallet)
}

func (s *WalletServiceTestSuite) TestCreateCreateWalletError() {
	req := &request.CreateWallet{
		ReferenceID: "ref-id",
	}

	wallet := entity.Wallet{
		ReferenceID: req.ReferenceID,
	}

	s.repoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(wallet, cmpopts.IgnoreFields(entity.Wallet{}, "ID", "CreatedAt", "UpdatedAt"))).Return(entity.Wallet{}, context.DeadlineExceeded)

	wallet, err := s.svc.Create(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(wallet)
}

func (s *WalletServiceTestSuite) TestCreateProjectionError() {
	req := &request.CreateWallet{
		ReferenceID: "ref-id",
	}

	wallet := entity.Wallet{
		ReferenceID: req.ReferenceID,
	}

	s.repoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(wallet, cmpopts.IgnoreFields(entity.Wallet{}, "ID", "CreatedAt", "UpdatedAt"))).Return(entity.Wallet{
		ReferenceID: req.ReferenceID,
	}, nil)
	s.projectionRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(entity.WalletProjection{}, context.DeadlineExceeded)

	wallet, err := s.svc.Create(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *WalletServiceTestSuite) TestGetSuccess() {
	req := &request.GetWallet{
		WalletID: "wallet-id",
	}

	wallet := entity.Wallet{
		ID:          req.WalletID,
		ReferenceID: "ref-id",
	}

	projection := entity.WalletProjection{
		WalletID: req.WalletID,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(wallet, nil)
	s.projectionRepoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(projection, nil)

	result, err := s.svc.Get(context.Background(), req)
	s.NoError(err)
	s.Equal(entity.WalletBalanceProjection{
		Wallet:           wallet,
		WalletProjection: projection,
	}, result)
}

func (s *WalletServiceTestSuite) TestGetWalletError() {
	req := &request.GetWallet{
		WalletID: "wallet-id",
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{}, context.DeadlineExceeded)

	result, err := s.svc.Get(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestGetProjectionError() {
	req := &request.GetWallet{
		WalletID: "wallet-id",
	}

	wallet := entity.Wallet{
		ID:          req.WalletID,
		ReferenceID: "ref-id",
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(wallet, nil)
	s.projectionRepoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.WalletProjection{}, context.DeadlineExceeded)

	result, err := s.svc.Get(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestDebitTransferSuccess() {
	req := &request.DebitTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		EventType:   entity.EventTypeDebitTransfer,
		Status:      req.Status,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(nil)

	result, err := s.svc.DebitTransfer(context.Background(), req)
	s.NoError(err)
	s.Equal(event, result)
}

func (s *WalletServiceTestSuite) TestDebitTransferGetError() {
	req := &request.DebitTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{}, context.DeadlineExceeded)

	result, err := s.svc.DebitTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestDebitTransferCreateError() {
	req := &request.DebitTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		EventType:   entity.EventTypeDebitTransfer,
		Status:      req.Status,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	result, err := s.svc.DebitTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestDebitTransferPublishCreatedError() {
	req := &request.DebitTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		EventType:   entity.EventTypeDebitTransfer,
		Status:      req.Status,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(context.DeadlineExceeded)

	_, err := s.svc.DebitTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *WalletServiceTestSuite) TestCreditTransferSuccess() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		EventType:   entity.EventTypeCreditTransfer,
		Status:      req.Status,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), req.WalletID).Return([]entity.WalletEvent{
		{
			Version:     entity.WalletEventVersionOne,
			TransferID:  req.TransferID,
			ReferenceID: req.ReferenceID,
			WalletID:    req.WalletID,
			Amount:      decimal.NewFromInt(100),
			EventType:   entity.EventTypeDebitTransfer,
			Status:      entity.TransferStatusCompleted,
		},
	}, nil)

	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(nil)

	result, err := s.svc.CreditTransfer(context.Background(), req)
	s.NoError(err)
	s.Equal(event, result)
}

func (s *WalletServiceTestSuite) TestCreditTransferGetError() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{}, context.DeadlineExceeded)

	result, err := s.svc.CreditTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestCreditTransferListByWalletIDError() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), req.WalletID).Return(nil, context.DeadlineExceeded)

	result, err := s.svc.CreditTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestCreditTransferCreateError() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		EventType:   entity.EventTypeCreditTransfer,
		Status:      req.Status,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), req.WalletID).Return([]entity.WalletEvent{
		{
			Version:     entity.WalletEventVersionOne,
			TransferID:  req.TransferID,
			ReferenceID: req.ReferenceID,
			WalletID:    req.WalletID,
			Amount:      decimal.NewFromInt(100),
			EventType:   entity.EventTypeDebitTransfer,
			Status:      entity.TransferStatusCompleted,
		},
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	result, err := s.svc.CreditTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestCreditTransferPublishCreatedError() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(100),
		Status:      entity.TransferStatusPending,
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		EventType:   entity.EventTypeCreditTransfer,
		Status:      req.Status,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), req.WalletID).Return([]entity.WalletEvent{
		{
			Version:     entity.WalletEventVersionOne,
			TransferID:  req.TransferID,
			ReferenceID: req.ReferenceID,
			WalletID:    req.WalletID,
			Amount:      decimal.NewFromInt(100),
			EventType:   entity.EventTypeDebitTransfer,
			Status:      entity.TransferStatusCompleted,
		},
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(context.DeadlineExceeded)

	_, err := s.svc.CreditTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *WalletServiceTestSuite) TestCreditTransferInsufficientBalance() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(50),
		Status:      entity.TransferStatusPending,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), req.WalletID).Return([]entity.WalletEvent{
		{
			Version:     entity.WalletEventVersionOne,
			TransferID:  req.TransferID,
			ReferenceID: req.ReferenceID,
			WalletID:    req.WalletID,
			Amount:      decimal.NewFromInt(20),
			EventType:   entity.EventTypeDebitTransfer,
			Status:      entity.TransferStatusCompleted,
		},
	}, nil)

	_, err := s.svc.CreditTransfer(context.Background(), req)
	s.ErrorIs(err, entity.ErrInsufficientBalance)
}

func (s *WalletServiceTestSuite) TestCreditTransferProcessEventsError() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(50),
		Status:      entity.TransferStatusPending,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), req.WalletID).Return([]entity.WalletEvent{
		{
			Version:     entity.WalletEventVersionOne,
			TransferID:  req.TransferID,
			ReferenceID: req.ReferenceID,
			WalletID:    req.WalletID,
			Amount:      decimal.NewFromInt(20),
			EventType:   5,
			Status:      entity.TransferStatusCompleted,
		},
	}, nil)

	_, err := s.svc.CreditTransfer(context.Background(), req)
	s.ErrorIs(err, entity.ErrUnsupportedEventType)
}

func (s *WalletServiceTestSuite) TestCreditTransferNegativeAmountError() {
	req := &request.CreditTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(-50),
		Status:      entity.TransferStatusPending,
	}

	_, err := s.svc.CreditTransfer(context.Background(), req)
	s.ErrorIs(err, entity.ErrNegativeAmount)
}

func (s *WalletServiceTestSuite) TestDebitTransferNegativeAmountError() {
	req := &request.DebitTransfer{
		WalletID:    "wallet-id",
		ReferenceID: "123",
		TransferID:  "1234",
		Amount:      decimal.NewFromInt(-50),
		Status:      entity.TransferStatusPending,
	}

	_, err := s.svc.DebitTransfer(context.Background(), req)
	s.ErrorIs(err, entity.ErrNegativeAmount)
}

func (s *WalletServiceTestSuite) TestCompleteTransferSuccess() {
	req := &request.CompleteTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		EventType:   entity.EventTypeUpdateTransferStatus,
		Status:      entity.TransferStatusCompleted,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(nil)

	result, err := s.svc.CompleteTransfer(context.Background(), req)
	s.NoError(err)
	s.Equal(event, result)
}

func (s *WalletServiceTestSuite) TestCompleteTransferGetError() {
	req := &request.CompleteTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{}, context.DeadlineExceeded)

	result, err := s.svc.CompleteTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestCompleteTransferCreateError() {
	req := &request.CompleteTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		EventType:   entity.EventTypeUpdateTransferStatus,
		Status:      entity.TransferStatusCompleted,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	result, err := s.svc.CompleteTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestCompleteTransferPublishCreatedError() {
	req := &request.CompleteTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		EventType:   entity.EventTypeUpdateTransferStatus,
		Status:      entity.TransferStatusCompleted,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(context.DeadlineExceeded)

	_, err := s.svc.CompleteTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *WalletServiceTestSuite) TestRevertTransferSuccess() {
	req := &request.RevertTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		EventType:   entity.EventTypeUpdateTransferStatus,
		Status:      entity.TransferStatusFailed,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(nil)

	result, err := s.svc.RevertTransfer(context.Background(), req)
	s.NoError(err)
	s.Equal(event, result)
}

func (s *WalletServiceTestSuite) TestRevertTransferGetError() {
	req := &request.RevertTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{}, context.DeadlineExceeded)

	result, err := s.svc.RevertTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestRevertTransferCreateError() {
	req := &request.RevertTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		EventType:   entity.EventTypeUpdateTransferStatus,
		Status:      entity.TransferStatusFailed,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(entity.WalletEvent{}, context.DeadlineExceeded)

	result, err := s.svc.RevertTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(result)
}

func (s *WalletServiceTestSuite) TestRevertTransferPublishCreatedError() {
	req := &request.RevertTransfer{
		WalletID:    "wallet-id",
		TransferID:  "1234",
		ReferenceID: "123",
	}

	event := entity.WalletEvent{
		Version:     entity.WalletEventVersionOne,
		TransferID:  req.TransferID,
		ReferenceID: req.ReferenceID,
		WalletID:    req.WalletID,
		EventType:   entity.EventTypeUpdateTransferStatus,
		Status:      entity.TransferStatusFailed,
	}

	s.repoMock.EXPECT().Get(gomock.Any(), req.WalletID).Return(entity.Wallet{
		ID: req.WalletID,
	}, nil)
	s.eventRepoMock.EXPECT().Create(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(event, nil)
	s.publisherMock.EXPECT().PublishCreated(gomock.Any(), testutils.NewMatcher(event, cmpopts.IgnoreFields(entity.WalletEvent{}, "ID", "CreatedAt"))).Return(context.DeadlineExceeded)

	_, err := s.svc.RevertTransfer(context.Background(), req)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *WalletServiceTestSuite) TestRebuildWalletProjectionSuccess() {
	event := &entity.WalletEvent{
		ID:          "123",
		Version:     entity.WalletEventVersionOne,
		TransferID:  "1234",
		ReferenceID: "123",
		WalletID:    "wallet-id",
		Amount:      decimal.NewFromInt(100),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusCompleted,
	}
	projection := entity.WalletProjection{
		WalletID:      "wallet-id",
		Balance:       decimal.NewFromInt(100),
		PendingDebit:  decimal.Decimal{},
		PendingCredit: decimal.Decimal{},
		LastEventID:   event.ID,
	}
	s.projectionRepoMock.EXPECT().Get(gomock.Any(), event.WalletID).Return(entity.WalletProjection{
		WalletID: "12",
	}, nil)

	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), event.WalletID).Return([]entity.WalletEvent{*event}, nil)
	s.projectionRepoMock.EXPECT().Update(gomock.Any(), testutils.NewMatcher(projection, cmpopts.IgnoreFields(entity.WalletProjection{}, "UpdatedAt", "CreatedAt"))).Return(entity.WalletProjection{}, nil)
	got, err := s.svc.RebuildWalletProjection(context.Background(), event)

	s.NoError(err)
	s.NotEmpty(got.UpdatedAt)
	got.UpdatedAt = projection.UpdatedAt
	s.Equal(projection, got)
}

func (s *WalletServiceTestSuite) TestRebuildWalletProjectionNoopOnSameLastEvent() {
	event := &entity.WalletEvent{
		ID:          "123",
		Version:     entity.WalletEventVersionOne,
		TransferID:  "1234",
		ReferenceID: "123",
		WalletID:    "wallet-id",
		Amount:      decimal.NewFromInt(100),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusCompleted,
	}

	s.projectionRepoMock.EXPECT().Get(gomock.Any(), event.WalletID).Return(entity.WalletProjection{
		WalletID:    "123",
		LastEventID: event.ID,
	}, nil)

	got, err := s.svc.RebuildWalletProjection(context.Background(), event)
	s.Empty(got)
	s.NoError(err)
}

func (s *WalletServiceTestSuite) TestRebuildWalletProjectionGetError() {
	event := &entity.WalletEvent{
		ID:          "123",
		Version:     entity.WalletEventVersionOne,
		TransferID:  "1234",
		ReferenceID: "123",
		WalletID:    "wallet-id",
		Amount:      decimal.NewFromInt(100),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusCompleted,
	}

	s.projectionRepoMock.EXPECT().Get(gomock.Any(), event.WalletID).Return(entity.WalletProjection{}, context.DeadlineExceeded)

	got, err := s.svc.RebuildWalletProjection(context.Background(), event)

	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(got)
}

func (s *WalletServiceTestSuite) TestRebuildWalletProjectionListByWalletIDError() {
	event := &entity.WalletEvent{
		ID:          "123",
		Version:     entity.WalletEventVersionOne,
		TransferID:  "1234",
		ReferenceID: "123",
		WalletID:    "wallet-id",
		Amount:      decimal.NewFromInt(100),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusCompleted,
	}

	s.projectionRepoMock.EXPECT().Get(gomock.Any(), event.WalletID).Return(entity.WalletProjection{
		WalletID: "12",
	}, nil)

	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), event.WalletID).Return(nil, context.DeadlineExceeded)

	got, err := s.svc.RebuildWalletProjection(context.Background(), event)

	s.ErrorIs(err, context.DeadlineExceeded)
	s.Empty(got)
}

func (s *WalletServiceTestSuite) TestRebuildWalletProjectionUpdateError() {
	event := &entity.WalletEvent{
		ID:          "123",
		Version:     entity.WalletEventVersionOne,
		TransferID:  "1234",
		ReferenceID: "123",
		WalletID:    "wallet-id",
		Amount:      decimal.NewFromInt(100),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusCompleted,
	}
	projection := entity.WalletProjection{
		WalletID:      "wallet-id",
		Balance:       decimal.NewFromInt(100),
		PendingDebit:  decimal.Decimal{},
		PendingCredit: decimal.Decimal{},
		LastEventID:   event.ID,
	}
	s.projectionRepoMock.EXPECT().Get(gomock.Any(), event.WalletID).Return(entity.WalletProjection{
		WalletID: "12",
	}, nil)

	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), event.WalletID).Return([]entity.WalletEvent{*event}, nil)
	s.projectionRepoMock.EXPECT().Update(gomock.Any(), testutils.NewMatcher(projection, cmpopts.IgnoreFields(entity.WalletProjection{}, "UpdatedAt", "CreatedAt"))).Return(entity.WalletProjection{}, context.DeadlineExceeded)
	_, err := s.svc.RebuildWalletProjection(context.Background(), event)

	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *WalletServiceTestSuite) TestRebuildWalletProjectionProcessEventsError() {
	event := &entity.WalletEvent{
		ID:          "123",
		Version:     entity.WalletEventVersionOne,
		TransferID:  "1234",
		ReferenceID: "123",
		WalletID:    "wallet-id",
		Amount:      decimal.NewFromInt(100),
		EventType:   entity.EventTypeDebitTransfer + 4,
		Status:      entity.TransferStatusCompleted,
	}

	s.projectionRepoMock.EXPECT().Get(gomock.Any(), event.WalletID).Return(entity.WalletProjection{
		WalletID: "12",
	}, nil)

	s.eventRepoMock.EXPECT().ListByWalletID(gomock.Any(), event.WalletID).Return([]entity.WalletEvent{*event}, nil)
	_, err := s.svc.RebuildWalletProjection(context.Background(), event)

	s.ErrorIs(err, entity.ErrUnsupportedEventType)
}

func TestWalletServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WalletServiceTestSuite))
}
