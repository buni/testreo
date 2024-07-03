package wallet_test

import (
	"context"
	"testing"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/pubsub"
	pubsub_mock "github.com/buni/wallet/internal/pkg/pubsub/mock"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type PublisherTestSuite struct {
	suite.Suite
	ctrl      *gomock.Controller
	pubMock   *pubsub_mock.MockPublisher
	publisher *wallet.Publisher
}

func (s *PublisherTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.pubMock = pubsub_mock.NewMockPublisher(s.ctrl)
	s.publisher = wallet.NewPublisher(s.pubMock)
}

func (s *PublisherTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *PublisherTestSuite) TestPublishCreated() {
	event := entity.WalletEvent{
		ID:          uuid.Must(uuid.NewV7()).String(),
		Version:     entity.WalletEventVersionOne,
		TransferID:  string(uuid.Must(uuid.NewV7()).String()),
		ReferenceID: string(uuid.Must(uuid.NewV7()).String()),
		WalletID:    string(uuid.Must(uuid.NewV7()).String()),
		Amount:      decimal.NewFromInt(1000),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusPending,
	}

	msg, err := pubsub.NewJSONMessage(event, nil)
	s.NoError(err)

	msg.Key = entity.WalletEventsCreated
	msg.Topic = entity.WalletEventsTopic

	s.pubMock.EXPECT().Publish(gomock.Any(), msg).Return(nil)
	err = s.publisher.PublishCreated(context.Background(), event)

	s.NoError(err)
}

func (s *PublisherTestSuite) TestPublishCreatedPublishError() {
	event := entity.WalletEvent{
		ID:          uuid.Must(uuid.NewV7()).String(),
		Version:     entity.WalletEventVersionOne,
		TransferID:  string(uuid.Must(uuid.NewV7()).String()),
		ReferenceID: string(uuid.Must(uuid.NewV7()).String()),
		WalletID:    string(uuid.Must(uuid.NewV7()).String()),
		Amount:      decimal.NewFromInt(1000),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusPending,
	}

	msg, err := pubsub.NewJSONMessage(event, nil)
	s.NoError(err)

	msg.Key = entity.WalletEventsCreated
	msg.Topic = entity.WalletEventsTopic

	s.pubMock.EXPECT().Publish(gomock.Any(), msg).Return(context.DeadlineExceeded)
	err = s.publisher.PublishCreated(context.Background(), event)

	s.ErrorIs(err, context.DeadlineExceeded)
}

func TestPublisherTestSuit(t *testing.T) {
	suite.Run(t, new(PublisherTestSuite))
}
