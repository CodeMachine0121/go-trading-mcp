package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheAssembledConnectorOffersEveryAbilityOverTheWire is the one test that proves
// the wiring itself, rather than any one piece of it.
//
// Everything below is already covered by its own tests; what only this can catch is a
// controller that was built but never registered, or a catalogue that was built and
// handed to nobody — both of which compile perfectly and do nothing.
func TestTheAssembledConnectorOffersEveryAbilityOverTheWire(t *testing.T) {
	applicationConfig := ApplicationConfig{
		McpPath:                      "/mcp",
		TradingServiceBaseUrl:        "http://127.0.0.1:1",
		TradingServiceRequestTimeout: time.Second,
		LiveUpdateWaitLimit:          time.Second,
	}

	connectorStandIn := httptest.NewServer(
		buildHttpHandler(applicationConfig, buildMcpServer(applicationConfig)))
	t.Cleanup(connectorStandIn.Close)

	assistantSession, connectError := mcp.NewClient(
		&mcp.Implementation{Name: "assistant", Version: "test"}, nil).
		Connect(context.Background(),
			&mcp.StreamableClientTransport{Endpoint: connectorStandIn.URL + "/mcp"}, nil)
	require.NoError(t, connectError)
	t.Cleanup(func() { _ = assistantSession.Close() })

	listed, listError := assistantSession.ListTools(context.Background(), nil)
	require.NoError(t, listError)

	offeredNames := map[string]bool{}
	for _, tool := range listed.Tools {
		offeredNames[tool.Name] = true
	}

	for relayedName := range everyAbilityTheTradingServiceOffers {
		assert.True(t, offeredNames[relayedName], "這件事沒有被掛上去：%s", relayedName)
	}

	for _, connectorOwnedName := range []string{
		"trading_sign_in", "trading_sign_out", "trading_renew_session"} {
		assert.True(t, offeredNames[connectorOwnedName],
			"這件事沒有被掛上去：%s", connectorOwnedName)
	}

	assert.Len(t, offeredNames, len(everyAbilityTheTradingServiceOffers)+3)
}

func TestTheConnectorSaysItIsAliveWithoutTouchingTheTradingService(t *testing.T) {
	applicationConfig := ApplicationConfig{
		McpPath:                      "/mcp",
		TradingServiceBaseUrl:        "http://127.0.0.1:1",
		TradingServiceRequestTimeout: time.Second,
		LiveUpdateWaitLimit:          time.Second,
	}

	connectorStandIn := httptest.NewServer(
		buildHttpHandler(applicationConfig, buildMcpServer(applicationConfig)))
	t.Cleanup(connectorStandIn.Close)

	answer, requestError := http.Get(connectorStandIn.URL + "/health")
	require.NoError(t, requestError)
	t.Cleanup(func() { _ = answer.Body.Close() })

	assert.Equal(t, http.StatusOK, answer.StatusCode)
}
