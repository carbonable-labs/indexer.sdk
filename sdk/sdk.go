package sdk

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
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
	// Indexer configuration
	Config struct {
		AppName    string     `json:"app_name"`
		Contracts  []Contract `json:"contracts"`
		StartBlock uint64     `json:"start_block"`
	}
	Contract struct {
		Events  map[string]string `json:"events"`
		Name    string            `json:"name"`
		Address string            `json:"address"`
	}

	// RegisterResponse
	RegisterResponse struct {
		AppName string `json:"app_name"`
		Hash    string `json:"hash"`
	}
)

var (
	indexerToken  = os.Getenv("INDEXER_TOKEN")
	indexerUrl    = os.Getenv("INDEXER_URL")
	indexerApi    = os.Getenv("INDEXER_API")
	indexerApiKey = os.Getenv("INDEXER_API_KEY")
)

// FilterByName method  
// Filters config contracts by name. Useful when you want to select a specific contract in your config
func (c Config) FilterByName(name string) Config {
	var contracts []Contract
	for _, contract := range c.Contracts {
		m, _ := regexp.MatchString(name, contract.Name)
		if m {
			contracts = append(contracts, contract)
		}
	}
	c.Contracts = contracts
	return c
}

// Call method  
// Execute `call` function on a contract of your config.
func (c Contract) Call(ctx context.Context, client rpc.RpcProvider, fn string, calldata ...*felt.Felt) ([]*felt.Felt, error) {
	addr, err := utils.HexToFelt(c.Address)
	if err != nil {
		return nil, err
	}

	tx := rpc.FunctionCall{
		ContractAddress:    addr,
		EntryPointSelector: utils.GetSelectorFromNameFelt(fn),
		Calldata:           calldata,
	}
	callResp, rpcErr := client.Call(ctx, tx, rpc.BlockID{Tag: "latest"})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return callResp, nil
}

// Configure indexer token
// Either use this methof or set the token in the env variable INDEXER_TOKEN
func WithToken(t string) error {
	indexerToken = t
	return nil
}

// Configure indexer url
// Either use this methof or set the url in the env variable INDEXER_URL
func WithUrl(u string) error {
	indexerUrl = u
	return nil
}

// Configure indexer api
// Either use this methof or set the api in the env variable INDEXER_API
func WithApi(a string) error {
	indexerApi = a
	return nil
}

// Configure indexer api key
// Either use this methof or set the api key in the env variable INDEXER_API_KEY
func WithApiKey(k string) error {
	indexerApiKey = k
	return nil
}

// Configure method  
// Given a input config, it will register the app in the indexer and return the hash of the app
func Configure(c Config) (*RegisterResponse, error) {
	slog.Debug("configure", "app_name", c.AppName)

	client := http.DefaultClient

	body, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", indexerApi+"/register", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var r RegisterResponse
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
func RegisterHandler(n string, s string, cb ConsumerHandleFunc) (HandlerCancelFunc, error) {
	slog.Debug("register handler", "app_name", n)

	nc, err := nats.Connect(indexerUrl, nats.Token(indexerToken))
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
