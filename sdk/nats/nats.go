package nats

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/carbonable-labs/indexer.sdk/sdk"
)

type (
	NatsSDK struct {
		opts NatsSDKOpts
	}
	NatsSDKOpts struct {
		indexerToken  string
		indexerUrl    string
		indexerApi    string
		indexerApiKey string
	}
	NatsSDKOptsFn func(NatsSDKOpts) NatsSDKOpts
)

// Creates a new NatsSDK instance
// Package entrypoint
func NewSDK(o ...NatsSDKOptsFn) *NatsSDK {
	opts := defaultNatsOpts()
	for _, optFn := range o {
		opts = optFn(opts)
	}

	return &NatsSDK{opts: opts}
}

// Configure method  
// Given a input config, it will register the app in the indexer and return the hash of the app
func (s *NatsSDK) Configure(ctx context.Context, c sdk.Config) (*sdk.RegisterResponse, error) {
	slog.Debug("configure", "app_name", c.AppName)

	client := http.DefaultClient

	body, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", s.opts.indexerApi+"/register", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var r sdk.RegisterResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &r, nil
}

// RegisterHandler method  
// Based on the config you sent to configuration it will wait messages
// from queue and use the callback you provide to integrate messages into your system
func (s *NatsSDK) RegisterHandler(ctx context.Context, name string, subject string, cb sdk.ConsumerHandleFunc) (sdk.HandlerCancelFunc, error) {
	slog.Debug("register handler", "app_name", name)

	nc, err := nats.Connect(s.opts.indexerUrl, nats.Token(s.opts.indexerToken))
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
		Name:          name,
		Durable:       name,
		FilterSubject: subject,
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

		var e sdk.RawEvent
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
		err := js.DeleteConsumer(context.Background(), "EVENTS", name)
		if err != nil {
			if !errors.Is(jetstream.ErrConsumerNotFound, err) {
				slog.Error("failed to delete consumer", "error", err)
			}
		}
		cctx.Stop()
	}, nil
}

func (s *NatsSDK) Start(ctx context.Context) error {
	return nil
}

// NatsSDKOpts default builder function
// creates options with default values overridable with env
func defaultNatsOpts() NatsSDKOpts {
	return NatsSDKOpts{
		indexerToken:  os.Getenv("INDEXER_TOKEN"),
		indexerUrl:    os.Getenv("INDEXER_URL"),
		indexerApi:    os.Getenv("INDEXER_API"),
		indexerApiKey: os.Getenv("INDEXER_API_KEY"),
	}
}

func WithToken(t string) NatsSDKOptsFn {
	return func(opts NatsSDKOpts) NatsSDKOpts {
		opts.indexerToken = t
		return opts
	}
}

func WithUrl(u string) NatsSDKOptsFn {
	return func(opts NatsSDKOpts) NatsSDKOpts {
		opts.indexerUrl = u
		return opts
	}
}

func WithApi(a string) NatsSDKOptsFn {
	return func(opts NatsSDKOpts) NatsSDKOpts {
		opts.indexerApi = a
		return opts
	}
}

func WithApiKey(k string) NatsSDKOptsFn {
	return func(opts NatsSDKOpts) NatsSDKOpts {
		opts.indexerApiKey = k
		return opts
	}
}
