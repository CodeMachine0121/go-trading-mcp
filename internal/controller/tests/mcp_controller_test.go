package controller_test

import (
	"context"
	"encoding/json"
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
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/clock"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/persistence"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/tradingservice"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectedAssistant is a real MCP client talking to a real MCP server over a real
// HTTP transport, with only the trading service itself replaced by a stand-in.
//
// Going through the whole stack is the point: the two things this layer is
// responsible for — which connection a call arrived on, and a proof the caller brought
// along in a header — do not exist anywhere below it, and neither can be tested by
// calling a method directly.
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
	authenticationService := service.NewAuthenticationService(
		tradingServiceProxy, persistence.NewSignedInSessionRepository(), clock.NewClock())
	apiToolService := service.NewApiToolService([]domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_health", "確認交易服務活著", vo.RequestVerbRead, "/health", false),
		domains.NewApiToolDomain(
			"trading_list_strategy_scripts", "列出你的策略腳本",
			vo.RequestVerbRead, "/strategy-scripts", true),
		domains.NewApiToolDomain(
			"trading_get_k_candle", "讀一根 K 線", vo.RequestVerbRead,
			"/k-candles/{symbol}/{openTime}", false,
			vo.NewToolParameterVo(
				"symbol", vo.ToolParameterKindString, "交易標的", true, vo.ToolParameterInPath),
			vo.NewToolParameterVo(
				"openTime", vo.ToolParameterKindString, "起始時間", true, vo.ToolParameterInPath),
		),
	}, authenticationService, tradingServiceProxy)

	server := mcp.NewServer(&mcp.Implementation{Name: "go-trading-mcp", Version: "test"}, nil)
	controller.NewAuthenticationController(
		application.NewAuthenticationApplication(authenticationService)).RegisterOn(server)
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

// headerAdding puts the caller's own headers on every request, which is how a caller
// that already holds a proof hands it over.
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

func TestTheAssistantIsShownEveryAbilityWithItsFormAndItsSignInNote(t *testing.T) {
	assistantSession := connectedAssistant(t, alwaysAnswering(http.StatusOK, `{}`), nil)

	listed, listError := assistantSession.ListTools(context.Background(), nil)

	require.NoError(t, listError)

	toolsPerNames := map[string]*mcp.Tool{}
	for _, tool := range listed.Tools {
		toolsPerNames[tool.Name] = tool
	}

	assert.Contains(t, toolsPerNames, "trading_sign_in")
	assert.Contains(t, toolsPerNames, "trading_sign_out")
	assert.Contains(t, toolsPerNames, "trading_renew_session")
	assert.Contains(t, toolsPerNames, "trading_health")

	assert.NotContains(t, toolsPerNames["trading_health"].Description, "需要身分")
	assert.Contains(t, toolsPerNames["trading_list_strategy_scripts"].Description, "需要身分")

	encodedSchema, _ := json.Marshal(toolsPerNames["trading_get_k_candle"].InputSchema)
	assert.Contains(t, string(encodedSchema), `"symbol"`)
	assert.Contains(t, string(encodedSchema), `"openTime"`)
	assert.Contains(t, string(encodedSchema), `"required"`)
}

func TestAnAbilityNeedingNoIdentityWorksStraightAway(t *testing.T) {
	assistantSession := connectedAssistant(t,
		alwaysAnswering(http.StatusOK, `{"status":"Healthy"}`), nil)

	result, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_health"})

	require.NoError(t, callError)
	assert.False(t, result.IsError)
	assert.JSONEq(t, `{"status":"Healthy"}`, textOf(t, result))
}

func TestAnAbilityNeedingIdentityAsksTheUserToSignInFirst(t *testing.T) {
	assistantSession := connectedAssistant(t, alwaysAnswering(http.StatusOK, `[]`), nil)

	result, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Equal(t, "請先登入", textOf(t, result))
}

