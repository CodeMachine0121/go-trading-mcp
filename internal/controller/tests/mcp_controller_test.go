package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/controller"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/tradingservice"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const callerAuthorization = "Bearer connector-authorization"

func connectedAssistant(
	t *testing.T,
	tradingServiceAnswers http.HandlerFunc,
	callerHeaders map[string]string,
) *mcp.ClientSession {
	t.Helper()

	tradingServiceStandIn := httptest.NewServer(tradingServiceAnswers)
	t.Cleanup(tradingServiceStandIn.Close)

	tradingServiceProxy := tradingservice.NewTradingServiceProxy(
		tradingServiceStandIn.URL, 5*time.Second)
	apiToolService := service.NewApiToolService([]domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_health", "確認交易服務活著", vo.RequestVerbRead, "/health"),
		domains.NewApiToolDomain(
			"trading_list_strategy_scripts", "列出你的策略腳本",
			vo.RequestVerbRead, "/strategy-scripts"),
		domains.NewApiToolDomain(
			"trading_peek_live_k_candle",
			"看一眼即時更新。unavailable 表示分不到名額，**不會自己好**，要改觀察清單",
			vo.RequestVerbRead, "/k-candles/live",
			vo.NewToolParameterVo(
				"symbol", vo.ToolParameterKindString, "交易標的", true, vo.ToolParameterInQuery),
		).Watching(time.Second),
		domains.NewApiToolDomain(
			"trading_get_k_candle", "讀一根 K 線", vo.RequestVerbRead,
			"/k-candles/{symbol}/{openTime}",
			vo.NewToolParameterVo(
				"symbol", vo.ToolParameterKindString, "交易標的", true, vo.ToolParameterInPath),
			vo.NewToolParameterVo(
				"openTime", vo.ToolParameterKindString, "起始時間", true, vo.ToolParameterInPath),
		),
	}, tradingServiceProxy)

	server := mcp.NewServer(&mcp.Implementation{Name: "go-trading-mcp", Version: "test"}, nil)
	controller.NewApiToolController(
		application.NewApiToolApplication(apiToolService)).RegisterOn(server)

	connectorStandIn := httptest.NewServer(mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server }, nil))
	t.Cleanup(connectorStandIn.Close)

	transport := &mcp.StreamableClientTransport{
		Endpoint:   connectorStandIn.URL,
		HTTPClient: &http.Client{Transport: headerAdding{headers: callerHeaders}},
	}

	assistantSession, connectError := mcp.NewClient(
		&mcp.Implementation{Name: "assistant", Version: "test"}, nil).
		Connect(context.Background(), transport, nil)
	require.NoError(t, connectError)
	t.Cleanup(func() { _ = assistantSession.Close() })

	return assistantSession
}

func authorizedAssistant(t *testing.T, tradingServiceAnswers http.HandlerFunc) *mcp.ClientSession {
	t.Helper()

	return connectedAssistant(t, tradingServiceAnswers,
		map[string]string{"Authorization": callerAuthorization})
}

type headerAdding struct {
	headers map[string]string
}

func (headerAdding headerAdding) RoundTrip(request *http.Request) (*http.Response, error) {
	for name, value := range headerAdding.headers {
		request.Header.Set(name, value)
	}

	return http.DefaultTransport.RoundTrip(request)
}

func alwaysAnswering(statusCode int, body string) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(statusCode)
		_, _ = writer.Write([]byte(body))
	}
}

func textOf(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()

	require.NotEmpty(t, result.Content)
	textContent, isText := result.Content[0].(*mcp.TextContent)
	require.True(t, isText)

	return textContent.Text
}

func TestTheAssistantIsShownEveryAbilityWithItsFormAndNoSignInNote(t *testing.T) {
	assistantSession := authorizedAssistant(t, alwaysAnswering(http.StatusOK, `{}`))

	listed, listError := assistantSession.ListTools(context.Background(), nil)

	require.NoError(t, listError)

	toolsPerNames := map[string]*mcp.Tool{}
	for _, tool := range listed.Tools {
		toolsPerNames[tool.Name] = tool
		assert.NotContains(t, tool.Description, "trading_sign_in")
		assert.NotContains(t, tool.Description, "需要身分")
	}

	assert.NotContains(t, toolsPerNames, "trading_sign_in")
	assert.NotContains(t, toolsPerNames, "trading_sign_out")
	assert.NotContains(t, toolsPerNames, "trading_renew_session")
	assert.Contains(t, toolsPerNames, "trading_health")

	encodedSchema, _ := json.Marshal(toolsPerNames["trading_get_k_candle"].InputSchema)
	assert.Contains(t, string(encodedSchema), `"symbol"`)
	assert.Contains(t, string(encodedSchema), `"openTime"`)
	assert.Contains(t, string(encodedSchema), `"required"`)
}

func TestEveryAbilityForwardsTheCallersConnectorAuthorization(t *testing.T) {
	testCases := []struct {
		name      string
		toolName  string
		arguments map[string]any
	}{
		{"個人資源", "trading_list_strategy_scripts", nil},
		{"公開資料", "trading_get_k_candle",
			map[string]any{"symbol": "BTCUSDT", "openTime": "2026-08-28T09:00:00Z"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			forwardedAuthorization := ""
			assistantSession := authorizedAssistant(t, func(writer http.ResponseWriter, request *http.Request) {
				forwardedAuthorization = request.Header.Get("Authorization")
				_, _ = writer.Write([]byte(`[{"id":9}]`))
			})

			result, callError := assistantSession.CallTool(context.Background(),
				&mcp.CallToolParams{Name: testCase.toolName, Arguments: testCase.arguments})

			require.NoError(t, callError)
			assert.False(t, result.IsError)
			assert.Equal(t, callerAuthorization, forwardedAuthorization)
		})
	}
}

