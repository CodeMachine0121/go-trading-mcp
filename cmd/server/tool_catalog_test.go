package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// everyAbilityTheTradingServiceOffers is this connector's promise written down.
//
// It is a literal list rather than something derived from the catalogue, which is the
// only way it can catch anything: a list read out of the thing it is checking agrees
// with it by construction. When the trading service gains an endpoint, this list gains
// a line and goes red until the catalogue gains one too — which is exactly the moment
// somebody should be told.
var everyAbilityTheTradingServiceOffers = map[string]bool{
	// 系統
	"trading_health": false,
	// 帳號（登入、登出、續用是外掛自己的，不在這份轉達清單裡）
	"trading_register_user":    false,
	"trading_get_current_user": true,
	"trading_change_password":  true,
	// 行情
	"trading_create_k_candle":           false,
	"trading_list_k_candles":            false,
	"trading_get_k_candle_series":       false,
	"trading_get_k_candle":              false,
	"trading_update_k_candle":           false,
	"trading_delete_k_candle":           false,
	"trading_backfill_k_candles":        false,
	"trading_sync_k_candle_history":     false,
	"trading_get_k_candle_history_sync": false,
	"trading_peek_live_k_candle":        false,
	// 交易標的
	"trading_list_trading_symbols":  false,
	"trading_add_to_watchlist":      false,
	"trading_remove_from_watchlist": false,
	// 指標
	"trading_calculate_indicator": true,
	// 策略腳本與市集
	"trading_create_strategy_script":   true,
	"trading_list_strategy_scripts":    true,
	"trading_get_strategy_script":      true,
	"trading_update_strategy_script":   true,
	"trading_delete_strategy_script":   true,
	"trading_publish_strategy_script":  true,
	"trading_withdraw_strategy_script": true,
	"trading_browse_marketplace":       true,
	"trading_adopt_strategy_script":    true,
	"trading_abandon_strategy_script":  true,
	// 交易策略
	"trading_create_trading_strategy": true,
	"trading_list_trading_strategies": true,
	"trading_get_trading_strategy":    true,
	"trading_update_trading_strategy": true,
	"trading_delete_trading_strategy": true,
	// 重演
	"trading_backtest_strategy_script":  true,
	"trading_backtest_trading_strategy": true,
	// 策略機器人
	"trading_create_strategy_bot":    true,
	"trading_list_strategy_bots":     true,
	"trading_get_strategy_bot":       true,
	"trading_update_strategy_bot":    true,
	"trading_delete_strategy_bot":    true,
	"trading_start_strategy_bot":     true,
	"trading_stop_strategy_bot":      true,
	"trading_list_strategy_bot_runs": true,
	"trading_run_strategy_bot_now":   true,
	// 通知
	"trading_get_telegram_delivery":      true,
	"trading_save_telegram_delivery":     true,
	"trading_remove_telegram_delivery":   true,
	"trading_send_telegram_test_message": true,
	// 行情對話助手
	"trading_ask_assistant":                true,
	"trading_list_assistant_conversations": true,
	"trading_get_assistant_conversation":   true,
}

func TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse(t *testing.T) {
	requiresSignInPerNames := map[string]bool{}
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		requiresSignInPerNames[apiTool.Name()] = apiTool.RequiresSignIn()
	}

	assert.Equal(t, everyAbilityTheTradingServiceOffers, requiresSignInPerNames)
}

func TestNoTwoAbilitiesShareAName(t *testing.T) {
	catalog := apiToolCatalog(10 * time.Second)

	seenNames := map[string]bool{}
	for _, apiTool := range catalog {
		require.False(t, seenNames[apiTool.Name()],
			"兩件事同名時，後放的會安靜地蓋掉先放的：%s", apiTool.Name())
		seenNames[apiTool.Name()] = true
	}
}

func TestEveryAbilityTellsTheAssistantWhatItIsForAndWhatEachBoxMeans(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		definitionDto := apiTool.ToDefinitionDto()

		assert.NotEmpty(t, definitionDto.Description,
			"沒有說明的能力，助理只能靠名字猜：%s", definitionDto.Name)

		for _, parameter := range definitionDto.Parameters {
			assert.NotEmpty(t, parameter.Description,
				"沒有說明的欄位，助理會一直填錯：%s 的 %s", definitionDto.Name, parameter.Name)
			assert.NotEmpty(t, parameter.Kind,
				"沒有型別的欄位，助理送什麼都可能：%s 的 %s", definitionDto.Name, parameter.Name)
		}
	}
}

func TestOnlyWatchingAnAbilityStaysOnTheLine(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		request, buildError := apiTool.BuildRequest(noFilledInBoxes())
		if buildError != nil {
			continue
		}

		if apiTool.Name() == "trading_peek_live_k_candle" {
			assert.Equal(t, 10*time.Second, request.LiveUpdateWaitLimit)
			continue
		}

		assert.Zero(t, request.LiveUpdateWaitLimit,
			"只有「看一眼即時更新」該停在線上：%s", apiTool.Name())
	}
}
