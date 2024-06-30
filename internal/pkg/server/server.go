package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	localConfiguration "github.com/buni/cskins/pkg/configuration"
	"github.com/buni/go-libs/chix"
	"github.com/buni/go-libs/configuration"
	"github.com/buni/go-libs/database"
	"github.com/buni/go-libs/database/pgxtx"
	"github.com/buni/go-libs/sloglog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Server ...
type Server struct {
	Context       context.Context
	cancel        func()
	Logger        *slog.Logger
	ZapLogger     *zap.Logger
	TxWrapper     *pgxtx.TxWrapper
	Router        *chi.Mux
	RouterX       *chix.Router
	Config        *localConfiguration.Configuration
	Txm           database.TransactionManager
	NatsConn      *nats.Conn
	JetstreamConn nats.JetStreamContext
	httpServer    *http.Server
	done          chan os.Signal
	gracePeriod   time.Duration
}

// Option ...
type Option func(*Server) error

// WithGracePeriod ...
func WithGracePeriod(gracePeriod time.Duration) Option {
	return func(s *Server) error {
		s.gracePeriod = gracePeriod

		return nil
	}
}

// NewServer ...
func NewServer(ctx context.Context, opts ...Option) (server *Server, err error) {
	server = &Server{
		gracePeriod: 1 * time.Second,
		done:        make(chan os.Signal, 1),
	}

	signal.Notify(server.done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	server.Config, err = configuration.NewConfiguration[localConfiguration.Configuration]()
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	server.Context, server.cancel = context.WithCancel(ctx)

	zapLogger, err := zap.NewProduction(zap.WithCaller(true), zap.AddStacktrace(zapcore.DPanicLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	server.Logger = slog.New(sloglog.ApplyMiddleware(zapslog.NewHandler(zapLogger.Core(), &zapslog.HandlerOptions{AddSource: true})))

	server.ZapLogger = zapLogger

	slog.SetDefault(server.Logger)

	zap.ReplaceGlobals(zapLogger)

	server.Context = sloglog.ToContext(server.Context, server.Logger)

	server.Router = chi.NewRouter()

	server.Router.NotFound(chix.NotFoundHandler)
	server.Router.MethodNotAllowed(chix.MethodNotAllowedHandler)
	server.Router.Use(middleware.Recoverer)
	server.Router.Use(middleware.Logger)
	server.RouterX = chix.NewRouter()

	if server.Config.Database.Host != "" || server.Config.Database.URL != "" { //
		var pgxConf *pgxpool.Config
		server.Config.Database.URL = strings.ReplaceAll(server.Config.Database.URL, "\n", "")
		pgxConf, err = pgxpool.ParseConfig(server.Config.Database.ToURL())
		if err != nil {
			return nil, fmt.Errorf("failed to parse pgx config: %w", err)
		}

		pgxPool, err := pgxpool.NewWithConfig(ctx, pgxConf)
		if err != nil {
			return nil, fmt.Errorf("failed to create pg session: %w", err)
		}

		err = pgxPool.Ping(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to ping pg: %w", err)
		}

		server.TxWrapper = pgxtx.NewTxWrapper(pgxPool, pgx.TxOptions{})

		server.Txm = pgxtx.NewTransactionManager(pgxPool, pgx.TxOptions{})
	}

	if server.Config.NATS.Address != "" || server.Config.NATS.URL != "" {
		natsOpts := []nats.Option{}
		if server.Config.NATS.JWT != "" && server.Config.NATS.Seed != "" {
			natsOpts = append(natsOpts, nats.UserJWTAndSeed(server.Config.NATS.JWT, server.Config.NATS.Seed), nats.Secure(&tls.Config{ //nolint:gosec
				InsecureSkipVerify: true, //nolint:gosec
			}))
		}

		server.NatsConn, err = nats.Connect(server.Config.NATS.ToURL(), natsOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to nats: %w", err)
		}
		server.JetstreamConn, err = server.NatsConn.JetStream()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jetstream: %w", err)
		}
	}

	for _, opt := range opts {
		err = opt(server)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return server, nil
}

// Start ...
func (a *Server) Start() error {
	a.Logger.InfoContext(a.Context, "starting server")
	a.httpServer = &http.Server{
		Addr:              a.Config.Service.ToHost(),
		Handler:           h2c.NewHandler(a.Router, &http2.Server{}),
		ReadHeaderTimeout: 120 * time.Second,
		IdleTimeout:       120 * time.Second,
		BaseContext:       func(_ net.Listener) context.Context { return a.Context },
	}

	go func() {
		err := a.httpServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			a.Logger.ErrorContext(a.Context, "http server error", sloglog.Error(err))
		}
	}()

	return nil
}

// Shutdown ...
func (a *Server) Wait(shutdownFuncs ...func()) {
	<-a.done
	ctx, cancel := context.WithTimeout(a.Context, a.gracePeriod)
	defer cancel()

	a.Logger.InfoContext(a.Context, "shutting down")
	a.httpServer.Shutdown(ctx) //nolint:errcheck,revive
	for _, shutdownFunc := range shutdownFuncs {
		shutdownFunc()
	}

	a.cancel()
	<-ctx.Done()
	a.Logger.InfoContext(a.Context, "shutdown")
}
