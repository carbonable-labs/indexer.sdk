package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/carbonable-labs/indexer.sdk/sdk"
	"github.com/stretchr/testify/assert"
)

var fakeConfigs = sdk.Config{
	AppName: "test_config",
	Contracts: []sdk.Contract{
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

func TestNewSDKDefaultValues(t *testing.T) {
	os.Setenv("INDEXER_TOKEN", "thisisasuperprivateroken")
	os.Setenv("INDEXER_URL", "http://localhost:9999")

	sdk := NewSDK()

	assert.Equal(t, "thisisasuperprivateroken", sdk.opts.indexerToken)
	assert.Equal(t, "http://localhost:9999", sdk.opts.indexerUrl)
	assert.Equal(t, "", sdk.opts.indexerApi)
	assert.Equal(t, "", sdk.opts.indexerApiKey)
}

func TestSDKOverride(t *testing.T) {
	os.Setenv("INDEXER_TOKEN", "thisisasuperprivateroken")
	os.Setenv("INDEXER_URL", "http://localhost:8888")
	os.Setenv("INDEXER_API", "http://localhost:6000")
	os.Setenv("INDEXER_API_KEY", "privateapikey")

	sdk := NewSDK(WithToken("notsoprivatetoken"), WithUrl("http://localhost:9999"), WithApi("http://localhost:3000"), WithApiKey("apikey"))

	assert.Equal(t, "notsoprivatetoken", sdk.opts.indexerToken)
	assert.Equal(t, "http://localhost:9999", sdk.opts.indexerUrl)
	assert.Equal(t, "http://localhost:3000", sdk.opts.indexerApi)
	assert.Equal(t, "apikey", sdk.opts.indexerApiKey)
}

func TestConfigure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" {
			t.Errorf("expected /register, got %s", r.URL.Path)
		}
		var c sdk.Config
		_ = json.NewDecoder(r.Body).Decode(&c)
		defer r.Body.Close()
		response := fmt.Sprintf(`{ "app_name": "%s", "hash": "test_hash" }`, c.AppName)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	s := NewSDK(WithApi(server.URL))
	conf, err := s.Configure(context.Background(), fakeConfigs)

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
	var c sdk.Config
	err := json.Unmarshal([]byte(conf), &c)
	if err != nil {
		t.Errorf("failed to parse config")
	}

	s := NewSDK(WithApi("https://carbonable-event-indexer-sepolia.fly.dev"))
	res, err := s.Configure(context.Background(), c)
	fmt.Println(res, err)
}
