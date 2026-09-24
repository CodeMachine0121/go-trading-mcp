package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aBotWriteForm is the least a bot can be written with, plus whatever this case adds.
func aBotWriteForm(extraBoxes map[string]json.RawMessage) domains.ToolArgumentsDomain {
	boxes := map[string]json.RawMessage{
		"id":                     json.RawMessage(`"7"`),
		"name":                   json.RawMessage(`"BTC 永續五倍"`),
		"symbol":                 json.RawMessage(`"BTCUSDT"`),
		"tradingStrategyId":      json.RawMessage(`3`),
		"triggerIntervalMinutes": json.RawMessage(`15`),
	}
	for name, value := range extraBoxes {
		boxes[name] = value
	}

	return domains.NewToolArgumentsDomain(boxes)
}

// Writing a bot says which kind it is — and which trading strategies and symbols each
// kind takes — before the assistant is refused for mixing them.
func TestWritingAStrategyBotSaysWhichKindItIs(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_bot", "trading_update_strategy_bot"} {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)

			marketDataKind, isDeclared := boxNamed(ability, "marketDataKind")
			require.True(t, isDeclared, "沒有這一格，助理建不出合約機器人")
			// Optional: leaving it out is a spot bot, exactly as before.
			assert.False(t, marketDataKind.IsRequired)
			for _, phrase := range []string{
				"kCandle", "contractKCandle", "現貨機器人", "合約機器人",
				"建立時不給就是 kCandle", "修改時不給就是保留原本的", "建立後不得更換",
			} {
				assert.Contains(t, marketDataKind.Description, phrase)
			}

			tradingStrategyID, _ := boxNamed(ability, "tradingStrategyId")
			assert.Contains(t, tradingStrategyID.Description, "現貨機器人只能引用吃 K 線的交易策略")
			assert.Contains(t, tradingStrategyID.Description, "合約機器人只能引用吃合約行情的交易策略")
			// The spot-only sentence would send the assistant away from a bot that works.
			assert.NotContains(t, tradingStrategyID.Description, "只能是吃 K 線的交易策略")
			assert.NotContains(t, tradingStrategyID.Description, "目前只跑 K 線")

			symbol, _ := boxNamed(ability, "symbol")
			assert.Contains(t, symbol.Description, "必須已經在合約追蹤名單上")
			assert.Contains(t, symbol.Description, "trading_add_to_contract_watchlist")
		})
	}
}

// What the assistant says about the kind is forwarded; what it leaves out is not
// filled in for it — a contract bot's rename must not arrive as a request to be spot.
func TestWritingAStrategyBotForwardsTheKindOnlyWhenGiven(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_bot", "trading_update_strategy_bot"} {
		t.Run(abilityName+" given", func(t *testing.T) {
			request, buildError := apiToolNamed(t, abilityName).BuildRequest(
				aBotWriteForm(map[string]json.RawMessage{
					"marketDataKind": json.RawMessage(`"contractKCandle"`),
				}))

			require.NoError(t, buildError)
			assert.Contains(t, string(request.Body), `"marketDataKind":"contractKCandle"`)
		})

		t.Run(abilityName+" left out", func(t *testing.T) {
			request, buildError := apiToolNamed(t, abilityName).BuildRequest(aBotWriteForm(nil))

			require.NoError(t, buildError)
			assert.NotContains(t, string(request.Body), "marketDataKind")
		})
	}
}

// The rewrite says the kind is the one box a whole rewrite does not clear.
func TestRewritingAStrategyBotSaysTheKindIsKept(t *testing.T) {
	description := abilityNamed(t, "trading_update_strategy_bot").Description

	assert.Contains(t, description, "例外是 marketDataKind**：不給就是保留原本的種類")
	assert.Contains(t, description, "換成另一種會被拒絕")
}

