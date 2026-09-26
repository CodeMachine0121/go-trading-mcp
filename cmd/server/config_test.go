package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
)

func noFilledInBoxes() domains.ToolArgumentsDomain {
	return domains.NewToolArgumentsDomain(map[string]json.RawMessage{})
}

func TestTheConnectorRunsWithNoSettingsAtAll(t *testing.T) {
	applicationConfig := loadApplicationConfig()

	assert.Equal(t, "8090", applicationConfig.ServerPort)
	assert.Equal(t, "/mcp", applicationConfig.McpPath)
	assert.Equal(t, "http://localhost:8080", applicationConfig.TradingServiceBaseUrl)
	assert.Equal(t, "http://localhost:8080", applicationConfig.TradingServicePublicUrl)
	assert.Equal(t, "http://localhost:8090/mcp", applicationConfig.ProtectedResourceUrl())
	assert.Equal(t, "http://localhost:8090/.well-known/oauth-protected-resource/mcp",
		applicationConfig.ResourceMetadataUrl())
	assert.Equal(t, 30*time.Second, applicationConfig.TradingServiceRequestTimeout)
	assert.Equal(t, 120*time.Second, applicationConfig.TradingServiceReplayTimeout)
	assert.Equal(t, 10*time.Second, applicationConfig.LiveUpdateWaitLimit)
}

func TestASettingThatWasActuallySaidIsUsed(t *testing.T) {
	t.Setenv("TRADING_SERVICE_BASE_URL", "http://trading.internal:9000")
	t.Setenv("SERVER_PORT", "9999")
	t.Setenv("LIVE_UPDATE_WAIT_LIMIT_SECONDS", "3")
	t.Setenv("TRADING_SERVICE_REPLAY_TIMEOUT_SECONDS", "300")

	applicationConfig := loadApplicationConfig()

	assert.Equal(t, "http://trading.internal:9000", applicationConfig.TradingServiceBaseUrl)
	assert.Equal(t, "9999", applicationConfig.ServerPort)
	assert.Equal(t, 3*time.Second, applicationConfig.LiveUpdateWaitLimit)
	assert.Equal(t, 300*time.Second, applicationConfig.TradingServiceReplayTimeout)
}

func TestATimingThatMakesNoSenseFallsBackRatherThanStoppingTheConnector(t *testing.T) {
	testCases := []struct {
		name     string
		declared string
	}{
		{"不是數字", "十秒"},
		{"零", "0"},
		{"負數", "-5"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("LIVE_UPDATE_WAIT_LIMIT_SECONDS", testCase.declared)

			assert.Equal(t, 10*time.Second, loadApplicationConfig().LiveUpdateWaitLimit,
				"一個打錯的時間不該讓整個外掛起不來")
		})
	}
}

func TestTheConnectorOnlyListensOnLoopbackUnlessTheOperatorSaysOtherwise(t *testing.T) {
	assert.Equal(t, "127.0.0.1", loadApplicationConfig().ServerBindAddress)
}

func TestThePublicAddressesComeFromConfigurationWithoutATrailingSlash(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "https://trading-mcp.example.com/")
	t.Setenv("TRADING_SERVICE_PUBLIC_URL", "https://trading-api.example.com/")
	t.Setenv("MCP_PATH", "/connector")

	applicationConfig := loadApplicationConfig()

	assert.Equal(t, "https://trading-mcp.example.com/connector", applicationConfig.ProtectedResourceUrl())
	assert.Equal(t, "https://trading-mcp.example.com/.well-known/oauth-protected-resource/connector",
		applicationConfig.ResourceMetadataUrl())
	assert.Equal(t, "https://trading-api.example.com", applicationConfig.TradingServicePublicUrl)
}

func TestOpeningItUpIsPossibleButHasToBeSaidOutLoud(t *testing.T) {
	t.Setenv("SERVER_BIND_ADDRESS", "0.0.0.0")

	assert.Equal(t, "0.0.0.0", loadApplicationConfig().ServerBindAddress)
}

func TestAnIdleConnectionIsNotHeldForever(t *testing.T) {
	assert.Equal(t, 60*time.Minute, loadApplicationConfig().IdleConnectionTimeout)

	t.Setenv("IDLE_CONNECTION_TIMEOUT_MINUTES", "15")
	assert.Equal(t, 15*time.Minute, loadApplicationConfig().IdleConnectionTimeout)
}
