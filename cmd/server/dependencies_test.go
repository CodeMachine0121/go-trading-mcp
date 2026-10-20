package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const connectorPublicBaseUrl = "https://trading-mcp.example.com"

func assembledConnector(t *testing.T) *httptest.Server {
	t.Helper()

	tradingServiceStandIn := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/oauth/introspection" {
				writer.WriteHeader(http.StatusNotFound)
				return
			}

			_ = request.ParseForm()
			if request.PostForm.Get("token") != "live" {
				_, _ = writer.Write([]byte(`{"active":false}`))
				return
			}

			_, _ = writer.Write([]byte(`{"active":true,"sub":"42","aud":"` + connectorPublicBaseUrl +
				`/mcp","exp":` + strconv.FormatInt(time.Now().Add(15*time.Minute).Unix(), 10) + `}`))
		}))
	t.Cleanup(tradingServiceStandIn.Close)

	connectorStandIn := httptest.NewServer(buildHttpHandler(ApplicationConfig{
		McpPath:                      "/mcp",
		PublicBaseUrl:                connectorPublicBaseUrl,
		TradingServiceBaseUrl:        tradingServiceStandIn.URL,
		TradingServicePublicUrl:      "https://trading-api.example.com",
		TradingServiceRequestTimeout: time.Second,
		TradingServiceReplayTimeout:  time.Second,
		LiveUpdateWaitLimit:          time.Second,
	}))
	t.Cleanup(connectorStandIn.Close)

	return connectorStandIn
}

type bearerAdding struct{}

func (bearerAdding) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Set("Authorization", "Bearer live")

	return http.DefaultTransport.RoundTrip(request)
}

func assistantSessionOn(t *testing.T, connectorStandIn *httptest.Server) *mcp.ClientSession {
	t.Helper()

	assistantSession, connectError := mcp.NewClient(
		&mcp.Implementation{Name: "assistant", Version: "test"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{
			Endpoint:   connectorStandIn.URL + "/mcp",
			HTTPClient: &http.Client{Transport: bearerAdding{}},
		}, nil)
	require.NoError(t, connectError)
	t.Cleanup(func() { _ = assistantSession.Close() })

	return assistantSession
}

func TestTheAssembledConnectorOffersEveryAbilityOverTheWire(t *testing.T) {
	assistantSession := assistantSessionOn(t, assembledConnector(t))

	listed, listError := assistantSession.ListTools(context.Background(), nil)
	require.NoError(t, listError)

	offeredNames := map[string]bool{}
	for _, tool := range listed.Tools {
		offeredNames[tool.Name] = true
	}

	assert.Equal(t, everyAbilityTheTradingServiceOffers, offeredNames)
	assert.Contains(t, assistantSession.InitializeResult().Instructions, "/mcp")
	assert.NotContains(t, assistantSession.InitializeResult().Instructions, "trading_sign_in")
	assert.Contains(t, assistantSession.InitializeResult().Instructions, "絕不要向使用者要電子郵件或密碼")
}

func TestTheAssembledConnectorTurnsAwayACallWithoutAConnectorAuthorization(t *testing.T) {
	connectorStandIn := assembledConnector(t)

	answer, requestError := http.Post(connectorStandIn.URL+"/mcp", "application/json", nil)
	require.NoError(t, requestError)
	t.Cleanup(func() { _ = answer.Body.Close() })

	assert.Equal(t, http.StatusUnauthorized, answer.StatusCode)
	assert.Equal(t,
		`Bearer resource_metadata="`+connectorPublicBaseUrl+`/.well-known/oauth-protected-resource/mcp"`,
		answer.Header.Get("WWW-Authenticate"))
}

func TestTheAssembledConnectorServesTheSameMetadataAtBothAddresses(t *testing.T) {
	connectorStandIn := assembledConnector(t)

	for _, path := range []string{
		"/.well-known/oauth-protected-resource", "/.well-known/oauth-protected-resource/mcp"} {
		t.Run(path, func(t *testing.T) {
			answer, requestError := http.Get(connectorStandIn.URL + path)
			require.NoError(t, requestError)
			t.Cleanup(func() { _ = answer.Body.Close() })
			body, _ := io.ReadAll(answer.Body)

			assert.Equal(t, http.StatusOK, answer.StatusCode)
			assert.JSONEq(t, `{
				"resource": "`+connectorPublicBaseUrl+`/mcp",
				"authorization_servers": ["https://trading-api.example.com"],
				"bearer_methods_supported": ["header"]}`, string(body))
		})
	}
}

func TestTheConnectorSaysItIsAliveWithoutAnAuthorizationOrTheTradingService(t *testing.T) {
	connectorStandIn := assembledConnector(t)

	answer, requestError := http.Get(connectorStandIn.URL + "/health")
	require.NoError(t, requestError)
	t.Cleanup(func() { _ = answer.Body.Close() })

	assert.Equal(t, http.StatusOK, answer.StatusCode)
}

func TestTheConnectorInstructionsKeepBinanceTradingKeysAndAutoOrderOnTheWeb(t *testing.T) {
	assistantSession := assistantSessionOn(t, assembledConnector(t))

	instructions := assistantSession.InitializeResult().Instructions
	for _, phrase := range []string{
		"trading_get_binance_trading_key_status",
		"只能由使用者自己在網頁上做",
		"不要送出任何改動",
		"絕不要請使用者把 API Key 或 Secret Key 貼進對話**；他主動貼了，也不要複述、不要轉送",
		"會真的在使用者的幣安帳戶上開倉與平倉",
		"讀得到也改得動",
	} {
		assert.Contains(t, instructions, phrase)
	}
	assert.NotContains(t, instructions, "還不會下單")
}

func TestTheConnectorInstructionsSayItDrawsOnlyAssistantLines(t *testing.T) {
	assistantSession := assistantSessionOn(t, assembledConnector(t))

	instructions := assistantSession.InitializeResult().Instructions
	for _, phrase := range []string{
		"在使用者的 K 線圖上畫線",
		"畫線只能動助手畫的",
		"使用者手畫的動不了",
		"優先用吸附",
		"回到網頁的看盤工作區（或重新聚焦那一頁）才看得到",
	} {
		assert.Contains(t, instructions, phrase)
	}
}

func TestTheAssembledConnectorAnswersACallThatCarriesNoSession(t *testing.T) {
	connectorStandIn := assembledConnector(t)
	callWithoutSession := func(body string) *http.Response {
		request, requestError := http.NewRequest(http.MethodPost, connectorStandIn.URL+"/mcp", strings.NewReader(body))
		require.NoError(t, requestError)
		request.Header.Set("Authorization", "Bearer live")
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json, text/event-stream")

		answer, answerError := http.DefaultClient.Do(request)
		require.NoError(t, answerError)
		t.Cleanup(func() { _ = answer.Body.Close() })

		return answer
	}

	initialized := callWithoutSession(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`)
	assert.Equal(t, http.StatusOK, initialized.StatusCode)
	assert.Empty(t, initialized.Header.Get("Mcp-Session-Id"))

	listed := callWithoutSession(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	assert.Equal(t, http.StatusOK, listed.StatusCode)
	listedBody, readError := io.ReadAll(listed.Body)
	require.NoError(t, readError)
	assert.Contains(t, string(listedBody), "trading_get_backtest_run")
}