func TestSigningInThenWorkingNeedsNoFurtherMentionOfTheAccount(t *testing.T) {
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/sessions" {
			_, _ = writer.Write([]byte(`{
				"accessToken":"james-token","expiresAt":"2099-01-01T00:00:00Z",
				"refreshToken":"james-r","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`))
			return
		}

		if request.Header.Get("Authorization") != "Bearer james-token" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, _ = writer.Write([]byte(`[{"id":1,"name":"布林通道"}]`))
	}, nil)

	signedIn, _ := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})
	listed, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	assert.False(t, signedIn.IsError)
	assert.Contains(t, textOf(t, signedIn), "james@example.com")
	assert.False(t, listed.IsError)
	assert.Contains(t, textOf(t, listed), "布林通道")
}

func TestSigningInHandsBackNoProofAtAll(t *testing.T) {
	assistantSession := connectedAssistant(t, alwaysAnswering(http.StatusOK, `{
		"accessToken":"secret-access","expiresAt":"2099-01-01T00:00:00Z",
		"refreshToken":"secret-refresh","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`), nil)

	signedIn, _ := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})

	everythingTheAssistantSees := textOf(t, signedIn)
	assert.NotContains(t, everythingTheAssistantSees, "secret-access")
	assert.NotContains(t, everythingTheAssistantSees, "secret-refresh")
	assert.NotContains(t, everythingTheAssistantSees, "correct horse")
}

func TestSigningOutLeavesTheConnectionAsNobody(t *testing.T) {
	assistantSession := connectedAssistant(t, alwaysAnswering(http.StatusOK, `{
		"accessToken":"james-token","expiresAt":"2099-01-01T00:00:00Z",
		"refreshToken":"james-r","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`), nil)

	_, _ = assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})
	signedOut, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_sign_out"})
	afterwards, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	assert.False(t, signedOut.IsError)
	assert.True(t, afterwards.IsError)
	assert.Equal(t, "請先登入", textOf(t, afterwards))
}

func TestSigningOutWithoutHavingSignedInIsStillSuccess(t *testing.T) {
	assistantSession := connectedAssistant(t, alwaysAnswering(http.StatusOK, `{}`), nil)

	signedOut, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_sign_out"})

	require.NoError(t, callError)
	assert.False(t, signedOut.IsError)
}

func TestAProofTheCallerCarriesInAHeaderIsUsedWithoutSigningIn(t *testing.T) {
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer brought-along" {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, _ = writer.Write([]byte(`[{"id":9}]`))
	}, map[string]string{"Authorization": "Bearer brought-along"})

	result, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	require.NoError(t, callError)
	assert.False(t, result.IsError)
	assert.Contains(t, textOf(t, result), `"id":9`)
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
	assistantSession := connectedAssistant(t, alwaysAnswering(http.StatusOK, `{}`), nil)

	_, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_place_an_order"})

	require.Error(t, callError)
	assert.True(t, strings.Contains(strings.ToLower(callError.Error()), "unknown") ||
		strings.Contains(callError.Error(), "trading_place_an_order"))
}

