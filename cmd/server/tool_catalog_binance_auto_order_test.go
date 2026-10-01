package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var everyAbilityThatAnswersWithAStrategyBot = []string{
	"trading_create_strategy_bot",
	"trading_list_strategy_bots",
	"trading_get_strategy_bot",
	"trading_update_strategy_bot",
	"trading_start_strategy_bot",
	"trading_stop_strategy_bot",
	"trading_run_strategy_bot_now",
}

func TestEveryAbilityThatAnswersWithAStrategyBotSaysAutoOrderDoesNotPlaceOrdersYet(t *testing.T) {
	for _, abilityName := range everyAbilityThatAnswersWithAStrategyBot {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			assert.Contains(t, description, "autoOrderEnabled")
			assert.Contains(t, description, "目前開著也還不會下單")
			assert.Contains(t, description, "只送 Telegram 通知")
		})
	}
}

func TestReadingTheBinanceTradingKeyStatusAsksTheStatusViewOnly(t *testing.T) {
	ability := apiToolNamed(t, "trading_get_binance_trading_key_status")

	request, buildError := ability.BuildRequest(domains.NewToolArgumentsDomain(nil))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbRead, request.Verb)
	assert.Equal(t, "/users/me/binance-trading-key/status", request.Path)
	assert.Empty(t, ability.ToDefinitionDto().Parameters)
}

func TestReadingTheBinanceTradingKeyStatusSaysHowToReportEachAnswer(t *testing.T) {
	description := abilityNamed(t, "trading_get_binance_trading_key_status").Description

	for _, phrase := range []string{
		"configured", "tradableMarkets", "spot 現貨", "contract 合約",
		"不含任何一段金鑰內容", "連 API Key 的結尾也沒有",
		"configured 為 false 時只說「尚未設定」", "網頁的設定頁",
		"不要把它說成日期",
	} {
		assert.Contains(t, description, phrase)
	}
}

func TestNoAbilityWritesABinanceTradingKeyOrSwitchesAutoOrder(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10*time.Second, 120*time.Second) {
		request, _ := apiTool.BuildRequest(domains.NewToolArgumentsDomain(nil))

		assert.NotContains(t, request.Path, "/auto-order", apiTool.Name())
		if strings.HasPrefix(request.Path, "/users/me/binance-trading-key") {
			assert.Equal(t, vo.RequestVerbRead, request.Verb, apiTool.Name())
			assert.Equal(t, "/users/me/binance-trading-key/status", request.Path, apiTool.Name())
		}
	}
}

func TestWritingAStrategyBotNeverForwardsAutoOrder(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_bot", "trading_update_strategy_bot"} {
		t.Run(abilityName, func(t *testing.T) {
			_, isDeclared := boxNamed(abilityNamed(t, abilityName), "autoOrderEnabled")
			assert.False(t, isDeclared)

			request, buildError := apiToolNamed(t, abilityName).BuildRequest(
				aBotWriteForm(map[string]json.RawMessage{"autoOrderEnabled": json.RawMessage(`true`)}))

			require.NoError(t, buildError)
			assert.NotContains(t, string(request.Body), "autoOrderEnabled")
			assert.Contains(t, string(request.Body), `"triggerIntervalMinutes":15`)
		})
	}
}

func TestWritingAStrategyBotSaysAutoOrderIsSwitchedOnlyOnTheWeb(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_bot", "trading_update_strategy_bot"} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			assert.Contains(t, description, "這個外掛都做不到")
			assert.Contains(t, description, "機器人的詳細頁")
			assert.Contains(t, description, "絕不要請使用者把 API Key 或 Secret Key 貼進對話")
			assert.Contains(t, description, "不要複述、不要轉送")
		})
	}

	assert.Contains(t, abilityNamed(t, "trading_create_strategy_bot").Description, "新建的機器人自動下單一律是關的")
	assert.Contains(t, abilityNamed(t, "trading_update_strategy_bot").Description, "改完維持原本的開或關")
}
