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
			// Exactly two, said as exactly two.
			assert.Contains(t, box.Description, "二選一：kCandle（現貨 K 線）或 contractKCandle")
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
		"舊資料沒有指數價格與溢價指數", "第一次結算之前沒有費率",
		// A statistic exists wherever it was recorded or synced, not only for thirty days.
		"持倉統計**只有錄到或同步過的那段才有值**", "trading_sync_contract_k_candle_history",
		// A carried rate is not a payment.
		"資金費率每一格都延續上一次結算的費率，只有 FundingSettledInBar 為真的那一格才是真的收付",
	} {
		assert.Contains(t, contractKCandleScriptNote, phrase)
	}
}

// Where a contract script can go — a contract bot included — said before the assistant
// takes it there, and never rewritten into a spot one to fit a spot bot.
func TestTheContractScriptNoteSaysWhereItCanGo(t *testing.T) {
	for _, phrase := range []string{
		"trading_calculate_contract_indicator",
		"trading_backtest_contract_strategy_script",
		"當吃合約行情的交易策略的信號來源",
		"掛上一台合約機器人",
		"marketDataKind 給 contractKCandle",
		"不要替使用者改寫成一支吃 K 線的",
	} {
		assert.Contains(t, contractKCandleScriptNote, phrase)
	}
	// The old sentences would now turn a person away from things that work.
	assert.NotContains(t, contractKCandleScriptNote, "持倉統計只留三十天")
	assert.NotContains(t, contractKCandleScriptNote, "重演、交易策略、策略機器人還不能用它")
	assert.NotContains(t, contractKCandleScriptNote, "策略機器人目前只跑 K 線")
	assert.NotContains(t, contractKCandleScriptNote, "直接告訴他目前做不到")
}

// Every ability that has the assistant write or run a contract script reads the note
// word for word, not a paraphrase of it: two wordings of one shape get a script that
// fits only one of them.
func TestTheContractScriptShapeIsTaughtOnceInBothPlaces(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_strategy_script",
		"trading_update_strategy_script",
		"trading_calculate_contract_indicator",
	} {
		t.Run(abilityName, func(t *testing.T) {
			assert.Contains(t, abilityNamed(t, abilityName).Description, contractKCandleScriptNote)
		})
	}
}

// Every read of a strategy script says the kind comes back with it, because the kind
// decides which calculation the script can be handed to.
//
// The whole claim, not the bare word: a description that named marketDataKind only to
// say something else about it would leave the assistant not looking for it.
func TestEveryStrategyScriptReadSaysItCarriesTheKind(t *testing.T) {
	testCases := []struct {
		abilityName string
		mustSay     string
	}{
		{abilityName: "trading_list_strategy_scripts", mustSay: "每一支都帶著 marketDataKind"},
		{abilityName: "trading_get_strategy_script", mustSay: "以及它吃哪一種行情（marketDataKind）"},
		{abilityName: "trading_browse_marketplace", mustSay: "每一支都帶著 marketDataKind"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			assert.Contains(t, abilityNamed(t, testCase.abilityName).Description, testCase.mustSay)
		})
	}
}

// aContractCalculation is one filled-in contract indicator calculation naming an
// existing strategy script.
func aContractCalculation() map[string]json.RawMessage {
	return map[string]json.RawMessage{
		"strategyScriptId":    json.RawMessage(`7`),
		"symbol":              json.RawMessage(`"BTCUSDT"`),
		"startTime":           json.RawMessage(`"2026-09-22T08:00:00Z"`),
		"endTime":             json.RawMessage(`"2026-09-23T08:00:00Z"`),
		"aggregationInterval": json.RawMessage(`"1h"`),
	}
}

// Naming a contract strategy script for one day of BTCUSDT asks the contract line's
// calculation, as whoever is signed in, with everything that was said.
func TestTheContractIndicatorCalculationAsksTheContractLine(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_calculate_contract_indicator").
		BuildRequest(domains.NewToolArgumentsDomain(aContractCalculation()))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbSubmit, request.Verb)
	assert.Equal(t, "/contract-indicator-calculations", request.Path)
	assert.JSONEq(t, `{
		"strategyScriptId": 7,
		"symbol": "BTCUSDT",
		"startTime": "2026-09-22T08:00:00Z",
		"endTime": "2026-09-23T08:00:00Z",
		"aggregationInterval": "1h"
	}`, string(request.Body))
}

// A script brought along goes out with its result type, its knobs and this time's
// settings of them, untouched.
func TestTheContractIndicatorCalculationCarriesABroughtAlongScript(t *testing.T) {
	filledIn := aContractCalculation()
	delete(filledIn, "strategyScriptId")
	filledIn["script"] = json.RawMessage(`"package main"`)
	filledIn["resultType"] = json.RawMessage(`"float"`)
	filledIn["parameters"] = json.RawMessage(`[{"name":"期數","kind":"lookbackCount","defaultValue":20}]`)
	filledIn["parameterValues"] = json.RawMessage(`[{"name":"期數","value":12}]`)

	request, buildError := apiToolNamed(t, "trading_calculate_contract_indicator").
		BuildRequest(domains.NewToolArgumentsDomain(filledIn))

	require.NoError(t, buildError)
	sentBody := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(request.Body, &sentBody))
	assert.JSONEq(t, `"package main"`, string(sentBody["script"]))
	assert.JSONEq(t, `"float"`, string(sentBody["resultType"]))
	assert.JSONEq(t, `[{"name":"期數","kind":"lookbackCount","defaultValue":20}]`, string(sentBody["parameters"]))
	assert.JSONEq(t, `[{"name":"期數","value":12}]`, string(sentBody["parameterValues"]))
}

