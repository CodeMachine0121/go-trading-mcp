package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aStrategyScriptWrite is everything writing a strategy script requires, and nothing
// else, so each test adds only the box it is about.
func aStrategyScriptWrite() map[string]json.RawMessage {
	return map[string]json.RawMessage{
		"id":     json.RawMessage(`"7"`),
		"name":   json.RawMessage(`"資金費率觀察"`),
		"script": json.RawMessage(`"package main"`),
	}
}

// The kind goes out exactly as it was said, and only when it was said. A blank that
// left here as a kind would be the connector deciding the trading service's default
// for it — and on a rewrite that default is "keep", which no spelling can say.
func TestWritingAStrategyScriptForwardsTheMarketKindOnlyWhenGiven(t *testing.T) {
	testCases := []struct {
		name          string
		abilityName   string
		marketKind    string
		expectedInKey bool
	}{
		{name: "建立時說合約行情", abilityName: "trading_create_strategy_script",
			marketKind: `"contractKCandle"`, expectedInKey: true},
		{name: "建立時沒說", abilityName: "trading_create_strategy_script"},
		// The connector does not know which spellings exist; the trading service does.
		{name: "建立時說一種不認得的", abilityName: "trading_create_strategy_script",
			marketKind: `"options"`, expectedInKey: true},
		{name: "修改時照抄合約行情", abilityName: "trading_update_strategy_script",
			marketKind: `"contractKCandle"`, expectedInKey: true},
		{name: "修改時沒提", abilityName: "trading_update_strategy_script"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			filledIn := aStrategyScriptWrite()
			if testCase.marketKind != "" {
				filledIn["marketDataKind"] = json.RawMessage(testCase.marketKind)
			}

			request, buildError := apiToolNamed(t, testCase.abilityName).
				BuildRequest(domains.NewToolArgumentsDomain(filledIn))
			require.NoError(t, buildError)

			sentBody := map[string]json.RawMessage{}
			require.NoError(t, json.Unmarshal(request.Body, &sentBody))
			sentKind, isSent := sentBody["marketDataKind"]
			assert.Equal(t, testCase.expectedInKey, isSent)
			if testCase.expectedInKey {
				assert.JSONEq(t, testCase.marketKind, string(sentKind))
			}
			// The rest of the script still goes out, so this is not passing on an
			// empty body.
			assert.JSONEq(t, `"資金費率觀察"`, string(sentBody["name"]))
		})
	}
}

// One box, shown by both writing abilities, says both meanings of leaving it out —
// and names the only two kinds there are.
func TestTheMarketKindBoxSaysWhatLeavingItOutMeans(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_script", "trading_update_strategy_script"} {
		t.Run(abilityName, func(t *testing.T) {
			box, isDeclared := boxNamed(abilityNamed(t, abilityName), "marketDataKind")

			require.True(t, isDeclared)
			assert.False(t, box.IsRequired, "不說行情種類的舊寫法要照舊有效")
			assert.Equal(t, string(vo.ToolParameterKindString), box.Kind)
			assert.Contains(t, box.Description, "kCandle（現貨 K 線）或 contractKCandle")
			assert.Contains(t, box.Description, "建立時不給就是 kCandle")
			assert.Contains(t, box.Description, "修改時不給就是保留原本的")
			assert.Contains(t, box.Description, "建立後不得更換")
		})
	}
}

// Every other box on a rewrite empties when left out; this one does not, and an
// assistant reading only the ability's own sentence would believe it does.
func TestRewritingAStrategyScriptSaysTheKindIsKept(t *testing.T) {
	description := abilityNamed(t, "trading_update_strategy_script").Description

	assert.Contains(t, description, "沒帶到的欄位會變成空的")
	assert.Contains(t, description, "唯一的例外是 marketDataKind")
	assert.Contains(t, description, "不給就是保留原本的行情種類")
	assert.Contains(t, description, "行情種類建立後不得更換")
	assert.Contains(t, description, "要吃另一種請另建一支")
}

// The script box tells the two entry points apart, because the one written for the
// other kind does not run.
func TestTheScriptBoxTeachesBothEntryPoints(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_script", "trading_update_strategy_script"} {
		t.Run(abilityName, func(t *testing.T) {
			box, isDeclared := boxNamed(abilityNamed(t, abilityName), "script")

			require.True(t, isDeclared)
			assert.Contains(t, box.Description, "吃 K 線的收 []indicator.KCandle")
			assert.Contains(t, box.Description, "吃合約行情的收 []indicator.ContractKCandle")
		})
	}
}

// Every figure a contract bar carries, by the name a script reads it under.
func TestTheContractScriptNoteNamesEveryFigure(t *testing.T) {
	for _, phrase := range []string{
		"func Calculate(data []indicator.ContractKCandle)",
		// The spot figures, under the same names.
		"現貨 K 線有的每一項這裡都有、而且同名",
		"Symbol、OpenTimeUnixSeconds、Open、High、Low、Close、Volume、QuoteVolume、TakerBuyBaseVolume、TakerBuyQuoteVolume",
		// And what is added.
		"TradeCount", "indicator.PriceLine", "Mark、Index、PremiumIndex", "Open／High／Low／Close",
		"FundingRate", "FundingSettledInBar", "OpenInterest、OpenInterestValue",
		"AccountLongShare、AccountShortShare、AccountLongShortRatio",
		"TopTraderPositionLongShare、TopTraderPositionShortShare、TopTraderPositionLongShortRatio",
		// A missing value is a zero, in the three ways it happens.
		"沒有值一律是零，分不出「沒錄到」與「真的是零」",
		"舊資料沒有指數價格與溢價指數", "第一次結算之前沒有費率", "持倉統計只留三十天",
		// A carried rate is not a payment.
		"資金費率每一格都延續上一次結算的費率，只有 FundingSettledInBar 為真的那一格才是真的收付",
	} {
		assert.Contains(t, contractKCandleScriptNote, phrase)
	}
}

// Where a contract script cannot go yet, said before the assistant takes it there.
func TestTheContractScriptNoteSaysWhereItCannotGoYet(t *testing.T) {
	assert.Contains(t, contractKCandleScriptNote, "目前只能用在 trading_calculate_contract_indicator")
	assert.Contains(t, contractKCandleScriptNote, "重演、交易策略、策略機器人還不能用它")
	assert.Contains(t, contractKCandleScriptNote, "直接告訴他目前做不到")
}

// The writing abilities read the note word for word, not a paraphrase of it.
func TestWritingAStrategyScriptTeachesTheContractShape(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_script", "trading_update_strategy_script"} {
		t.Run(abilityName, func(t *testing.T) {
			assert.Contains(t, abilityNamed(t, abilityName).Description, contractKCandleScriptNote)
		})
	}
}

// Every read of a strategy script says the kind comes back with it, because the kind
// decides which calculation the script can be handed to.
func TestEveryStrategyScriptReadSaysItCarriesTheKind(t *testing.T) {
	for _, abilityName := range []string{
		"trading_list_strategy_scripts",
		"trading_get_strategy_script",
		"trading_browse_marketplace",
	} {
		t.Run(abilityName, func(t *testing.T) {
			assert.Contains(t, abilityNamed(t, abilityName).Description, "marketDataKind")
		})
	}
}
