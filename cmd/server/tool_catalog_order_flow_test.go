package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Each kind is pinned next to what it is, so swapping two definitions fails rather than passes.
func TestEveryParameterBoxNamesTheFourKindsTheSameWay(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_strategy_script", "trading_update_strategy_script",
		"trading_calculate_indicator", "trading_calculate_contract_indicator",
		"trading_backtest_strategy_script", "trading_backtest_contract_strategy_script",
	} {
		box, hasBox := boxNamed(abilityNamed(t, abilityName), "parameters")
		require.True(t, hasBox, abilityName)

		for _, phrase := range []string{
			"lookbackCount（大於零的整數，這條線要回看幾根",
			"number（任何數字，原樣交給算式）",
			"boolean（是非，0 為否、非零為是）",
			"orderFlowPriceStep（**訂單流價位寬度**：大於零、可以有小數、最細 0.00000001",
			"**只給吃合約行情（contractKCandle）的算式、一支至多一個**，吃 K 線的宣告會被拒絕",
			"訂單流價位寬度的值就在每一格的 Footprint.PriceStep，也可以用 indicator.Number(\"名字\") 讀",
		} {
			assert.Contains(t, box.Description, phrase, abilityName)
		}
		assert.NotContains(t, box.Description, "只有兩種", abilityName)
	}
}

// Only a run takes this time's values, so only a run's box talks about them.
func TestOnlyARunsParameterBoxTalksAboutThisTimesValues(t *testing.T) {
	for abilityName, takesValues := range map[string]bool{
		"trading_create_strategy_script": false, "trading_update_strategy_script": false,
		"trading_calculate_indicator": true, "trading_calculate_contract_indicator": true,
		"trading_backtest_strategy_script": true, "trading_backtest_contract_strategy_script": true,
	} {
		box, _ := boxNamed(abilityNamed(t, abilityName), "parameters")

		assert.Equal(t, takesValues,
			strings.Contains(box.Description, "這一次的 parameterValues 同樣照這些規則檢查——例如把 orderFlowPriceStep 調成 0 會整次被拒絕"),
			abilityName)
	}
}

func TestTheContractScriptNoteNamesTheFootprint(t *testing.T) {
	for _, phrase := range []string{
		"Footprint（**訂單流足跡**，indicator.Footprint：PriceStep 價位寬度、Levels 一串 indicator.PriceLevel",
		"Price、BuyVolume（買方主動吃單）、SellVolume（賣方主動吃單）、TradeCount，價位由低到高",
		"Delta 買賣差、PointOfControl 成交最多的價位，平手取低",
		"**訂單流足跡要先宣告一個 kind 為 orderFlowPriceStep 的參數才有**，沒宣告時整份是空的",
		// Remembered in memory only, so a repeat can still be slow.
		"訂單流**不存進資料庫**",
		"重啟、或請求落在另一台分身上時會重拿",
	} {
		assert.Contains(t, contractKCandleScriptNote, phrase)
	}
	assert.NotContains(t, contractKCandleScriptNote, "每次現拿")
}

// The limit travels with every ability where it bites, said the same way.
func TestTheOrderFlowReachIsExplainedWhereverItCanRefuse(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_strategy_script", "trading_calculate_contract_indicator",
		"trading_backtest_contract_strategy_script", "trading_backtest_contract_trading_strategy",
		"trading_create_strategy_bot", "trading_update_strategy_bot",
	} {
		description := abilityNamed(t, abilityName).Description

		for _, phrase := range []string{
			"一次至多拿 31 個世界標準時間日曆日的訂單流",
			"**回看根數與重演的暖機都算在內**",
			"跨過午夜就多算一天",
			"重演在**送出時**就照整段判",
			"合約機器人每一輪要的是「回看根數 × 彙總刻度」那麼長",
			"**一個回看根數都沒宣告的算式，重演會先讀預設的 200 根暖機**",
			"宣告一個小的 lookbackCount（例如算式真正需要的根數）就會取代那 200 根",
			"交易策略的重演以**所有信號來源裡最大的回看根數**當暖機",
			"縮短區間、減少（或宣告一個小的）回看根數、換細一點的彙總刻度；價位太多就放寬 orderFlowPriceStep",
			"**要的超過上述任何一個上限，機器人會被停下**",
		} {
			assert.Contains(t, description, phrase, abilityName)
		}
	}
}

func TestAnOrderFlowPriceStepIsForwardedUntouched(t *testing.T) {
	declared := `[{"name":"寬","kind":"orderFlowPriceStep","defaultValue":0.5}]`

	request, buildError := apiToolNamed(t, "trading_create_strategy_script").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"name":           json.RawMessage(`"足跡"`),
			"script":         json.RawMessage(`"package main"`),
			"marketDataKind": json.RawMessage(`"contractKCandle"`),
			"parameters":     json.RawMessage(declared),
		}))

	require.NoError(t, buildError)
	sentBody := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(request.Body, &sentBody))
	assert.JSONEq(t, declared, string(sentBody["parameters"]))
}
