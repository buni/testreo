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

type WalletProjectionRepositoryTestSuite struct {
	suite.Suite
	ctx            context.Context
	pgxPoolWrapper *pgxtx.TxWrapper
	repo           *wallet.ProjectionRepository
	tt             time.Time
}

func (s *WalletProjectionRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.pgxPoolWrapper = pgxtx.NewTxWrapper(dt.DB, pgx.TxOptions{})
	s.repo = wallet.NewProjectionRepository(s.pgxPoolWrapper)
	s.tt = time.Now().UTC().Truncate(time.Millisecond)
}

func (s *WalletProjectionRepositoryTestSuite) TearDownTest() {
	_, err := s.pgxPoolWrapper.Exec(s.ctx, "TRUNCATE wallet_projections")
	s.NoError(err)
}

func (s *WalletProjectionRepositoryTestSuite) newWalletProjection(walletID string) entity.WalletProjection {
	projection := entity.NewWalletProjection(
		walletID,
		uuid.Must(uuid.NewV7()).String(),
		decimal.NewFromInt(int64(rand.Intn(100000))),
		decimal.NewFromInt(int64(rand.Intn(100000))),
		decimal.NewFromInt(int64(rand.Intn(100000))),
	)
	return projection
}

func (s *WalletProjectionRepositoryTestSuite) newRandomProjection() entity.WalletProjection {
	return s.newWalletProjection(uuid.Must(uuid.NewV7()).String())
}

func (s *WalletProjectionRepositoryTestSuite) TestCreateSuccess() {
	_, err := s.repo.Create(s.ctx, s.newRandomProjection())
	s.NoError(err)
}

func (s *WalletProjectionRepositoryTestSuite) TestGetSuccess() {
	want := s.newRandomProjection()
	want, err := s.repo.Create(s.ctx, want)
	s.NoError(err)

	got, err := s.repo.Get(s.ctx, want.WalletID)
	s.NoError(err)
	s.Equal(want, got)
}

func (s *WalletProjectionRepositoryTestSuite) TestGetNotFound() {
	_, err := s.repo.Get(s.ctx, uuid.Must(uuid.NewV7()).String())
	s.ErrorIs(err, entity.ErrEntityNotFound)
}

func (s *WalletProjectionRepositoryTestSuite) TestUpdateSuccess() {
	want := s.newRandomProjection()
	_, err := s.repo.Create(s.ctx, want)
	s.NoError(err)

	want.Balance = decimal.NewFromInt(100)
	want.LastEventID = uuid.Must(uuid.NewV7()).String()
	want.PendingCredit = decimal.NewFromInt(50)
	want.PendingDebit = decimal.NewFromInt(50)

	got, err := s.repo.Update(s.ctx, want)
	s.NoError(err)
	s.Equal(want, got)
}

func (s *WalletProjectionRepositoryTestSuite) TestUpdateNotFound() {
	want := s.newRandomProjection()

	got, err := s.repo.Update(s.ctx, want)
	s.ErrorIs(err, entity.ErrEntityNotFound)
	s.Empty(got)
}

func TestWalletProjectionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(WalletProjectionRepositoryTestSuite))
}
