package domains_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func argumentsOf(t *testing.T, filledIn map[string]any) domains.ToolArgumentsDomain {
	t.Helper()

	values := map[string]json.RawMessage{}
	for name, value := range filledIn {
		encodedValue, encodeError := json.Marshal(value)
		require.NoError(t, encodeError)
		values[name] = encodedValue
	}

	return domains.NewToolArgumentsDomain(values)
}

func TestBuildRequestPlacesEachValueWhereItsDeclarationSays(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_update_k_candle", "改一根 K 線", vo.RequestVerbReplace,
		"/k-candles/{symbol}/{openTime}",
		vo.NewToolParameterVo("symbol", vo.ToolParameterKindString, "交易標的", true, vo.ToolParameterInPath),
		vo.NewToolParameterVo("openTime", vo.ToolParameterKindString, "起始時間", true, vo.ToolParameterInPath),
		vo.NewToolParameterVo("interval", vo.ToolParameterKindString, "刻度", false, vo.ToolParameterInQuery),
		vo.NewToolParameterVo("close", vo.ToolParameterKindString, "收盤價", true, vo.ToolParameterInBody),
	)

	request, buildError := apiTool.BuildRequest(argumentsOf(t, map[string]any{
		"symbol":   "BTCUSDT",
		"openTime": "2026-08-28T09:00:00Z",
		"interval": "1h",
		"close":    "110",
	}))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbReplace, request.Verb)
	assert.Equal(t, "/k-candles/BTCUSDT/2026-08-28T09:00:00Z", request.Path)
	assert.Equal(t, map[string]string{"interval": "1h"}, request.Query)
	assert.JSONEq(t, `{"close":"110"}`, string(request.Body))
}

func TestBuildRequestEscapesAValueThatWouldOtherwiseNameSomethingElse(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_remove_from_watchlist", "移出觀察清單", vo.RequestVerbRemove, "/watchlist/{symbol}",
		vo.NewToolParameterVo("symbol", vo.ToolParameterKindString, "代號", true, vo.ToolParameterInPath),
	)

	request, buildError := apiTool.BuildRequest(argumentsOf(t, map[string]any{"symbol": "a/b"}))

	require.NoError(t, buildError)
	assert.Equal(t, "/watchlist/a%2Fb", request.Path)
}

func TestBuildRequestLeavesOutBoxesThatWereNotFilledIn(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_list_k_candles", "查 K 線", vo.RequestVerbRead, "/k-candles",
		vo.NewToolParameterVo("symbol", vo.ToolParameterKindString, "交易標的", true, vo.ToolParameterInQuery),
		vo.NewToolParameterVo("interval", vo.ToolParameterKindString, "刻度", false, vo.ToolParameterInQuery),
	)

	request, buildError := apiTool.BuildRequest(argumentsOf(t, map[string]any{"symbol": "BTCUSDT"}))

	require.NoError(t, buildError)
	assert.Equal(t, map[string]string{"symbol": "BTCUSDT"}, request.Query)
}

func TestBuildRequestSendsNoContentsWhenNothingBelongsInside(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_list_trading_symbols", "列出交易標的", vo.RequestVerbRead, "/trading-symbols",
	)

	request, buildError := apiTool.BuildRequest(argumentsOf(t, map[string]any{}))

	require.NoError(t, buildError)
	assert.Nil(t, request.Body)
}

func TestBuildRequestNamesTheRequiredBoxThatWasLeftEmpty(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_sync_k_candle_history", "同步歷史", vo.RequestVerbSubmit, "/k-candles/history",
		vo.NewToolParameterVo("symbol", vo.ToolParameterKindString, "標的", true, vo.ToolParameterInBody),
		vo.NewToolParameterVo("lookbackDays", vo.ToolParameterKindInteger, "天數", true, vo.ToolParameterInBody),
	)

	_, buildError := apiTool.BuildRequest(argumentsOf(t, map[string]any{"symbol": "BTCUSDT"}))

	require.ErrorIs(t, buildError, domains.ErrRequiredArgumentMissing)
	assert.Contains(t, buildError.Error(), "lookbackDays")
}

