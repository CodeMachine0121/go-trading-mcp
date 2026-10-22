package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var chartDrawingAbilityNames = []string{
	"trading_get_chart_drawings",
	"trading_add_chart_drawings",
	"trading_remove_chart_drawings",
}

const horizontalLineSnappedToLow = `{"kind":"horizontalLine","anchors":[{"openTime":"2026-10-19T08:00:00Z","snapTo":"low"}],` +
	`"color":"red","lineWidth":2,"lineStyle":"dashed"}`

func TestEveryChartDrawingAbilityAsksTheRightBoard(t *testing.T) {
	testCases := []struct {
		abilityName   string
		arguments     map[string]string
		expectedVerb  vo.RequestVerb
		expectedPath  string
		expectedQuery map[string]string
	}{
		{"trading_get_chart_drawings", map[string]string{"market": `"contract"`, "symbol": `"btcusdt"`},
			vo.RequestVerbRead, "/users/me/chart-drawings/contract/btcusdt", map[string]string{}},
		{"trading_get_chart_drawings", map[string]string{"market": `"spot"`, "symbol": `"ETHUSDT"`},
			vo.RequestVerbRead, "/users/me/chart-drawings/spot/ETHUSDT", map[string]string{}},
		{"trading_get_chart_drawings", map[string]string{"market": `"期貨"`, "symbol": `"BTCUSDT"`},
			vo.RequestVerbRead, "/users/me/chart-drawings/%E6%9C%9F%E8%B2%A8/BTCUSDT", map[string]string{}},
		{"trading_add_chart_drawings",
			map[string]string{"market": `"contract"`, "symbol": `"BTCUSDT"`, "interval": `"1h"`,
				"drawings": "[" + horizontalLineSnappedToLow + "]"},
			vo.RequestVerbSubmit, "/users/me/chart-drawings/contract/BTCUSDT/assistant-drawings", map[string]string{}},
		{"trading_remove_chart_drawings",
			map[string]string{"market": `"contract"`, "symbol": `"BTCUSDT"`, "ids": `"12"`},
			vo.RequestVerbRemove, "/users/me/chart-drawings/contract/BTCUSDT/assistant-drawings",
			map[string]string{"ids": "12"}},
		{"trading_remove_chart_drawings",
			map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`, "ids": `"12,15"`},
			vo.RequestVerbRemove, "/users/me/chart-drawings/spot/BTCUSDT/assistant-drawings",
			map[string]string{"ids": "12,15"}},
		{"trading_remove_chart_drawings",
			map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`},
			vo.RequestVerbRemove, "/users/me/chart-drawings/spot/BTCUSDT/assistant-drawings", map[string]string{}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName+" "+testCase.expectedPath, func(t *testing.T) {
			request, buildError := apiToolNamed(t, testCase.abilityName).BuildRequest(
				filledInArguments(testCase.arguments))

			require.NoError(t, buildError)
			assert.Equal(t, testCase.expectedVerb, request.Verb)
			assert.Equal(t, testCase.expectedPath, request.Path)
			assert.Equal(t, testCase.expectedQuery, request.Query)
		})
	}
}

func TestReadingOrRemovingSendsNothingInTheContents(t *testing.T) {
	testCases := []struct {
		abilityName string
		arguments   map[string]string
	}{
		{"trading_get_chart_drawings", map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`}},
		{"trading_remove_chart_drawings", map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`, "ids": `"12"`}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, testCase.abilityName).BuildRequest(
				filledInArguments(testCase.arguments))

			require.NoError(t, buildError)
			assert.Nil(t, request.Body)
		})
	}
}

