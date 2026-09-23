package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A contract script replay goes to the contract replay with the contract conditions
// the assistant filled in, and only those.
func TestTheContractScriptReplayForwardsTheContractConditions(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_backtest_contract_strategy_script").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"strategyScriptId":    json.RawMessage(`21`),
			"symbol":              json.RawMessage(`"BTCUSDT"`),
			"aggregationInterval": json.RawMessage(`"1h"`),
			"startTime":           json.RawMessage(`"2026-08-01T00:00:00Z"`),
			"endTime":             json.RawMessage(`"2026-09-01T00:00:00Z"`),
			"initialCapital":      json.RawMessage(`"10000"`),
			"leverage":            json.RawMessage(`"5"`),
			"tradingMode":         json.RawMessage(`"shortOnly"`),
			"slippagePercentage":  json.RawMessage(`"0.05"`),
		}))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbSubmit, request.Verb)
	assert.Equal(t, "/contract-backtests", request.Path)
	assert.True(t, request.CarriesIdentity)
	body := string(request.Body)
	for _, sent := range []string{
		`"leverage":"5"`, `"tradingMode":"shortOnly"`, `"slippagePercentage":"0.05"`,
		`"symbol":"BTCUSDT"`, `"strategyScriptId":21`, `"aggregationInterval":"1h"`,
	} {
		assert.Contains(t, body, sent)
	}
}

// What the assistant left blank never leaves as a value: the trading service's own
// defaults decide — one times, long and short, no slippage.
func TestTheContractScriptReplaySendsNothingForWhatWasLeftBlank(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_backtest_contract_strategy_script").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"script":         json.RawMessage(`"package main"`),
			"symbol":         json.RawMessage(`"BTCUSDT"`),
			"startTime":      json.RawMessage(`"2026-08-01T00:00:00Z"`),
			"endTime":        json.RawMessage(`"2026-09-01T00:00:00Z"`),
			"initialCapital": json.RawMessage(`"10000"`),
		}))

	require.NoError(t, buildError)
	for _, blank := range []string{"leverage", "tradingMode", "slippagePercentage"} {
		assert.NotContains(t, string(request.Body), blank)
	}
}

// A replay without its symbol is held back before it is sent, naming the box.
func TestTheContractReplaysHoldBackAMissingSymbol(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_contract_strategy_script",
		"trading_backtest_contract_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			_, buildError := apiToolNamed(t, abilityName).
				BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
					"id":             json.RawMessage(`"11"`),
					"startTime":      json.RawMessage(`"2026-08-01T00:00:00Z"`),
					"endTime":        json.RawMessage(`"2026-09-01T00:00:00Z"`),
					"initialCapital": json.RawMessage(`"10000"`),
				}))

			require.ErrorIs(t, buildError, domains.ErrRequiredArgumentMissing)
			assert.ErrorContains(t, buildError, "symbol")
		})
	}
}

// A contract trading strategy replay goes under the trading strategy it replays, and
// has no trading mode to give: the trading strategy says its own.
func TestTheContractTradingStrategyReplayTakesNoTradingMode(t *testing.T) {
	ability := abilityNamed(t, "trading_backtest_contract_trading_strategy")
	_, isDeclared := boxNamed(ability, "tradingMode")
	assert.False(t, isDeclared)
	assert.Contains(t, ability.Description, "都取自那份交易策略本身")
	assert.Contains(t, ability.Description, "沒有 tradingMode 那一格")
	assert.Contains(t, ability.Description, "trading_update_trading_strategy")

	request, buildError := apiToolNamed(t, "trading_backtest_contract_trading_strategy").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"id":             json.RawMessage(`"11"`),
			"symbol":         json.RawMessage(`"BTCUSDT"`),
			"startTime":      json.RawMessage(`"2026-08-01T00:00:00Z"`),
			"endTime":        json.RawMessage(`"2026-09-01T00:00:00Z"`),
			"initialCapital": json.RawMessage(`"10000"`),
			"leverage":       json.RawMessage(`"3"`),
			"tradingMode":    json.RawMessage(`"shortOnly"`),
		}))

	require.NoError(t, buildError)
	assert.Equal(t, "/trading-strategies/11/contract-backtests", request.Path)
	assert.Contains(t, string(request.Body), `"leverage":"3"`)
	// A mode the assistant sends anyway is dropped rather than forwarded to be refused.
	assert.NotContains(t, string(request.Body), "tradingMode")
}

// The script replay's trading mode names all three spellings and says spot is not one.
func TestTheContractScriptReplayOffersTheThreeTradingModes(t *testing.T) {
	tradingMode, isDeclared := boxNamed(
		abilityNamed(t, "trading_backtest_contract_strategy_script"), "tradingMode")
	require.True(t, isDeclared)

	for _, spelling := range []string{"longShort", "longOnly", "shortOnly", "spot"} {
		assert.Contains(t, tradingMode.Description, spelling)
	}
}