func TestOneAssistantsSigningInIsInvisibleToAnother(t *testing.T) {
	tradingServiceStandIn := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/sessions" {
				_, _ = writer.Write([]byte(`{
					"accessToken":"james-token","expiresAt":"2099-01-01T00:00:00Z",
					"refreshToken":"james-r","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`))
				return
			}

			_, _ = writer.Write([]byte(`[]`))
		}))
	t.Cleanup(tradingServiceStandIn.Close)

	tradingServiceProxy := tradingservice.NewTradingServiceProxy(
		tradingServiceStandIn.URL, 5*time.Second)
	authenticationService := service.NewAuthenticationService(
		tradingServiceProxy, persistence.NewSignedInSessionRepository(), clock.NewClock())
	apiToolService := service.NewApiToolService([]domains.ApiToolDomain{
		domains.NewApiToolDomain("trading_list_strategy_scripts", "列出策略腳本",
			vo.RequestVerbRead, "/strategy-scripts", true),
	}, authenticationService, tradingServiceProxy)

	server := mcp.NewServer(&mcp.Implementation{Name: "go-trading-mcp", Version: "test"}, nil)
	controller.NewAuthenticationController(
		application.NewAuthenticationApplication(authenticationService)).RegisterOn(server)
	controller.NewApiToolController(
		application.NewApiToolApplication(apiToolService)).RegisterOn(server)

	connectorStandIn := httptest.NewServer(mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server }, nil))
	t.Cleanup(connectorStandIn.Close)

	connect := func() *mcp.ClientSession {
		assistantSession, connectError := mcp.NewClient(
			&mcp.Implementation{Name: "assistant", Version: "test"}, nil).
			Connect(context.Background(),
				&mcp.StreamableClientTransport{Endpoint: connectorStandIn.URL}, nil)
		require.NoError(t, connectError)
		t.Cleanup(func() { _ = assistantSession.Close() })

		return assistantSession
	}

	signedInAssistant := connect()
	strangerAssistant := connect()

	_, _ = signedInAssistant.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})

	signedInResult, _ := signedInAssistant.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})
	strangerResult, _ := strangerAssistant.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	assert.False(t, signedInResult.IsError)
	assert.True(t, strangerResult.IsError)
	assert.Equal(t, "請先登入", textOf(t, strangerResult))
}

func TestARefusedSigningInReachesTheAssistantInTheTradingServicesOwnWords(t *testing.T) {
	assistantSession := connectedAssistant(t,
		alwaysAnswering(http.StatusUnauthorized, `{"message":"電子郵件或密碼不正確"}`), nil)

	result, callError := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "wrong"},
	})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Contains(t, textOf(t, result), "電子郵件或密碼不正確")
}

func TestSigningInWhileTheTradingServiceIsAwaySaysItIsAway(t *testing.T) {
	tradingServiceProxy := tradingservice.NewTradingServiceProxy("http://127.0.0.1:1", time.Second)
	authenticationService := service.NewAuthenticationService(
		tradingServiceProxy, persistence.NewSignedInSessionRepository(), clock.NewClock())

	server := mcp.NewServer(&mcp.Implementation{Name: "go-trading-mcp", Version: "test"}, nil)
	controller.NewAuthenticationController(
		application.NewAuthenticationApplication(authenticationService)).RegisterOn(server)

	connectorStandIn := httptest.NewServer(mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server }, nil))
	t.Cleanup(connectorStandIn.Close)

	assistantSession, connectError := mcp.NewClient(
		&mcp.Implementation{Name: "assistant", Version: "test"}, nil).
		Connect(context.Background(),
			&mcp.StreamableClientTransport{Endpoint: connectorStandIn.URL}, nil)
	require.NoError(t, connectError)
	t.Cleanup(func() { _ = assistantSession.Close() })

	signedIn, _ := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})
	signedOut, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_sign_out"})

	assert.True(t, signedIn.IsError)
	assert.Contains(t, textOf(t, signedIn), "連不到交易服務")
	assert.False(t, signedOut.IsError, "沒登入過的連線登出仍然是成功")
}

func TestRenewingOnDemandReachesTheAssistantAsOneOfThreeDifferentAnswers(t *testing.T) {
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/sessions" {
			_, _ = writer.Write([]byte(`{
				"accessToken":"james-token","expiresAt":"2099-01-01T00:00:00Z",
				"refreshToken":"james-r","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`))
			return
		}

		_, _ = writer.Write([]byte(`{
			"accessToken":"fresh","expiresAt":"2099-01-01T00:00:00Z",
			"refreshToken":"fresh-r","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`))
	}, nil)

	beforeSigningIn, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_renew_session"})

	_, _ = assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})

	afterSigningIn, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_renew_session"})

	assert.True(t, beforeSigningIn.IsError)
	assert.Equal(t, "請先登入", textOf(t, beforeSigningIn))
	assert.False(t, afterSigningIn.IsError)
}

