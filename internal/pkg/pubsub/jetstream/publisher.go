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
	natsConn      *nats.Conn
	jetstreamConn nats.JetStreamContext
}

func NewJetStreamPublisher(natsConn *nats.Conn, jetstreamConn nats.JetStreamContext) (*Publisher, error) {
	return &Publisher{
		natsConn:      natsConn,
		jetstreamConn: jetstreamConn,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, msg *pubsub.Message, _ ...pubsub.PublishOption) error { // TODO: add traces
	_, err := p.jetstreamConn.AddStream(&nats.StreamConfig{ //nolint:exhaustruct
		// FIXME: add all fields
		Name:     msg.Topic,
		Subjects: []string{msg.Topic + ".*"},
		Storage:  nats.FileStorage,
		// Replicas:    3,
		AllowDirect: true,
		// MirrorDirect: true,
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
