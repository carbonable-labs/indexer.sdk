package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/stretchr/testify/assert"
)

var testConfigs = Config{
	Appname: "test_config",
	Contracts: []Contract{
		{
			Name:    "project",
			Address: "0x0516d0acb6341dcc567e85dc90c8f64e0c33d3daba0a310157d6bba0656c8769",
			Events: map[string]string{
				"URI": "project:uri",
			},
		},
		{
			Name:    "project_karathuru",
			Address: "0x05a667adc04676fba78a29371561a0bf91dab25847d5dc4709a93a4cfb5ff293",
			Events: map[string]string{
				"URI": "project:uri",
			},
		},
		{
			Name:    "yielder_banegas_farm",
			Address: "0x03d25473be5a6316f351e8f964d0c303357c006f7107779f648d9879b7c6d58a",
			Events: map[string]string{
				"Deposit": "yielder:deposit",
			},
		},
		{
			Name:    "yielder_las_delicias",
			Address: "0x00426d4e86913759bcc49b7f992b1fe62e6571e8f8089c23d95fea815dbad471",
			Events: map[string]string{
				"Deposit": "yielder:deposit",
			},
		},
		{
			Name:    "yielder_manjarisoa",
			Address: "0x03afe61732ed9b226309775ac4705129319729d3bee81da5632146ffd72652ae",
			Events: map[string]string{
				"Deposit": "yielder:deposit",
			},
		},
	},
	StartBlock: 0,
}

func TestConfigFilterByName(t *testing.T) {
	c := testConfigs.FilterByName("project")
	assert.Equal(t, 2, len(c.Contracts))
	c = testConfigs.FilterByName("yielder")
	assert.Equal(t, 3, len(c.Contracts))
	c = testConfigs.FilterByName("banegas_farm")
	assert.Equal(t, 1, len(c.Contracts))
}

func TestConfigCall(t *testing.T) {
	client, err := rpc.NewProvider("https://free-rpc.nethermind.io/mainnet-juno")
	if err != nil {
		t.Error("failed to dial in rpc")
	}

	c := testConfigs.FilterByName("project").Contracts[0]
	resp, err := c.Call(context.Background(), client, "slot_count")

	assert.Equal(t, uint64(3), resp[0].Uint64())
	assert.Equal(t, nil, err)

	resp, err = c.Call(context.Background(), client, "not_a_function")
	assert.Equal(t, 0, len(resp))
	assert.Equal(t, &rpc.RPCError{Code: -32603, Message: "Internal Error", Data: "Contract error"}, err)
}

func TestConfigure(t *testing.T) {
	err := WithToken("test_token")
	assert.Equal(t, nil, err)
	assert.Equal(t, "test_token", indexerToken)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{ "app_name": "test_config", "hash": "test_hash" }`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	_ = WithApi(server.URL)
	conf, err := Configure(testConfigs)

	assert.Equal(t, nil, err)
	assert.Equal(t, "test_config", conf.AppName)
	assert.Equal(t, "test_hash", conf.Hash)
}