func TestAddingDrawingsHandsTheIntervalAndEveryDrawingOverUntouched(t *testing.T) {
	twentyDrawings := make([]string, 0, 20)
	for index := range 20 {
		twentyDrawings = append(twentyDrawings, fmt.Sprintf(
			`{"kind":"verticalLine","anchors":[{"openTime":"2026-10-19T%02d:00:00Z","price":"61200.5"}]}`, index))
	}
	testCases := []struct {
		name     string
		interval string
		drawings string
	}{
		{name: "a horizontal line snapped to the low", interval: `"1h"`, drawings: "[" + horizontalLineSnappedToLow + "]"},
		{name: "twenty drawings at once", interval: `"1h"`, drawings: "[" + strings.Join(twentyDrawings, ",") + "]"},
		{name: "an anchor off the candle start the trading service will refuse", interval: `"1h"`,
			drawings: `[{"kind":"horizontalLine","anchors":[{"openTime":"2026-10-19T08:30:00Z","snapTo":"low"}]}]`},
		{name: "a brush the trading service will refuse", interval: `"1h"`,
			drawings: `[{"kind":"brush","anchors":[{"openTime":"2026-10-19T08:37:12Z","price":"61200.5"},` +
				`{"openTime":"2026-10-19T08:41:05Z","price":"61250"}]}]`},
		{name: "a trend line priced by number", interval: `"4h"`,
			drawings: `[{"kind":"trendLine","anchors":[{"openTime":"2026-10-18T00:00:00Z","price":60000},` +
				`{"openTime":"2026-10-19T04:00:00Z","price":"63000"}]}]`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, buildError := apiToolNamed(t, "trading_add_chart_drawings").BuildRequest(filledInArguments(
				map[string]string{"market": `"contract"`, "symbol": `"BTCUSDT"`,
					"interval": testCase.interval, "drawings": testCase.drawings}))

			require.NoError(t, buildError)
			assert.JSONEq(t, `{"interval":`+testCase.interval+`,"drawings":`+testCase.drawings+`}`, string(request.Body))
		})
	}
}

func TestEveryChartDrawingAbilityNeedsItsBoxesFilledIn(t *testing.T) {
	testCases := []struct {
		abilityName string
		arguments   map[string]string
		missingName string
	}{
		{"trading_get_chart_drawings", map[string]string{"symbol": `"BTCUSDT"`}, "market"},
		{"trading_get_chart_drawings", map[string]string{"market": `"spot"`}, "symbol"},
		{"trading_add_chart_drawings",
			map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`, "drawings": "[" + horizontalLineSnappedToLow + "]"},
			"interval"},
		{"trading_add_chart_drawings",
			map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`, "interval": `"1h"`}, "drawings"},
		{"trading_remove_chart_drawings", map[string]string{"symbol": `"BTCUSDT"`}, "market"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName+" without "+testCase.missingName, func(t *testing.T) {
			_, buildError := apiToolNamed(t, testCase.abilityName).BuildRequest(
				filledInArguments(testCase.arguments))

			require.ErrorIs(t, buildError, domains.ErrRequiredArgumentMissing)
			assert.Contains(t, buildError.Error(), testCase.missingName)
		})
	}
}

func TestRemovingNamesWhichDrawingsAsOptionalText(t *testing.T) {
	box, isDeclared := boxNamed(abilityNamed(t, "trading_remove_chart_drawings"), "ids")

	require.True(t, isDeclared)
	assert.False(t, box.IsRequired)
	assert.Equal(t, string(vo.ToolParameterKindString), box.Kind)
	assert.Contains(t, box.Description, "逗號分隔")
}

func TestNoAbilityCanRewriteAWholeChartDrawingBoard(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10*time.Second, 120*time.Second) {
		arguments := map[string]string{}
		for _, parameter := range apiTool.ToDefinitionDto().Parameters {
			arguments[parameter.Name] = `"filled"`
		}

		request, buildError := apiTool.BuildRequest(filledInArguments(arguments))
		require.NoError(t, buildError, apiTool.Name())
		if !strings.HasPrefix(request.Path, "/users/me/chart-drawings/") {
			continue
		}

		isReadingTheBoard := request.Verb == vo.RequestVerbRead && !strings.HasSuffix(request.Path, "/assistant-drawings")
		isTouchingAssistantDrawings := (request.Verb == vo.RequestVerbSubmit || request.Verb == vo.RequestVerbRemove) &&
			strings.HasSuffix(request.Path, "/assistant-drawings")
		assert.True(t, isReadingTheBoard || isTouchingAssistantDrawings,
			"外掛只能讀繪圖板、新增與移除助手繪圖：%s", apiTool.Name())
	}
}

