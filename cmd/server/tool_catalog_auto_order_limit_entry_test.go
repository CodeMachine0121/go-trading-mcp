package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const entrySlippageToleranceBox = "autoOrderEntrySlippageTolerancePercentage"

var bothStrategyBotWrites = []string{"trading_create_strategy_bot", "trading_update_strategy_bot"}

func TestWritingAStrategyBotOffersAnOptionalEntrySlippageTolerance(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			box, isDeclared := boxNamed(abilityNamed(t, abilityName), entrySlippageToleranceBox)

			require.True(t, isDeclared)
			assert.False(t, box.IsRequired)
			assert.Equal(t, string(vo.ToolParameterKindString), box.Kind)
		})
	}
}

func TestAGivenEntrySlippageToleranceTravelsAsGiven(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		for _, givenTolerance := range []string{`"0.05"`, `"0"`} {
			t.Run(abilityName+" "+givenTolerance, func(t *testing.T) {
				request, buildError := apiToolNamed(t, abilityName).BuildRequest(
					aBotWriteForm(map[string]json.RawMessage{entrySlippageToleranceBox: json.RawMessage(givenTolerance)}))

				require.NoError(t, buildError)
				assert.Contains(t, string(request.Body), `"`+entrySlippageToleranceBox+`":`+givenTolerance)
			})
		}
	}
}

func TestAnUngivenEntrySlippageToleranceIsNotSentSoTheTradingServiceDefaultsOrKeepsIt(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, abilityName).BuildRequest(aBotWriteForm(nil))

			require.NoError(t, buildError)
			assert.NotContains(t, string(request.Body), entrySlippageToleranceBox)
		})
	}
}

func TestTheEntrySlippageToleranceSaysWhatItMeansBeforeAnythingIsSent(t *testing.T) {
	box, _ := boxNamed(abilityNamed(t, "trading_create_strategy_bot"), entrySlippageToleranceBox)

	for _, phrase := range []string{
		"百分點", "介於 0 到 5（含）", "建立時不給就是 0.1", "修改時不給就是保留原本的值",
		"參考價 ×（1 ＋ 容忍度 ÷ 100）", "參考價 ×（1 − 容忍度 ÷ 100）",
		"最多等一個觸發間隔", "一點都沒成交就撤單",
		"0 代表只肯在參考價成交，快速行情裡可能常常不成交",
		"平倉一律立刻以市場價成交，不受它影響",
		"現貨機器人目前不下單，所以沒有作用",
		"下一輪說出新結論、使用者關掉自動下單或停止機器人，進場單就立刻撤掉",
		"部分成交的那一份在進場單收尾前還沒有止損",
		"只有已停止的機器人改得動",
	} {
		assert.Contains(t, box.Description, phrase)
	}
}

func TestRewritingAStrategyBotSaysTheEntrySlippageToleranceIsKeptWhenNotGiven(t *testing.T) {
	description := abilityNamed(t, "trading_update_strategy_bot").Description

	assert.Contains(t, description, "這是整份改寫")
	assert.Contains(t, description, "只有已停止的機器人改得動**：在跑的會被拒絕，先用 trading_stop_strategy_bot 停掉")
	assert.Contains(t, description, "autoOrderEntrySlippageTolerancePercentage（進場滑價容忍度，不給不會變成 0，也不會回到 0.1）")
}

func TestReadingABotsRoundsSaysHowEachAutoOrderEntryWent(t *testing.T) {
	for _, abilityName := range []string{"trading_list_strategy_bot_runs", "trading_run_strategy_bot_now"} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"autoOrder（下單結果）",
				"pending 等待下單", "filled 全部完成", "partiallyDone 部分完成",
				"unfilled **沒有成交**（進場單送出了卻一點都沒成交、已撤單，沒有開倉也沒有平倉——不是失敗）",
				"notPlaced 沒有下單", "abandoned 送出前就放棄",
				"referencePrice 參考價", "entryLimitPrice 進場限價",
				"entrySlippageTolerancePercentage 當時用的容忍度",
				"requestedQuantity 要求數量", "entryExecutedQuantity 已成交數量",
				"entrySlippagePercentage 相對參考價的滑價百分比——**正數表示對使用者不利**",
				"working 等待成交（最多等一個觸發間隔）", "partiallyFilled 部分成交（其餘已撤單",
				"unfilled 沒有成交（已撤單）",
				"這個功能上線前的紀錄也沒有",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestDeletingABotSaysWhatToDoWhileItsEntryOrderStillRests(t *testing.T) {
	description := abilityNamed(t, "trading_delete_strategy_bot").Description

	for _, phrase := range []string{
		"autoOrderEntryOutstanding",
		"自己到網頁上這台機器人的詳細頁關掉自動下單",
		"等進場單撤掉",
		"這個外掛不能替他關，不要重送同一個刪除",
		"絕不要請使用者把 API Key 或 Secret Key 貼進對話",
	} {
		assert.Contains(t, description, phrase)
	}
}
