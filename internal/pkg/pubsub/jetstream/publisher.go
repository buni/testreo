package jetstream

import (
	"context"
	"errors"
	"fmt"

	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/nats-io/nats.go"
)

var _ pubsub.Publisher = (*Publisher)(nil)

type Publisher struct {
	jetstreamConn nats.JetStreamContext
}

func NewJetStreamPublisher(jetstreamConn nats.JetStreamContext) (*Publisher, error) {
	return &Publisher{
		jetstreamConn: jetstreamConn,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, msg *pubsub.Message, _ ...pubsub.PublishOption) error {
	_, err := p.jetstreamConn.AddStream(&nats.StreamConfig{
		Name:        msg.Topic,
		Subjects:    []string{msg.Topic + ".*"},
		Storage:     nats.FileStorage,
		AllowDirect: true,
	}, nats.Context(ctx))
	if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	pubOpts := []nats.PubOpt{
		nats.Context(ctx),
	}

	if msg.ID != nil {
		pubOpts = append(pubOpts, nats.MsgId(*msg.ID))
	}

	_, err = p.jetstreamConn.PublishMsg(
		&nats.Msg{
			Reply:   "",
			Sub:     nil,
			Subject: msg.Topic + "." + msg.Key,
			Header:  HeadersToNatsHeaders(msg.Headers),
			Data:    msg.Payload,
		}, pubOpts...)
	if err != nil {
		return fmt.Errorf("failed to publish jetstream msg: %w", err)
	}

	return nil
}

func HeadersToNatsHeaders(headers pubsub.Headers) nats.Header {
	natsHeaders := nats.Header{}
	for k, v := range headers {
		natsHeaders.Add(k, v)
	}
	return natsHeaders
}
