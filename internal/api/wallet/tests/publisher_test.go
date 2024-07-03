package wallet_test

import (
	"context"
	"testing"
	"time"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/buni/wallet/internal/pkg/pubsub/jetstream"
	"github.com/buni/wallet/internal/pkg/testing/dt"
	"github.com/gofrs/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
)

type WalletPublisherSuite struct {
	jetstreamConn        nats.JetStreamContext
	publisher            pubsub.Publisher
	walletEventPublisher *wallet.Publisher
	suite.Suite
}

func (s *WalletPublisherSuite) SetupSuite() {
	var err error

	s.jetstreamConn, err = dt.NATSConn.JetStream()
	s.NoError(err)

	s.publisher, err = jetstream.NewJetStreamPublisher(s.jetstreamConn)
	s.NoError(err)

	s.walletEventPublisher = wallet.NewPublisher(s.publisher)
}

func (s *WalletPublisherSuite) TestPublishSuccessfully() {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)

	err := s.walletEventPublisher.PublishCreated(ctx, entity.WalletEvent{
		ID:          uuid.Must(uuid.NewV7()).String(),
		Version:     entity.WalletEventVersionOne,
		TransferID:  uuid.Must(uuid.NewV7()).String(),
		ReferenceID: uuid.Must(uuid.NewV7()).String(),
		WalletID:    uuid.Must(uuid.NewV7()).String(),
		EventType:   entity.EventTypeDebitTransfer,
		Status:      entity.TransferStatusCompleted,
	})
	s.NoError(err)

	sub, err := s.jetstreamConn.PullSubscribe("wallet_events.created", "test", nats.AckExplicit(), nats.MaxAckPending(1))
	s.NoError(err)

	msgs, err := sub.Fetch(2)
	s.NoError(err)
	s.Len(msgs, 1)
	err = msgs[0].Ack()
	s.NoError(err)
}

func TestWalletPublisherSuite(t *testing.T) {
	suite.Run(t, new(WalletPublisherSuite))
}