func TestAnAuthorizationTheTradingServiceDoesNotRecognizeAsksToReconnectFromClaudeCode(t *testing.T) {
	askedTimes := 0
	assistantSession := authorizedAssistant(t, func(writer http.ResponseWriter, _ *http.Request) {
		askedTimes++
		writer.WriteHeader(http.StatusUnauthorized)
	})

	result, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Contains(t, textOf(t, result), "/mcp")
	assert.Contains(t, textOf(t, result), "重新連線")
	assert.Equal(t, 1, askedTimes)
}

func TestAnAuthorizationThatIsNotABearerProofIsNotForwarded(t *testing.T) {
	forwardedAuthorization := "untouched"
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, request *http.Request) {
		forwardedAuthorization = request.Header.Get("Authorization")
		_, _ = writer.Write([]byte(`[]`))
	}, map[string]string{"Authorization": "Basic amFtZXM6c2VjcmV0"})

	_, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	require.NoError(t, callError)
	assert.Empty(t, forwardedAuthorization)
}

func TestAnAccountNobodyHasLetInYetIsRefusedInTheTradingServicesOwnWords(t *testing.T) {
	assistantSession := authorizedAssistant(t, alwaysAnswering(http.StatusForbidden,
		`{"message":"帳號尚未開通，請寄信申請開通","activationInstruction":{"requestMailbox":"gatekeeper@example.com"}}`))

	result, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Contains(t, textOf(t, result), "帳號尚未開通")
	assert.Contains(t, textOf(t, result), "gatekeeper@example.com")
	assert.NotContains(t, textOf(t, result), "重新連線")
}

func TestEveryAttemptLeavesATraceThatNamesNoSecret(t *testing.T) {
	var recorded bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&recorded, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	assistantSession := authorizedAssistant(t, alwaysAnswering(http.StatusOK, `{"email":"james@example.com"}`))

	_, _ = assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_health"})

	trace := recorded.String()

	assert.Contains(t, trace, "trading_health")
	assert.Contains(t, trace, "outcome=succeeded")
	assert.NotContains(t, trace, "connector-authorization")
	assert.NotContains(t, trace, "james@example.com")
}

func TestAHalfFilledFormIsAnsweredWithoutTroublingTheTradingService(t *testing.T) {
	askedAnyway := false
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, _ *http.Request) {
		askedAnyway = true
		writer.WriteHeader(http.StatusOK)
	}, nil)

	result, callError := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_get_k_candle",
		Arguments: map[string]any{"symbol": "BTCUSDT"},
	})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Contains(t, textOf(t, result), "openTime")
	assert.False(t, askedAnyway, "少了必填欄位就不必去打擾交易服務")
}

func TestAnAbilityThisConnectorDoesNotHaveIsNotOfferedAtAll(t *testing.T) {
	assistantSession := authorizedAssistant(t, alwaysAnswering(http.StatusOK, `{}`))

	_, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_place_an_order"})

	require.Error(t, callError)
	assert.True(t, strings.Contains(strings.ToLower(callError.Error()), "unknown") ||
		strings.Contains(callError.Error(), "trading_place_an_order"))
}

func TestBoxesThatAreNotEvenASetOfBoxesAreAnsweredRatherThanCrashed(t *testing.T) {
	assistantSession := authorizedAssistant(t, alwaysAnswering(http.StatusOK, `{}`))

	result, callError := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_health",
		Arguments: 5,
	})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Contains(t, textOf(t, result), "不是一組可以讀的資料")
}

func TestATraceSaysWhichKindOfFailureItWas(t *testing.T) {
	var recorded bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&recorded, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	assistantSession := connectedAssistant(t,
		alwaysAnswering(http.StatusNotFound, `{"message":"這根 K 線不存在"}`), nil)

	_, _ = assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_get_k_candle",
		Arguments: map[string]any{"symbol": "BTCUSDT", "openTime": "2026-08-28T09:00:00Z"},
	})

	trace := recorded.String()

	assert.Contains(t, trace, "outcome=refusedByTradingService")
	assert.NotContains(t, trace, "這根 K 線不存在",
		"回覆內容不進紀錄——它可能裝著使用者的資料")
}

func TestAnAbilityWithNoLiveSlotSaysSoAndSaysItWillNotFixItself(t *testing.T) {
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte(
			"data: {\"symbol\":\"2330\",\"status\":\"unavailable\"}\n\n"))
	}, nil)

	result, callError := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_peek_live_k_candle",
		Arguments: map[string]any{"symbol": "2330"},
	})

	require.NoError(t, callError)
	assert.False(t, result.IsError, "分不到名額是查到的事，不是這次呼叫做錯了")
	assert.Contains(t, textOf(t, result), "unavailable")

	listed, _ := assistantSession.ListTools(context.Background(), nil)
	for _, tool := range listed.Tools {
		if tool.Name == "trading_peek_live_k_candle" {
			assert.Contains(t, tool.Description, "不會自己好",
				"助理要知道這一種等下去也不會好，得去改觀察清單")
			assert.Contains(t, tool.Description, "觀察清單")
		}
	}
}

func TestARefusalReachesTheAssistantInTheTradingServicesOwnWordsAndIsMarkedAsOne(t *testing.T) {
	assistantSession := connectedAssistant(t,
		alwaysAnswering(http.StatusNotFound, `{"message":"這根 K 線不存在"}`), nil)

	result, callError := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "trading_get_k_candle",
		Arguments: map[string]any{
			"symbol": "BTCUSDT", "openTime": "2026-08-28T09:00:00Z"},
	})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Contains(t, textOf(t, result), "這根 K 線不存在")
}
