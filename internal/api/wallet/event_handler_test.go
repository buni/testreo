package wallet_test

import (
	"context"
	"testing"

	contract_mock "github.com/buni/wallet/internal/api/app/contract/mock"
	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/testing/testutils"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/exp/rand"
)

type WalletEventCreatedHandler struct {
	suite.Suite
	svcMock *contract_mock.MockWalletService
	ctrl    *gomock.Controller
	handler *wallet.EventCreatedHandler
}

func (s *WalletEventCreatedHandler) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.svcMock = contract_mock.NewMockWalletService(s.ctrl)
	s.handler = wallet.NewWalletEventCreatedHandler(s.svcMock, testutils.NoopTransactionManager{})
}

func (s *WalletEventCreatedHandler) TearDownTest() {
	s.ctrl.Finish()
}

func (s *WalletEventCreatedHandler) TestHandlerName() {
	s.Equal("WalletEventCreatedHandler", s.handler.HandlerName())
}

func (s *WalletEventCreatedHandler) TestTopic() {
	s.Equal("wallet_events.created", s.handler.Topic())
}

func (s *WalletEventCreatedHandler) TestSubscriberOptions() {
	s.Empty(s.handler.SubscriberOptions())
}

func (s *WalletEventCreatedHandler) TestHandleSuccess() {
	event, err := entity.NewWalletEvent(uuid.Must(uuid.NewV7()).String(),
		uuid.Must(uuid.NewV7()).String(),
		uuid.Must(uuid.NewV7()).String(),
		decimal.NewFromInt(int64(rand.Intn(100000))),
		entity.WalletEventType(rand.Intn(2)+1),
		entity.TransferStatus(rand.Intn(2)+1),
	)
	s.NoError(err)
	s.svcMock.EXPECT().RebuildWalletProjection(gomock.Any(), &event).Return(entity.WalletProjection{}, nil)

	err = s.handler.Handle(context.Background(), &event, nil)
	s.NoError(err)
}

func (s *WalletEventCreatedHandler) TestHandleError() {
	event, err := entity.NewWalletEvent(uuid.Must(uuid.NewV7()).String(),
		uuid.Must(uuid.NewV7()).String(),
		uuid.Must(uuid.NewV7()).String(),
		decimal.NewFromInt(int64(rand.Intn(100000))),
		entity.WalletEventType(rand.Intn(2)+1),
		entity.TransferStatus(rand.Intn(2)+1),
	)
	s.NoError(err)

	s.svcMock.EXPECT().RebuildWalletProjection(gomock.Any(), &event).Return(entity.WalletProjection{}, context.DeadlineExceeded)

	err = s.handler.Handle(context.Background(), &event, nil)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func TestWalletEventCreatedHandler(t *testing.T) {
	suite.Run(t, new(WalletEventCreatedHandler))
}