func TestChartDrawingAbilitiesTeachTheAssistantToDrawOnlyItsOwnLines(t *testing.T) {
	testCases := []struct {
		abilityName string
		taught      []string
	}{
		{"trading_get_chart_drawings", []string{
			"revision 0", "manual", "assistant", "畫線之前先讀一次", "手畫的（source 是 manual）動不了", "不要改用別的辦法去刪",
			"kind 是 brush", "一連串點", "openTime 不一定是 K 線的起始時間", "奇異筆例外",
		}},
		{"trading_add_chart_drawings", []string{
			"trading_get_k_candle_series", "trading_get_contract_k_candle_series", "interval 給與這次相同的刻度", "不要用 displayableCandleCount",
			"正好是那個刻度上一根已經存在的 K 線的起始時間", "價格優先用吸附", "不能兩個都給",
			"最多 200 個、一次最多 20 個", "還剩幾個空位", "照它說的改一改再送，不要原樣重送",
			"不必先問使用者確認", "回到網頁的看盤工作區（或重新聚焦那一頁）", "手畫的（source 是 manual）動不了",
			"沒有任何能力可以移動、改色、鎖定",
			"奇異筆（brush）不能畫", "只能由使用者手畫，助手不能新增", "送了交易服務會整批拒絕", "請他自己在網頁上畫",
		}},
		{"trading_remove_chart_drawings", []string{
			"不給 ids 即移除這份繪圖板上助手畫的全部", "整次拒絕、一個都不刪", "這份繪圖板上沒有的編號也整次拒絕", "移除了 0 個",
			"不要改用別的辦法去刪", "在網頁的看盤工作區上動",
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			description := abilityNamed(t, testCase.abilityName).Description

			for _, phrase := range testCase.taught {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestAddingDrawingsPointsOnlyAtSeriesThatHonourTheInterval(t *testing.T) {
	description := abilityNamed(t, "trading_add_chart_drawings").Description

	for _, oneMinuteOnly := range []string{"trading_list_k_candles", "trading_list_contract_k_candles"} {
		assert.NotContains(t, description, oneMinuteOnly)
	}
}

func TestAddingDrawingsSpellsOutEveryAllowedValue(t *testing.T) {
	ability := abilityNamed(t, "trading_add_chart_drawings")
	intervalBox, _ := boxNamed(ability, "interval")
	drawingsBox, _ := boxNamed(ability, "drawings")

	for _, interval := range []string{"1m", "5m", "15m", "1h", "4h", "1d"} {
		assert.Contains(t, intervalBox.Description, interval)
	}
	for _, allowed := range []string{
		"trendLine", "ray", "fibonacciRetracement", "rectangle", "dateAndPriceRange",
		"horizontalLine", "horizontalRay", "verticalLine", "各 **2 個**", "各 **1 個**",
		"snapTo", "open／high／low／close", "price",
		"red", "orange", "yellow", "green", "teal", "cyan", "blue", "indigo", "purple", "pink", "gray", "silver",
		"1 到 4", "solid", "dashed", "dotted", "省略即照那一種繪圖的預設樣式", "1 到 20 個",
		"brush（奇異筆）只能由使用者手畫，不能新增",
	} {
		assert.Contains(t, drawingsBox.Description, allowed)
	}
	assert.Equal(t, string(vo.ToolParameterKindArray), drawingsBox.Kind)
}

func TestEveryChartDrawingAbilityNamesTheBoardByMarketAndSymbol(t *testing.T) {
	for _, abilityName := range chartDrawingAbilityNames {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)

			for _, boxName := range []string{"market", "symbol"} {
				box, isDeclared := boxNamed(ability, boxName)
				require.True(t, isDeclared, boxName)
				assert.True(t, box.IsRequired, boxName)
			}
			marketBox, _ := boxNamed(ability, "market")
			assert.Contains(t, marketBox.Description, "spot（現貨）或 contract（永續合約）")
		})
	}
}

func TestChartDrawingAbilitiesWaitOnlyTheUsualTime(t *testing.T) {
	testCases := []struct {
		abilityName string
		arguments   map[string]string
	}{
		{"trading_get_chart_drawings", map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`}},
		{"trading_add_chart_drawings", map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`,
			"interval": `"1h"`, "drawings": "[" + horizontalLineSnappedToLow + "]"}},
		{"trading_remove_chart_drawings", map[string]string{"market": `"spot"`, "symbol": `"BTCUSDT"`}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, testCase.abilityName).BuildRequest(filledInArguments(testCase.arguments))

			require.NoError(t, buildError)
			assert.Zero(t, request.ResponseWaitLimit)
			assert.Zero(t, request.LiveUpdateWaitLimit)
		})
	}
}
