package wallet_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/database/pgxtx"
	"github.com/buni/wallet/internal/pkg/testing/dt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type WalletEventRepositoryTestSuite struct {
	suite.Suite
	ctx            context.Context
	pgxPoolWrapper *pgxtx.TxWrapper
	repo           *wallet.EventRepository
	tt             time.Time
}

func (s *WalletEventRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.pgxPoolWrapper = pgxtx.NewTxWrapper(dt.DB, pgx.TxOptions{})
	s.repo = wallet.NewEventRepository(s.pgxPoolWrapper)
	s.tt = time.Now().UTC().Truncate(time.Millisecond)
}

func (s *WalletEventRepositoryTestSuite) TearDownTest() {
	_, err := s.pgxPoolWrapper.Exec(s.ctx, "TRUNCATE wallet_events")
	s.NoError(err)
}

func (s *WalletEventRepositoryTestSuite) newWalletEvent(
	walletID, transferID, referenceID string,
	eventType entity.WalletEventType,
	transferStatus entity.TransferStatus,
	amount decimal.Decimal,
) entity.WalletEvent {
	wallet, err := entity.NewWalletEvent(walletID, transferID, referenceID, amount, eventType, transferStatus)
	s.NoError(err)
	return wallet
}

func (s *WalletEventRepositoryTestSuite) newRandomWalletEvent() entity.WalletEvent {
	return s.newWalletEvent(
		uuid.Must(uuid.NewV7()).String(),
		uuid.Must(uuid.NewV7()).String(),
		uuid.Must(uuid.NewV7()).String(),
		entity.WalletEventType(rand.Intn(2)+1),
		entity.TransferStatus(rand.Intn(2)+1),
		decimal.NewFromInt(int64(rand.Intn(100000))),
	)
}

func (s *WalletEventRepositoryTestSuite) TestCreateSuccess() {
	walletEvent, err := s.repo.Create(s.ctx, s.newRandomWalletEvent())
	s.NoError(err)
	s.NotEmpty(walletEvent.ID)
	s.NotEmpty(walletEvent.CreatedAt)
}

func (s *WalletEventRepositoryTestSuite) seedEvents(num int, walletID string) []entity.WalletEvent {
	result := make([]entity.WalletEvent, 0, num)
	for range num {
		walletEvent := s.newRandomWalletEvent()
		walletEvent.WalletID = walletID
		_, err := s.repo.Create(s.ctx, walletEvent)
		s.NoError(err)

		result = append(result, walletEvent)
	}

	return result
}

func (s *WalletEventRepositoryTestSuite) TestCreateDuplicateTransferIDForEventType() {
	walletEvent, err := s.repo.Create(s.ctx, s.newRandomWalletEvent())
	s.NoError(err)

	walletEvent.ID = uuid.Must(uuid.NewV7()).String()

	walletEvent, err = s.repo.Create(s.ctx, walletEvent)

	s.Error(err)
}

func (s *WalletEventRepositoryTestSuite) TestListByWalletIDSuccess() {
	walletID := uuid.Must(uuid.NewV7()).String()
	want := s.seedEvents(5, walletID)
	_ = s.seedEvents(5, uuid.Must(uuid.NewV7()).String())

	events, err := s.repo.ListByWalletID(s.ctx, walletID)
	s.NoError(err)
	s.Equal(want, events)
}

func (s *WalletEventRepositoryTestSuite) TestListByWalletIDEmpty() {
	walletID := uuid.Must(uuid.NewV7()).String()

	events, err := s.repo.ListByWalletID(s.ctx, walletID)
	s.NoError(err)
	s.Empty(events)
	s.NotNil(events)
}

func TestWalletEventRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(WalletEventRepositoryTestSuite))
}
