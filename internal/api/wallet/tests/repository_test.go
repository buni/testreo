package wallet_test

import (
	"context"
	"testing"
	"time"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/database/pgxtx"
	"github.com/buni/wallet/internal/pkg/testing/dt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"
)

type WalletRepositoryTestSuite struct {
	suite.Suite
	ctx            context.Context
	pgxPoolWrapper *pgxtx.TxWrapper
	repo           *wallet.Repository
	tt             time.Time
}

func (s *WalletRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.pgxPoolWrapper = pgxtx.NewTxWrapper(dt.DB, pgx.TxOptions{})
	s.repo = wallet.NewRepository(s.pgxPoolWrapper)
	s.tt = time.Now().UTC().Truncate(time.Millisecond)
}

func (s *WalletRepositoryTestSuite) TearDownTest() {
	_, err := s.pgxPoolWrapper.Exec(s.ctx, "TRUNCATE wallets")
	s.NoError(err)
}

func (s *WalletRepositoryTestSuite) newWallet() entity.Wallet {
	wallet, err := entity.NewWallet(uuid.Must(uuid.NewV7()).String())
	s.NoError(err)
	return wallet
}

func (s *WalletRepositoryTestSuite) TestCreateSuccess() {
	wallet, err := s.repo.Create(s.ctx, s.newWallet())
	s.NoError(err)
	s.NotEmpty(wallet.ID)
	s.NotEmpty(wallet.CreatedAt)
}

func (s *WalletRepositoryTestSuite) TestCreateDuplicateReferenceError() {
	want := s.newWallet()
	_, err := s.repo.Create(s.ctx, want)
	s.NoError(err)

	want.ID = uuid.Must(uuid.NewV7()).String()
	_, err = s.repo.Create(s.ctx, want)
	s.Error(err)
}

func (s *WalletRepositoryTestSuite) TestGetSuccess() {
	want := s.newWallet()
	want, err := s.repo.Create(s.ctx, want)
	s.NoError(err)

	got, err := s.repo.Get(s.ctx, want.ID)
	s.NoError(err)
	s.Equal(want, got)
}

func (s *WalletRepositoryTestSuite) TestGetNotFound() {
	_, err := s.repo.Get(s.ctx, uuid.Must(uuid.NewV7()).String())
	s.ErrorIs(err, entity.ErrEntityNotFound)
}

func TestWalletRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(WalletRepositoryTestSuite))
}