// Both contract replays read the account rules out of one list, and that list is the
// spot one extended — word for word — with the leverage and the slippage.
func TestBothContractReplaysShareOneListExtendingTheSpotOne(t *testing.T) {
	scriptReplay := abilityNamed(t, "trading_backtest_contract_strategy_script")
	strategyReplay := abilityNamed(t, "trading_backtest_contract_trading_strategy")
	spotReplay := abilityNamed(t, "trading_backtest_trading_strategy")

	for _, strategyReplayBox := range strategyReplay.Parameters {
		if strategyReplayBox.Name == "id" {
			continue
		}
		scriptReplayBox, isDeclared := boxNamed(scriptReplay, strategyReplayBox.Name)
		require.Truef(t, isDeclared, "兩支合約重演都要這一格：%s", strategyReplayBox.Name)
		assert.Equal(t, strategyReplayBox, scriptReplayBox)
	}

	for _, spotBox := range spotReplay.Parameters {
		if spotBox.Name == "id" {
			continue
		}
		contractBox, isDeclared := boxNamed(strategyReplay, spotBox.Name)
		require.Truef(t, isDeclared, "合約重演少了現貨那一格：%s", spotBox.Name)
		assert.Equal(t, spotBox, contractBox)
	}

	for _, contractOnlyBox := range []string{"leverage", "slippagePercentage"} {
		box, isDeclared := boxNamed(strategyReplay, contractOnlyBox)
		require.Truef(t, isDeclared, "合約重演要有這一格：%s", contractOnlyBox)
		assert.NotEmpty(t, box.Description)
	}
}

// Both contract replays say how a contract account is replayed, word for word, and say
// the figures an assistant would otherwise misread.
func TestBothContractReplaysSayHowTheContractAccountIsReplayed(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_contract_strategy_script",
		"trading_backtest_contract_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			assert.Contains(t, abilityNamed(t, abilityName).Description, contractAccountReplayNote)
		})
	}

	for _, phrase := range []string{
		"逐倉", "強制平倉看標記價格", "資金費率一律計入", "強平價會往進場價靠近", "離進場價近的先到",
		"liquidationExitCount", "totalFundingFee", "shortWinRate", "blockedOpeningCount", "maintenanceMarginBasis",
		"還沒有交易規格", "maintenanceMarginRate", "多空反手一旦進場就一直在場內",
	} {
		assert.Contains(t, contractAccountReplayNote, phrase)
	}
}

// The spot replays still take none of the contract boxes.
func TestTheSpotReplaysTakeNoneOfTheContractBoxes(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			_, isDeclared := boxNamed(abilityNamed(t, abilityName), "slippagePercentage")
			assert.False(t, isDeclared)
		})
	}
}

// Writing a trading strategy says which market it eats and, for a contract one, its
// trading mode — and forwards only what was said.
func TestWritingATradingStrategySaysItsKindAndTradingMode(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_trading_strategy",
		"trading_update_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)

			marketDataKind, isDeclared := boxNamed(ability, "marketDataKind")
			require.True(t, isDeclared)
			for _, phrase := range []string{"kCandle", "contractKCandle", "建立後不得更換", "修改時不給就是保留原本的", "都要吃同一種行情"} {
				assert.Contains(t, marketDataKind.Description, phrase)
			}

			tradingMode, isDeclared := boxNamed(ability, "tradingMode")
			require.True(t, isDeclared)
			for _, phrase := range []string{"只有吃合約行情的交易策略才有", "longShort", "longOnly", "shortOnly", "吃 K 線的交易策略沒有交易模式"} {
				assert.Contains(t, tradingMode.Description, phrase)
			}
		})
	}

	request, buildError := apiToolNamed(t, "trading_update_trading_strategy").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"id":            json.RawMessage(`"11"`),
			"name":          json.RawMessage(`"合約黃金交叉"`),
			"signalSources": json.RawMessage(`[]`),
			"buyCondition":  json.RawMessage(`{}`),
			"sellCondition": json.RawMessage(`{}`),
			"tradingMode":   json.RawMessage(`"longOnly"`),
		}))
	require.NoError(t, buildError)
	assert.Contains(t, string(request.Body), `"tradingMode":"longOnly"`)
	assert.NotContains(t, string(request.Body), "marketDataKind")
}

// A bot says it only follows K candle trading strategies, before the assistant tries.
func TestABotSaysItOnlyFollowsKCandleTradingStrategies(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_bot", "trading_update_strategy_bot"} {
		t.Run(abilityName, func(t *testing.T) {
			tradingStrategyID, isDeclared := boxNamed(abilityNamed(t, abilityName), "tradingStrategyId")
			require.True(t, isDeclared)
			assert.Contains(t, tradingStrategyID.Description, "只能是吃 K 線的交易策略")
		})
	}
}

// Each contract replay says which kind of script or trading strategy gets it refused,
// and where that one goes instead.
func TestTheContractReplaysSayWhatTheyRefuseAndWhereItGoes(t *testing.T) {
	scriptReplay := abilityNamed(t, "trading_backtest_contract_strategy_script").Description
	assert.Contains(t, scriptReplay, "吃 K 線的會被拒絕，那一支要用 trading_backtest_strategy_script")
	assert.Contains(t, scriptReplay, "沒有 spot 這一種")

	strategyReplay := abilityNamed(t, "trading_backtest_contract_trading_strategy").Description
	assert.Contains(t, strategyReplay, "吃 K 線的交易策略會被拒絕，那一份要用 trading_backtest_trading_strategy")
}
