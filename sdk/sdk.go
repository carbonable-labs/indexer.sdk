package sdk

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	HandlerCancelFunc    func()
	ConsumerHandleFunc   func(string, uint64, RawEvent) error
	MiddlewareHandleFunc func(ConsumerHandleFunc) ConsumerHandleFunc
	// Event directly pulled onchain
	RawEvent struct {
		RecordedAt  time.Time `json:"recorded_at"`
		EventId     string    `json:"event_id"`
		FromAddress string    `json:"from_address"`
		Keys        []string  `json:"keys"`
		Data        []string  `json:"data"`
	}
)

func RegisterHandler(n string, s string, cb ConsumerHandleFunc) (HandlerCancelFunc, error) {
	slog.Debug("register handler", "app_name", n)

	nc, err := nats.Connect("carbonable-nats-sepolia.fly.dev", nats.Token("ibLtRZVRLFVNDZa9ZabGZVWuAxxq3d"))
	if err != nil {
		slog.Error("failed to connect to nats", "error", err)
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		slog.Error("failed to create jetstream", "error", err)
		return nil, err
	}

	c, err := js.CreateOrUpdateConsumer(context.Background(), "EVENTS", jetstream.ConsumerConfig{
		Name:          n,
		Durable:       n,
		FilterSubject: s,
	})
	if err != nil {
		slog.Error("failed to create or update consumer", "error", err)
		return nil, err
	}
	cctx, err := c.Consume(func(msg jetstream.Msg) {
		// NOTE: Here is the piece of software to send messages to consumers.
		// we can send the message plus some metadata to it
		// eg: msg.Data(), sequenceId,
		subject := msg.Subject()
		meta, _ := msg.Metadata()

		slog.Debug("received message", "subject", subject, "sequence", meta.Sequence.Stream)

		var e RawEvent
		decoder := gob.NewDecoder(bytes.NewReader(msg.Data()))
		if err := decoder.Decode(&e); err != nil {
			slog.Error("failed to decode raw event", "error", err)
			return
		}

		err = cb(msg.Subject(), meta.Sequence.Stream, e)
		if err != nil {
			slog.Error("failed to consume message", "error", err)
			return
		}

		_ = msg.Ack()
	})
	if err != nil {
		slog.Error("failed to consume stream", "error", err)
		return nil, err
	}

	return func() {
		err := js.DeleteConsumer(context.Background(), "EVENTS", n)
		if err != nil {
			if !errors.Is(jetstream.ErrConsumerNotFound, err) {
				slog.Error("failed to delete consumer", "error", err)
			}
		}
		cctx.Stop()
	}, nil
}
