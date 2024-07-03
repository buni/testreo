package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buni/wallet/internal/pkg/sloglog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Server ...
type Server struct {
	Context     context.Context
	cancel      func()
	Logger      *slog.Logger
	Router      *chi.Mux
	httpServer  *http.Server
	host        string
	done        chan os.Signal
	gracePeriod time.Duration
	IdleTimeout time.Duration
}

type Option func(*Server) error

func WithGracePeriod(gracePeriod time.Duration) Option {
	return func(s *Server) error {
		s.gracePeriod = gracePeriod
		return nil
	}
}

func WithHost(host string) Option {
	return func(s *Server) error {
		s.host = host
		return nil
	}
}

func NewServer(ctx context.Context, opts ...Option) (server *Server, err error) {
	server = &Server{
		gracePeriod: 10 * time.Second,
		host:        ":8080",
		done:        make(chan os.Signal, 1),
	}

	signal.Notify(server.done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	server.Context, server.cancel = context.WithCancel(ctx)

	zapLogger, err := zap.NewProduction(zap.WithCaller(true), zap.AddStacktrace(zapcore.DPanicLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	server.Logger = slog.New(sloglog.ApplyMiddleware(zapslog.NewHandler(zapLogger.Core(), &zapslog.HandlerOptions{AddSource: true})))

	slog.SetDefault(server.Logger)

	server.Context = sloglog.ToContext(server.Context, server.Logger)

	server.Router = chi.NewRouter()

	server.Router.Use(middleware.Recoverer)
	server.Router.Use(middleware.Logger)

	for _, opt := range opts {
		err = opt(server)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return server, nil
}

func (a *Server) Start() error {
	a.Logger.InfoContext(a.Context, "starting server")
	a.httpServer = &http.Server{
		Addr:        a.host,
		Handler:     h2c.NewHandler(a.Router, &http2.Server{}),
		BaseContext: func(_ net.Listener) context.Context { return a.Context },
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

	<-ctx.Done()
	a.cancel()
	for _, shutdownFunc := range shutdownFuncs {
		shutdownFunc()
	}

	a.Logger.InfoContext(a.Context, "shutdown")
}
