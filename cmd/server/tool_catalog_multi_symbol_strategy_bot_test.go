package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWritingAStrategyBotAsksForARequiredListOfSymbols(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)

			symbols, isDeclared := boxNamed(ability, "symbols")
			require.True(t, isDeclared)
			assert.True(t, symbols.IsRequired)
			assert.Equal(t, string(vo.ToolParameterKindArray), symbols.Kind)
			for _, phrase := range []string{
				"1 到 10 個、不得重複", "大小寫不同視為同一個", "超過 10 個都會被拒絕",
				"每一個標的都必須已經在合約追蹤名單上", "不得超過**任何一個**標的的上限",
				"修改時給的是修改後的完整清單", "新加的從空白開始", "沒列進去的就是要移除",
				"過去的執行紀錄仍保留", "漏列一個就是把它移除",
			} {
				assert.Contains(t, symbols.Description, phrase)
			}

			_, isSingleSymbolDeclared := boxNamed(ability, "symbol")
			assert.False(t, isSingleSymbolDeclared)
		})
	}
}

func TestWritingAStrategyBotForwardsTheSymbolListAsGiven(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		for _, givenSymbols := range []string{
			`["BTCUSDT","ETHUSDT"]`,
			`["BTCUSDT"]`,
			`["A1USDT","A2USDT","A3USDT","A4USDT","A5USDT","A6USDT","A7USDT","A8USDT","A9USDT","A10USDT"]`,
			`["A1USDT","A2USDT","A3USDT","A4USDT","A5USDT","A6USDT","A7USDT","A8USDT","A9USDT","A10USDT","A11USDT"]`,
			`["BTCUSDT","btcusdt"]`,
		} {
			t.Run(abilityName+" "+givenSymbols, func(t *testing.T) {
				request, buildError := apiToolNamed(t, abilityName).BuildRequest(
					aBotWriteForm(map[string]json.RawMessage{"symbols": json.RawMessage(givenSymbols)}))

				require.NoError(t, buildError)
				assert.Contains(t, string(request.Body), `"symbols":`+givenSymbols)
			})
		}
	}
}

// The single-symbol box is gone, so an assistant still filling it in sends nothing the trading service would refuse.
func TestWritingAStrategyBotNeverForwardsASingleSymbol(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, abilityName).BuildRequest(
				aBotWriteForm(map[string]json.RawMessage{"symbol": json.RawMessage(`"SOLUSDT"`)}))

			require.NoError(t, buildError)
			assert.NotContains(t, string(request.Body), `"symbol"`)
			assert.NotContains(t, string(request.Body), "SOLUSDT")
		})
	}
}

func TestWritingAStrategyBotWithoutSymbolsIsNotSent(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			_, buildError := apiToolNamed(t, abilityName).BuildRequest(
				domains.NewToolArgumentsDomain(map[string]json.RawMessage{
					"id":                     json.RawMessage(`"7"`),
					"name":                   json.RawMessage(`"BTC 永續五倍"`),
					"tradingStrategyId":      json.RawMessage(`3`),
					"triggerIntervalMinutes": json.RawMessage(`15`),
				}))

			assert.Error(t, buildError)
		})
	}
}

func TestCreatingAStrategyBotSaysItWatchesUpToTenSymbols(t *testing.T) {
	description := abilityNamed(t, "trading_create_strategy_bot").Description

	assert.Contains(t, description, "在 1 到 10 個交易標的（監聽標的）上定時自己跑")
	assert.NotContains(t, description, "在一個交易標的上")
}

func TestRewritingAStrategyBotSaysTheListIsCompleteAndWhyARemovalIsRefused(t *testing.T) {
	description := abilityNamed(t, "trading_update_strategy_bot").Description

	for _, phrase := range []string{
		"**symbols 是修改後的完整清單**", "也要把原本的標的照抄回去",
		"整次修改什麼都不改",
		"watchedSymbolHoldsPosition", "還有機器人持倉，要先讓它空手才能移除",
		"watchedSymbolAutoOrderOutstanding", "還有沒做完的自動下單，要等它做完才能移除",
		"**不要重送同一份**",
	} {
		assert.Contains(t, description, phrase)
	}
}

func TestEveryAbilityAnsweringWithABotSaysHowToReadEachWatchedSymbol(t *testing.T) {
	for _, abilityName := range everyAbilityThatAnswersWithAStrategyBot {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"watchedSymbols", "lastSentSignal 上一次送出的信號（沒送過就沒有）", "conflicting", "autoOrderPosition",
				"機器人本身**不再**有單一的 symbol",
				"**整台停擺**", "**標的停擺**", "**其他標的照常、機器人仍在跑**",
				"**兩種停擺看它寫在哪裡分，不看它的值**", "也可能是 scriptFailed 或 orderFlowRefused",
				"scriptFailed", "orderFlowRefused",
				"只盯一個標的時，它停擺就是整台停擺",
				"**停止再啟動**會清掉每個標的的停擺、上一次送出的信號與衝突記號，機器人持倉不動",
				"每一個標的都跑（含停擺中的），跑出結論的標的就不再停擺",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestReadingABotsRoundsCanAskForOneSymbol(t *testing.T) {
	ability := abilityNamed(t, "trading_list_strategy_bot_runs")

	symbol, isDeclared := boxNamed(ability, "symbol")
	require.True(t, isDeclared)
	assert.False(t, symbol.IsRequired)
	for _, phrase := range []string{"大小寫不拘", "不給就是全部標的", "格式不對會被拒絕"} {
		assert.Contains(t, symbol.Description, phrase)
	}

	narrowedRequest, narrowedError := apiToolNamed(t, "trading_list_strategy_bot_runs").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"id":     json.RawMessage(`"3"`),
			"symbol": json.RawMessage(`"ETHUSDT"`),
		}))
	require.NoError(t, narrowedError)
	assert.Equal(t, map[string]string{"symbol": "ETHUSDT"}, narrowedRequest.Query)

	everyRequest, everyError := apiToolNamed(t, "trading_list_strategy_bot_runs").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{"id": json.RawMessage(`"3"`)}))
	require.NoError(t, everyError)
	assert.Empty(t, everyRequest.Query)
}

func TestReadingABotsRoundsSaysEachRecordNamesItsSymbol(t *testing.T) {
	description := abilityNamed(t, "trading_list_strategy_bot_runs").Description

	for _, phrase := range []string{
		"**每一筆紀錄帶著 symbol",
		"輪次編號（runNumber）在同一台機器人的各標的之間不重複",
		"每個標的各記得最近 50 輪", "被移除的標的過去的紀錄仍在",
	} {
		assert.Contains(t, description, phrase)
	}
}

func TestASkippedContractRoundIsOnlyThatSymbols(t *testing.T) {
	for _, abilityName := range []string{"trading_list_strategy_bot_runs", "trading_run_strategy_bot_now"} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			assert.Contains(t, description, "**只有那個標的**那一輪會直接跳過")
			assert.Contains(t, description, "其他標的照常")
		})
	}
}

func TestRemovingAContractFromTheWatchlistBlindsOnlyThatWatchedSymbol(t *testing.T) {
	description := abilityNamed(t, "trading_remove_from_contract_watchlist").Description

	assert.Contains(t, description, "同一台機器人的其他監聽標的照常")
	assert.Contains(t, description, "watchedSymbols 裡有它")
}