func TestBuildRequestWritesNumbersAndTextIntoTheAddressTheSameWayTheyRead(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_get_strategy_script", "讀策略腳本", vo.RequestVerbRead, "/strategy-scripts/{id}",
		vo.NewToolParameterVo("id", vo.ToolParameterKindString, "識別碼", true, vo.ToolParameterInPath),
	)

	fromNumber, _ := apiTool.BuildRequest(argumentsOf(t, map[string]any{"id": 7}))
	fromText, _ := apiTool.BuildRequest(argumentsOf(t, map[string]any{"id": "7"}))

	assert.Equal(t, "/strategy-scripts/7", fromNumber.Path)
	assert.Equal(t, "/strategy-scripts/7", fromText.Path)
}

func TestWatchingSetsHowLongToStayOnTheLine(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_peek_live_k_candle", "看一眼", vo.RequestVerbRead, "/k-candles/live",
	).Watching(7_000_000_000)

	request, buildError := apiTool.BuildRequest(argumentsOf(t, map[string]any{}))

	require.NoError(t, buildError)
	assert.EqualValues(t, 7_000_000_000, request.LiveUpdateWaitLimit)
}

func TestToDefinitionDtoTellsTheAssistantEverythingItHasToKnowToChoose(t *testing.T) {
	apiTool := domains.NewApiToolDomain(
		"trading_add_to_watchlist", "加進觀察清單", vo.RequestVerbSubmit, "/watchlist",
		vo.NewToolParameterVo("symbol", vo.ToolParameterKindString, "代號", true, vo.ToolParameterInBody),
		vo.NewToolParameterVo("market", vo.ToolParameterKindString, "市場", false, vo.ToolParameterInBody),
	)

	definitionDto := apiTool.ToDefinitionDto()

	assert.Equal(t, "trading_add_to_watchlist", definitionDto.Name)
	assert.Equal(t, "加進觀察清單", definitionDto.Description)
	assert.Len(t, definitionDto.Parameters, 2)
	assert.Equal(t, "symbol", definitionDto.Parameters[0].Name)
	assert.Equal(t, "string", definitionDto.Parameters[0].Kind)
	assert.Equal(t, "代號", definitionDto.Parameters[0].Description)
	assert.True(t, definitionDto.Parameters[0].IsRequired)
	assert.False(t, definitionDto.Parameters[1].IsRequired)
}

func TestAReplayAbilityWaitsLongerAndCondensesOnlyWhatSucceeded(t *testing.T) {
	longContent := aReplayResultWith(500, 0, "")
	replay := domains.NewApiToolDomain(
		"trading_backtest_strategy_script", "重演", vo.RequestVerbSubmit, "/backtests",
	).Waiting(120 * time.Second).CondensingReplayResults()

	request, buildError := replay.BuildRequest(argumentsOf(t, map[string]any{}))
	require.NoError(t, buildError)
	assert.Equal(t, 120*time.Second, request.ResponseWaitLimit)

	succeeded := replay.Relayed(vo.TradingServiceResponseVo{Outcome: vo.TradingServiceSucceeded, Content: longContent})
	assert.NotEqual(t, longContent, succeeded.Content)
	assert.Contains(t, succeeded.Content, `"equityCurvePointTotalCount":500`)

	refused := replay.Relayed(vo.TradingServiceResponseVo{Outcome: vo.TradingServiceRefused, Content: longContent})
	assert.Equal(t, longContent, refused.Content)
}

func TestAnOrdinaryAbilityNeitherWaitsLongerNorCondenses(t *testing.T) {
	longContent := aReplayResultWith(500, 0, "")
	ordinary := domains.NewApiToolDomain("trading_list_k_candles", "查 K 線", vo.RequestVerbRead, "/k-candles")

	request, buildError := ordinary.BuildRequest(argumentsOf(t, map[string]any{}))
	require.NoError(t, buildError)
	assert.Zero(t, request.ResponseWaitLimit)

	relayed := ordinary.Relayed(vo.TradingServiceResponseVo{Outcome: vo.TradingServiceSucceeded, Content: longContent})
	assert.Equal(t, longContent, relayed.Content)
}
