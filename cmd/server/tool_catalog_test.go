package main

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// everyAbilityTheTradingServiceOffers is this connector's promise written down.
//
// The trading service's own chat assistant is deliberately absent, and this list is
// where that decision is enforced: adding it back turns this test red, which is the
// point. One AI must not be able to spend another AI's budget on its own initiative.
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

// TestNoAbilityLetsOneAssistantSpendAnothersBudget is a guard on a boundary, not a
// spelling check.
//
// The trading service has its own chat assistant, and relaying it would be trivial —
// three more entries in the catalogue. It is left out because an assistant that can
// call an assistant can run up a token bill with nobody in between deciding it was
// worth it, and each of those calls is minutes long and costs real money.
//
// The person can still use that assistant directly. What must not exist is a model's
// ability to reach for it unprompted, so this checks that no such entry has quietly
// come back.
func TestNoAbilityLetsOneAssistantSpendAnothersBudget(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		assert.NotContains(t, apiTool.Name(), "assistant",
			"外掛不代 AI 去問另一個 AI：%s", apiTool.Name())
		assert.NotContains(t, apiTool.Name(), "chat",
			"外掛不代 AI 去問另一個 AI：%s", apiTool.Name())
	}
}

// abilityNamed digs one ability out of the catalogue, so a test about one box does not
// have to walk the whole list in front of the reader.
func abilityNamed(t *testing.T, name string) dto.ToolDefinitionDto {
	t.Helper()

	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		if apiTool.Name() == name {
			return apiTool.ToDefinitionDto()
		}
	}

	require.FailNowf(t, "清單裡沒有這件能力", "%s", name)

	return dto.ToolDefinitionDto{}
}

// boxNamed is one declared box, and whether it was declared at all.
func boxNamed(definitionDto dto.ToolDefinitionDto, name string) (dto.ToolParameterDto, bool) {
	for _, parameter := range definitionDto.Parameters {
		if parameter.Name == name {
			return parameter, true
		}
	}

	return dto.ToolParameterDto{}, false
}

// boxNames is every box an ability declares, in the order it declares them.
func boxNames(definitionDto dto.ToolDefinitionDto) []string {
	names := make([]string, 0, len(definitionDto.Parameters))
	for _, parameter := range definitionDto.Parameters {
		names = append(names, parameter.Name)
	}

	return names
}

// An assistant picking a trading mode cannot see the person's broker and cannot see
// whether the market allows shorting. The sentence they typed is its only clue, so the
// description has to map that sentence onto one of the two spellings — otherwise it
// knows there are two and not which one was just described to it.
func TestWritingATradingStrategySaysWhichKindOfAccountItIsFor(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_trading_strategy",
		"trading_update_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			tradingMode, isDeclared := boxNamed(abilityNamed(t, abilityName), "tradingMode")

			require.True(t, isDeclared, "沒有這個欄位，助理就說不出這份規則是寫給哪一種帳戶的")
			assert.Contains(t, tradingMode.Description, "spot")
			assert.Contains(t, tradingMode.Description, "longShort")
			// What happens when it says nothing, and which one the person just
			// described. Both are things it can only learn here.
			assert.Contains(t, tradingMode.Description, "省略即 longShort")
			assert.Contains(t, tradingMode.Description, "不能放空")
			// Not required: saying nothing is a legitimate thing to do, and the
			// default belongs to the trading service rather than to this list.
			assert.False(t, tradingMode.IsRequired)
		})
	}
}

// Replaying a whole set of rules has no trading mode to give: that set of rules keeps
// its own. A box here would be one the trading service ignores in silence, and an
// assistant has no way to tell it was ignored — it would go on believing it replayed a
// spot account while reading a report card built the other way.
func TestReplayingATradingStrategyHasNoTradingModeToGive(t *testing.T) {
	replay := abilityNamed(t, "trading_backtest_trading_strategy")

	_, isDeclared := boxNamed(replay, "tradingMode")
	assert.False(t, isDeclared, "填了會被交易服務安靜忽略的欄位，比沒有這個欄位更糟")

	// Taking the knob away leaves the assistant knowing only that it does not have
	// one. Where the knob actually is has to be said, or "my account cannot short"
	// gets answered with "I cannot do that".
	assert.Contains(t, replay.Description, "交易模式")
	assert.Contains(t, replay.Description, "trading_update_trading_strategy")
	assert.Contains(t, replay.Description, "不能放空")
}

// Replaying a bare script still asks the caller: there is no trading strategy there to
// ask. Everything else the two replays want is still shared, so the split is one box
// and not a second copy of the replay conditions.
func TestReplayingAStrategyScriptStillAsksWhichWayToTrade(t *testing.T) {
	scriptReplay := abilityNamed(t, "trading_backtest_strategy_script")

	tradingMode, isDeclared := boxNamed(scriptReplay, "tradingMode")
	require.True(t, isDeclared)
	assert.False(t, tradingMode.IsRequired)
	assert.NotEmpty(t, tradingMode.Description)

	strategyReplayBoxes := boxNames(abilityNamed(t, "trading_backtest_trading_strategy"))
	scriptReplayBoxes := boxNames(scriptReplay)

	// Every condition the trading-strategy replay declares, apart from the identifier
	// in its address, the script replay declares too — by the same name, out of the
	// same shared list.
	for _, boxName := range strategyReplayBoxes {
		if boxName == "id" {
			continue
		}

		assert.Contains(t, scriptReplayBoxes, boxName,
			"這一欄兩支回測都要，應該還是共用的那一份：%s", boxName)
	}
}
