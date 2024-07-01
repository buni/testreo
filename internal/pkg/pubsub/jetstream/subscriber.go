package jetstream

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/buni/wallet/internal/pkg/sloglog"
	"github.com/nats-io/nats.go"
)

const (
	groupVersionPrefix = "v1pubsub"
)

var ErrInvalidTopicName = errors.New("invalid topic name")

type Subscriber struct {
	jetstreamConn nats.JetStreamContext
	wg            sync.WaitGroup
}

func NewJetstreamSubscriber(jetstreamConn nats.JetStreamContext) *Subscriber {
	return &Subscriber{
		jetstreamConn: jetstreamConn,
		wg:            sync.WaitGroup{},
	}
}

func (s *Subscriber) Subscribe(ctx context.Context, topic string, _ ...pubsub.SubscriberOption) (<-chan pubsub.SubscriberMessage, error) {
	logger := sloglog.FromContext(ctx)
	msgs := make(chan pubsub.SubscriberMessage, 1)
	group := strings.ReplaceAll(fmt.Sprintf("%s%s", groupVersionPrefix, topic), ".", "-")

	streamName := strings.Split(topic, ".")
	if len(streamName) == 0 {
		return nil, ErrInvalidTopicName
	}

	_, err := s.jetstreamConn.AddConsumer(streamName[0], &nats.ConsumerConfig{ //nolint:exhaustruct
		Durable:       group,
		FilterSubject: topic,
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverLastPolicy,
	}, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	sub, err := s.jetstreamConn.PullSubscribe(topic, group, nats.AckExplicit(), nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create pull subscriber: %w", err)
	}

	s.wg.Add(1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.wg.Done()
				return
			default:
				msg, err := sub.Fetch(1, nats.Context(ctx))
				if err != nil || len(msg) == 0 {
					logger.ErrorContext(ctx, "error fetching jetstream msgs ", sloglog.Error(err))
					continue
				}
				msgs <- &jetstreamSubscriberMessage{
					rw:        &sync.RWMutex{},
					msg:       msg[0],
					processed: false,
				}
			}
		}
	}()

	return msgs, nil
}

type jetstreamSubscriberMessage struct {
	rw        *sync.RWMutex
	msg       *nats.Msg
	processed bool
}

func HeaderToHeader(natsHeader nats.Header) pubsub.Headers {
	header := make(pubsub.Headers)
	for k, v := range natsHeader {
		header[k] = v[0]
	}
	return header
}

func (m *jetstreamSubscriberMessage) Message(_ context.Context) (*pubsub.Message, error) {
	return &pubsub.Message{
		ID:      new(string),
		Key:     m.msg.Subject,
		Topic:   m.msg.Subject,
		Payload: m.msg.Data,
		Headers: HeaderToHeader(m.msg.Header),
	}, nil
}

func (m *jetstreamSubscriberMessage) IsProcessed(_ context.Context) (bool, error) {
	m.rw.RLock()
	defer m.rw.RUnlock()
	return m.processed, nil
}

func (m *jetstreamSubscriberMessage) Ack(ctx context.Context) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	err := m.msg.AckSync(nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to ack msg: %w", err)
	}

	m.processed = true
	return nil
}

func (m *jetstreamSubscriberMessage) Nack(ctx context.Context) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	err := m.msg.Nak(nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to release msg: %w", err)
	}

	m.processed = true
	return nil
}

func (m *jetstreamSubscriberMessage) DLQ(ctx context.Context) error {
	m.rw.Lock()
	defer m.rw.Unlock()

	err := m.msg.Term(nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to DLQ msg: %w", err)
	}

	m.processed = true

	return nil
}
