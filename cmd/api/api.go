package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/buni/wallet/internal/api/wallet"
	"github.com/buni/wallet/internal/pkg/configuration"
	"github.com/buni/wallet/internal/pkg/database/pgxtx"
	"github.com/buni/wallet/internal/pkg/pubsub/jetstream"
	"github.com/buni/wallet/internal/pkg/pubsub/outbox"
	"github.com/buni/wallet/internal/pkg/render/errorhandler"
	"github.com/buni/wallet/internal/pkg/server"
	httpin_integration "github.com/ggicci/httpin/integration" //nolint
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Start the api service",
		Long:  "Start the api service",
	}
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return main()
	}

	return cmd
}

func main() error {
	httpin_integration.UseGochiURLParam("path", chi.URLParam)
	errorhandler.RegisterErrorHandler("validation_error_handler", errorhandler.ValidationErrorHandler)
	errorhandler.RegisterErrorHandler("validation_field_errors_handler", errorhandler.ValidationFieldErrorsHandler)
	errorhandler.RegisterErrorHandler("validation_field_error_handler", errorhandler.ValidationFieldErrorHandler)
	errorhandler.RegisterErrorHandler("not_found_error_handler", errorhandler.NotFoundErrorHandler)
	errorhandler.RegisterErrorHandler("unique_constraint_error_handler", errorhandler.ConflictErrorHandler)
	errorhandler.RegisterErrorHandler("insufficient_balance_error_handler", errorhandler.InsufficientBalanceErrorHandler)
	errorhandler.RegisterErrorHandler("negative_amount_error_handler", errorhandler.NegativeAmountErrorHandler)

	srv, err := server.NewServer(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	ctx := srv.Context

	config, err := configuration.NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	pgxConf, err := pgxpool.ParseConfig(config.Database.ToURL())
	if err != nil {
		return fmt.Errorf("failed to parse pgx config: %w", err)
	}

	pgxPool, err := pgxpool.NewWithConfig(ctx, pgxConf)
	if err != nil {
		return fmt.Errorf("failed to create pg session: %w", err)
	}

	err = pgxPool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping pg: %w", err)
	}

	txWrapper := pgxtx.NewTxWrapper(pgxPool, pgx.TxOptions{})

	txm := pgxtx.NewTransactionManager(pgxPool, pgx.TxOptions{})

	outboxRepo := outbox.NewPGxRepository(txWrapper)
	publisher := outbox.NewPublisher[any](outboxRepo, txm, jetstream.JetStreamPublisherType)

	walletRepo := wallet.NewRepository(txWrapper)
	walletEventRepo := wallet.NewEventRepository(txWrapper)
	walletProjectionRepo := wallet.NewProjectionRepository(txWrapper)
	walletEventPublisher := wallet.NewPublisher(publisher)
	walletSvc := wallet.NewService(walletRepo, walletProjectionRepo, walletEventRepo, walletEventPublisher, txm)
	walletHandler := wallet.NewHandler(walletSvc)

	srv.Router.Route("/v1", func(r chi.Router) {
		walletHandler.RegisterRoutes(r)
		r.Get("/healthz", func(http.ResponseWriter, *http.Request) {})
	})

	err = srv.Start()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	srv.Wait()

	return nil
}