func TestRenewingWhenTheRenewalItselfIsRefusedAsksForAFreshSigningIn(t *testing.T) {
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/sessions" {
			_, _ = writer.Write([]byte(`{
				"accessToken":"james-token","expiresAt":"2099-01-01T00:00:00Z",
				"refreshToken":"james-r","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`))
			return
		}

		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte(`{"message":"請重新登入"}`))
	}, nil)

	_, _ = assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})
	renewed, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_renew_session"})

	assert.True(t, renewed.IsError)
	assert.Equal(t, "登入已失效，請重新登入", textOf(t, renewed))
}

func TestBoxesThatAreNotEvenASetOfBoxesAreAnsweredRatherThanCrashed(t *testing.T) {
	assistantSession := connectedAssistant(t, alwaysAnswering(http.StatusOK, `{}`), nil)

	result, callError := assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_health",
		Arguments: 5,
	})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Contains(t, textOf(t, result), "不是一組可以讀的資料")
}

func TestAnAuthorizationThatIsNotABearerProofIsNotMistakenForOne(t *testing.T) {
	assistantSession := connectedAssistant(t, func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`[]`))
	}, map[string]string{"Authorization": "Basic amFtZXM6c2VjcmV0"})

	result, callError := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	require.NoError(t, callError)
	assert.True(t, result.IsError)
	assert.Equal(t, "請先登入", textOf(t, result),
		"不是 Bearer 的就不是這裡的憑證，硬拿去用只會被交易服務拒絕得莫名其妙")
}

func TestLosingTheTradingServiceWhileSigningOutStillGivesUpTheIdentityLocally(t *testing.T) {
	tradingServiceStandIn := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			_, _ = writer.Write([]byte(`{
				"accessToken":"james-token","expiresAt":"2099-01-01T00:00:00Z",
				"refreshToken":"james-r","refreshTokenExpiresAt":"2099-01-01T00:00:00Z"}`))
		}))

	tradingServiceProxy := tradingservice.NewTradingServiceProxy(
		tradingServiceStandIn.URL, time.Second)
	authenticationService := service.NewAuthenticationService(
		tradingServiceProxy, persistence.NewSignedInSessionRepository(), clock.NewClock())
	apiToolService := service.NewApiToolService([]domains.ApiToolDomain{
		domains.NewApiToolDomain("trading_list_strategy_scripts", "列出策略腳本",
			vo.RequestVerbRead, "/strategy-scripts", true),
	}, authenticationService, tradingServiceProxy)

	server := mcp.NewServer(&mcp.Implementation{Name: "go-trading-mcp", Version: "test"}, nil)
	controller.NewAuthenticationController(
		application.NewAuthenticationApplication(authenticationService)).RegisterOn(server)
	controller.NewApiToolController(
		application.NewApiToolApplication(apiToolService)).RegisterOn(server)

	connectorStandIn := httptest.NewServer(mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server }, nil))
	t.Cleanup(connectorStandIn.Close)

	assistantSession, connectError := mcp.NewClient(
		&mcp.Implementation{Name: "assistant", Version: "test"}, nil).
		Connect(context.Background(),
			&mcp.StreamableClientTransport{Endpoint: connectorStandIn.URL}, nil)
	require.NoError(t, connectError)
	t.Cleanup(func() { _ = assistantSession.Close() })

	_, _ = assistantSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "trading_sign_in",
		Arguments: map[string]any{"email": "james@example.com", "password": "correct horse"},
	})

	tradingServiceStandIn.Close()

	signedOut, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_sign_out"})
	afterwards, _ := assistantSession.CallTool(context.Background(),
		&mcp.CallToolParams{Name: "trading_list_strategy_scripts"})

	assert.True(t, signedOut.IsError, "換發鏈可能沒撤掉，這件事要說出來")
	assert.Contains(t, textOf(t, signedOut), "連不到交易服務")
	assert.Equal(t, "請先登入", textOf(t, afterwards),
		"但這台機器上的身分是真的放掉了——那一半不需要交易服務同意")
}
