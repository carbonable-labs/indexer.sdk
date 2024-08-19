package sdk

import (
	"context"
	"regexp"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
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

	SDK interface {
		Configure(context.Context, Config) (*RegisterResponse, error)
		RegisterHandler(context.Context, string, string, ConsumerHandleFunc) (HandlerCancelFunc, error)
		Start(context.Context) error
	}
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