// The trading service asks both calculations for the same body, so both declare the
// same boxes, required alike, and every one of them goes into the body. Apart from
// the symbol, word for word too — they are read out of one list.
func TestBothIndicatorCalculationsTakeTheSameBoxes(t *testing.T) {
	spot := abilityNamed(t, "trading_calculate_indicator")
	contract := abilityNamed(t, "trading_calculate_contract_indicator")

	require.Len(t, contract.Parameters, len(spot.Parameters))
	for _, spotBox := range spot.Parameters {
		contractBox, isDeclared := boxNamed(contract, spotBox.Name)

		require.True(t, isDeclared, "合約指標計算少了這一格：%s", spotBox.Name)
		assert.Equal(t, spotBox.IsRequired, contractBox.IsRequired, spotBox.Name)
		assert.Equal(t, spotBox.Kind, contractBox.Kind, spotBox.Name)
		if spotBox.Name != "symbol" {
			assert.Equal(t, spotBox, contractBox, "這一格應該讀自同一份清單：%s", spotBox.Name)
		}
	}

	for _, abilityName := range []string{"trading_calculate_indicator", "trading_calculate_contract_indicator"} {
		filledIn := everyBoxFilledIn()
		request, buildError := apiToolNamed(t, abilityName).BuildRequest(domains.NewToolArgumentsDomain(filledIn))
		require.NoError(t, buildError)

		sentBody := map[string]json.RawMessage{}
		require.NoError(t, json.Unmarshal(request.Body, &sentBody))
		assert.Empty(t, request.Query, abilityName)
		for _, box := range abilityNamed(t, abilityName).Parameters {
			assert.Contains(t, sentBody, box.Name, "%s 的 %s 該在內文裡", abilityName, box.Name)
		}
	}
}

// Without a contract, or without the start of the stretch, nothing leaves — and the
// answer names the box.
func TestTheContractIndicatorCalculationStopsAMissingBox(t *testing.T) {
	for _, leftOut := range []string{"symbol", "startTime"} {
		t.Run(leftOut, func(t *testing.T) {
			filledIn := aContractCalculation()
			delete(filledIn, leftOut)

			_, buildError := apiToolNamed(t, "trading_calculate_contract_indicator").
				BuildRequest(domains.NewToolArgumentsDomain(filledIn))

			assert.ErrorIs(t, buildError, domains.ErrRequiredArgumentMissing)
			assert.ErrorContains(t, buildError, leftOut)
		})
	}
}

// What will get a contract calculation refused, said before it is sent.
func TestTheContractIndicatorCalculationSaysWhatWillGetItRefused(t *testing.T) {
	description := abilityNamed(t, "trading_calculate_contract_indicator").Description

	assert.Contains(t, description, "會被拒絕的情況")
	assert.Contains(t, description, "指名的那一支**吃的是 K 線**（那一支要用 trading_calculate_indicator 算）")
	assert.Contains(t, description, "湊不出最少可算根數時整次拒絕")
	assert.Contains(t, description, "**可用根數**與**最少可算根數**")
	assert.Contains(t, description, "入口照現貨收 K 線")
	assert.Contains(t, description, "與 trading_calculate_indicator 一模一樣")
}

// The spot calculation refuses a contract strategy script, and says where it goes.
func TestTheSpotCalculationSaysAContractScriptIsRefused(t *testing.T) {
	description := abilityNamed(t, "trading_calculate_indicator").Description

	assert.Contains(t, description, "這一支只算現貨 K 線")
	assert.Contains(t, description, "指名一支吃合約行情（marketDataKind 為 contractKCandle）的策略腳本會被拒絕")
	assert.Contains(t, description, "那一支要用 trading_calculate_contract_indicator 算")
}

// Everything else about the spot calculation is as it was: the same address, the same
// boxes, the same ones required.
func TestTheSpotCalculationStillAsksWhereItAlwaysDid(t *testing.T) {
	spot := abilityNamed(t, "trading_calculate_indicator")

	type boxShape struct {
		kind       vo.ToolParameterKind
		isRequired bool
	}
	shapePerBoxNames := map[string]boxShape{}
	for _, box := range spot.Parameters {
		shapePerBoxNames[box.Name] = boxShape{kind: vo.ToolParameterKind(box.Kind), isRequired: box.IsRequired}
	}
	assert.Equal(t, map[string]boxShape{
		"strategyScriptId":    {vo.ToolParameterKindInteger, false},
		"symbol":              {vo.ToolParameterKindString, true},
		"startTime":           {vo.ToolParameterKindString, true},
		"endTime":             {vo.ToolParameterKindString, false},
		"aggregationInterval": {vo.ToolParameterKindString, false},
		"script":              {vo.ToolParameterKindString, false},
		"resultType":          {vo.ToolParameterKindString, false},
		"parameters":          {vo.ToolParameterKindArray, false},
		"parameterValues":     {vo.ToolParameterKindArray, false},
	}, shapePerBoxNames)

	symbol, isDeclared := boxNamed(spot, "symbol")
	require.True(t, isDeclared)
	assert.Equal(t, "要算哪一個交易標的", symbol.Description)

	filledIn := aContractCalculation()
	request, buildError := apiToolNamed(t, "trading_calculate_indicator").
		BuildRequest(domains.NewToolArgumentsDomain(filledIn))
	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbSubmit, request.Verb)
	assert.Equal(t, "/indicator-calculations", request.Path)
}
