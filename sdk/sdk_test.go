package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/stretchr/testify/assert"
)

var testConfigs = Config{
	AppName: "test_config",
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
		if r.URL.Path != "/register" {
			t.Errorf("expected /register, got %s", r.URL.Path)
		}
		var c Config
		_ = json.NewDecoder(r.Body).Decode(&c)
		defer r.Body.Close()
		response := fmt.Sprintf(`{ "app_name": "%s", "hash": "test_hash" }`, c.AppName)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	_ = WithApi(server.URL)
	conf, err := Configure(testConfigs)

	assert.Equal(t, nil, err)
	assert.Equal(t, "test_config", conf.AppName)
	assert.Equal(t, "test_hash", conf.Hash)
}

func TestConfigureReal(t *testing.T) {
	conf := `{
  "app_name": "carbonable_portfolio_backend",
  "start_block": 0,
  "contracts": [
    {
      "name": "project",
      "address": "0x00130b5a3035eef0470cff2f9a450a7a6856a3c5a4ea3f5b7886c2d03a50d2bf"
    },
    {
      "name": "minter_banegas_farm",
      "address": "0x2cf1693df4529343fed040fcefe33a50611aa93dd9c399e4baef0f08a82b99d"
    },
    {
      "name": "yielder_banegas_farm",
      "address": "0x00f6019754ab54ea7d806720d17b425c799db5ebb337e4b2d8c1ed71fc35f342"
    },
    {
      "name": "offseter_banegas_farm",
      "address": "0x008637332b17f5ffe7f21f076389e8a5461f25fbc0049ac0243b4e08591280df"
    },
    {
      "name": "minter_las_delicias",
      "address": "0x04c9c5303f0c0f40cdfd5f5631052288185e37abe3af54de9c37610b423b1b25"
    },
    {
      "name": "yielder_las_delicias",
      "address": "0x0370e85e8f315dc352eeef7e7f0f5d70e89c699384cbcb81a11a7089fa87ff66"
    },
    {
      "name": "offseter_las_delicias",
      "address": "0x04f634a74451bc19e4d537326dff7552c225040e9d9c16b26a32466eebdf9688"
    }
  ]
	}`
	var c Config
	err := json.Unmarshal([]byte(conf), &c)
	if err != nil {
		t.Errorf("failed to parse config")
	}

	_ = WithApi("https://carbonable-event-indexer-sepolia.fly.dev")

	res, err := Configure(c)
	fmt.Println(res, err)
}