// A contract bot's leverage travels inside the plan exactly as given, and a plan that
// gives none sends none.
func TestAContractBotsLeverageTravelsInsideItsPlan(t *testing.T) {
	givenRequest, givenError := apiToolNamed(t, "trading_create_strategy_bot").BuildRequest(
		aBotWriteForm(map[string]json.RawMessage{
			"marketDataKind": json.RawMessage(`"contractKCandle"`),
			"positionPlan":   json.RawMessage(`{"capital":"1000","leverage":"5"}`),
		}))
	require.NoError(t, givenError)
	assert.Contains(t, string(givenRequest.Body), `"positionPlan":{"capital":"1000","leverage":"5"}`)

	leftOutRequest, leftOutError := apiToolNamed(t, "trading_create_strategy_bot").BuildRequest(
		aBotWriteForm(map[string]json.RawMessage{
			"positionPlan": json.RawMessage(`{"capital":"1000"}`),
		}))
	require.NoError(t, leftOutError)
	assert.NotContains(t, string(leftOutRequest.Body), "leverage")
}

// Listing can ask for one kind only, and asks for nothing when no kind is given.
func TestListingStrategyBotsCanAskForOneKind(t *testing.T) {
	ability := abilityNamed(t, "trading_list_strategy_bots")
	assert.Contains(t, ability.Description, "每一台都帶著 marketDataKind")

	marketDataKind, isDeclared := boxNamed(ability, "marketDataKind")
	require.True(t, isDeclared)
	assert.False(t, marketDataKind.IsRequired)
	assert.Contains(t, marketDataKind.Description, "只列其中一種")
	assert.Contains(t, marketDataKind.Description, "不給就全部列出")

	narrowedRequest, narrowedError := apiToolNamed(t, "trading_list_strategy_bots").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"marketDataKind": json.RawMessage(`"contractKCandle"`),
		}))
	require.NoError(t, narrowedError)
	assert.Equal(t, map[string]string{"marketDataKind": "contractKCandle"}, narrowedRequest.Query)
	assert.Empty(t, narrowedRequest.Body)

	everyRequest, everyError := apiToolNamed(t, "trading_list_strategy_bots").BuildRequest(
		domains.NewToolArgumentsDomain(nil))
	require.NoError(t, everyError)
	assert.Empty(t, everyRequest.Query)
}

// Reading one bot says it carries its kind, and a contract bot its leverage.
func TestReadingAStrategyBotSaysItCarriesItsKind(t *testing.T) {
	description := abilityNamed(t, "trading_get_strategy_bot").Description

	assert.Contains(t, description, "marketDataKind")
	assert.Contains(t, description, "合約機器人的 positionPlan 另帶著它的槓桿倍數")
}

// A contract trading strategy is no longer described as something no bot can follow.
func TestCreatingATradingStrategySaysAContractOneFitsAContractBot(t *testing.T) {
	description := abilityNamed(t, "trading_create_trading_strategy").Description

	assert.Contains(t, description, "可以掛上**合約機器人**")
	assert.NotContains(t, description, "策略機器人目前掛不上它")
}

// Reading a bot's rounds says that a contract bot's run of holds may be rounds it skipped
// because its contract's candles stopped arriving — not a bot with no signal yet.
func TestReadingABotsRoundsSaysAContractBotsHoldsMayBeSkippedRounds(t *testing.T) {
	for _, abilityName := range []string{"trading_list_strategy_bot_runs", "trading_run_strategy_bot_now"} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"合約機器人一連串的 hold 不一定是沒有信號", "早超過 5 分鐘",
				"trading_list_contract_k_candles", "trading_add_to_contract_watchlist",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

// Taking a contract off the watchlist blinds every contract bot watching it, so the
// ability says so before the assistant tidies a watchlist a live bot depends on.
func TestRemovingAContractFromTheWatchlistWarnsOfTheBotsWatchingIt(t *testing.T) {
	description := abilityNamed(t, "trading_remove_from_contract_watchlist").Description

	for _, phrase := range []string{"合約機器人會就此失明", "每一輪都會跳過", "trading_list_strategy_bots", "contractKCandle"} {
		assert.Contains(t, description, phrase)
	}
}

// Reconciling a contract bot's stops points to the replay that takes contract rules.
func TestCreatingABotNamesTheReplayEachKindReconcilesWith(t *testing.T) {
	description := abilityNamed(t, "trading_create_strategy_bot").Description

	assert.Contains(t, description, "trading_backtest_trading_strategy（合約機器人用 trading_backtest_contract_trading_strategy）")
}
