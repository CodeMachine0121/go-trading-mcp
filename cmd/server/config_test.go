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
	assert.Equal(t, 30*time.Second, applicationConfig.TradingServiceRequestTimeout)
	assert.Equal(t, 10*time.Second, applicationConfig.LiveUpdateWaitLimit)
}

func TestASettingThatWasActuallySaidIsUsed(t *testing.T) {
	t.Setenv("TRADING_SERVICE_BASE_URL", "http://trading.internal:9000")
	t.Setenv("SERVER_PORT", "9999")
	t.Setenv("LIVE_UPDATE_WAIT_LIMIT_SECONDS", "3")

	applicationConfig := loadApplicationConfig()

	assert.Equal(t, "http://trading.internal:9000", applicationConfig.TradingServiceBaseUrl)
	assert.Equal(t, "9999", applicationConfig.ServerPort)
	assert.Equal(t, 3*time.Second, applicationConfig.LiveUpdateWaitLimit)
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
	assert.Equal(t, "127.0.0.1", loadApplicationConfig().ServerBindAddress,
		"這個端點沒有門鎖，所以聽在每一張網卡上必須是個明確的動作，不是預設值")
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
