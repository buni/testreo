package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/buni/wallet/internal/pkg/database"
	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/buni/wallet/internal/pkg/sloglog"
)

type PublisherSettings struct {
	Publisher     pubsub.Publisher
	PublisherType string
	OptionsStruct func() any
}

type Outbox struct {
	repo       Repository
	txm        database.TransactionManager
	publishers []PublisherSettings
	pollSize   uint64
	ticker     *time.Ticker
	logger     *slog.Logger
	wg         *sync.WaitGroup
}

type Option func(*Outbox) error

func NewOutboxWorker(
	repo Repository,
	txm database.TransactionManager,
	publishers []PublisherSettings,
	opts ...Option,
) (*Outbox, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	o := &Outbox{
		repo:       repo,
		txm:        txm,
		publishers: publishers,
		pollSize:   10,
		logger:     logger,
		ticker:     time.NewTicker(1 * time.Second),
		wg:         &sync.WaitGroup{},
	}

	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return o, nil
}

func (o *Outbox) Start(ctx context.Context) (err error) {
	for _, v := range o.publishers {
		o.wg.Add(1)
		go func(pubSettings PublisherSettings) {
			for {
				select {
				case <-ctx.Done():
					o.ticker.Stop()
					o.wg.Done()
					return
				case <-o.ticker.C:
					err := o.pollMessages(ctx, pubSettings.PublisherType, pubSettings.Publisher, pubSettings.OptionsStruct)
					if err != nil {
						o.logger.Error("failed to poll messages", sloglog.Error(err))
					}
				}
			}
		}(v)
	}

	return nil
}

func (o *Outbox) pollMessages(ctx context.Context, publisherType string, publisher pubsub.Publisher, _ func() any) error {
	logger := o.logger.With(slog.String("publisher_type", publisherType))

	err := o.txm.Run(ctx, func(ctx context.Context) error {
		messages, err := o.repo.List(ctx, o.pollSize, "queued", publisherType, time.Now().UTC().Truncate(time.Microsecond))
		if err != nil {
			return fmt.Errorf("failed to list messages: %w", err)
		}

		for _, msg := range messages { //nolint:gocritic
			msg := msg

			err = publisher.Publish(ctx, &msg.Payload)
			if err != nil {
				logger.Error("failed to publish message", slog.Any("err", err), slog.Any("message", msg))
				continue
			}

			msg.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

			msg.Status = "published"

			err = o.repo.Update(ctx, msg)
			if err != nil {
				logger.Error("failed to update message", slog.Any("err", err), slog.Any("message", msg))
				continue
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to run transaction: %w", err)
	}

	return nil
}

func (o *Outbox) Wait() {
	o.wg.Wait()
}
