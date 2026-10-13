package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Each kind is pinned together with what it is, so swapping two definitions fails rather than passes.
func TestTheParameterBoxNamesEveryKindTheTradingServiceAccepts(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_script", "trading_update_strategy_script"} {
		box, hasBox := boxNamed(abilityNamed(t, abilityName), "parameters")
		require.True(t, hasBox, abilityName)

		for _, phrase := range []string{
			"kind 有四種：lookbackCount（大於零的整數，這條線要回看幾根",
			"number（任何數字，原樣交給算式）",
			"boolean（是非，0 為否、非零為是）",
			"orderFlowPriceStep（**訂單流價位寬度**：大於零、可以有小數、最細 0.00000001；" +
				"**只給 marketDataKind 為 contractKCandle 的策略腳本、一支至多一個**，K 線種類宣告會被拒絕；" +
				"宣告了，每一格合約行情格的 Footprint 才有內容，價位依這個寬度分）",
			`indicator.Boolean("名字")`,
		} {
			assert.Contains(t, box.Description, phrase, abilityName)
		}
		assert.NotContains(t, box.Description, "只有兩種", abilityName)
	}
}

func TestTheContractScriptNoteNamesTheFootprint(t *testing.T) {
	for _, phrase := range []string{
		"Footprint（**訂單流足跡**，indicator.Footprint：PriceStep 價位寬度、Levels 一串 indicator.PriceLevel——" +
			"每一個有 Price、BuyVolume（買方主動吃單）、SellVolume（賣方主動吃單）、TradeCount，價位由低到高、只列有成交的——" +
			"Delta 買賣差、PointOfControl 成交最多的價位，平手取低）",
		"**訂單流足跡要先宣告一個 kind 為 orderFlowPriceStep 的參數才有**，沒宣告時整份是空的",
		"一次至多涵蓋 31 天、一格至多 1,000 個價位、一次至多 2,000,000 個價位",
		"訂單流不存進交易服務，每次現拿",
		"機器人一輪來不及備齊會跳過那一輪；要的超過 31 天會被停下",
	} {
		assert.Contains(t, contractKCandleScriptNote, phrase)
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
