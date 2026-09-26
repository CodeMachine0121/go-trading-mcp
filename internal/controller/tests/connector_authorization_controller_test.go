package controller_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/controller"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/clock"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/persistence"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/tradingservice"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	protectedResourceUrl   = "https://trading-mcp.example.com/mcp"
	resourceMetadataUrl    = "https://trading-mcp.example.com/.well-known/oauth-protected-resource/mcp"
	authorizationServerUrl = "https://trading-api.example.com"
)

func guardedConnector(t *testing.T, introspectionAnswers http.HandlerFunc) (*httptest.Server, *string) {
	t.Helper()

	tradingServiceStandIn := httptest.NewServer(introspectionAnswers)
	t.Cleanup(tradingServiceStandIn.Close)

	connectorAuthorizationController := controller.NewConnectorAuthorizationController(
		application.NewConnectorAuthorizationApplication(service.NewConnectorAuthorizationService(
			tradingservice.NewTradingServiceProxy(tradingServiceStandIn.URL, 5*time.Second),
			persistence.NewConnectorAuthorizationVerdictRepository(),
			clock.NewClock(),
			protectedResourceUrl,
		)),
		protectedResourceUrl, resourceMetadataUrl, authorizationServerUrl)

	reachedUserId := ""
	router := http.NewServeMux()
	router.Handle("/.well-known/oauth-protected-resource", connectorAuthorizationController.MetadataHandler())
	router.Handle("/mcp", connectorAuthorizationController.Guard(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			reachedUserId = auth.TokenInfoFromContext(request.Context()).UserID
			writer.WriteHeader(http.StatusOK)
		})))

	connectorStandIn := httptest.NewServer(router)
	t.Cleanup(connectorStandIn.Close)

	return connectorStandIn, &reachedUserId
}

func judging(body string) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(body))
	}
}

func liveJudgementFor(audience string) string {
	return `{"active":true,"sub":"42","aud":"` + audience + `","exp":` +
		jsonNumber(time.Now().Add(15*time.Minute).Unix()) + `}`
}

func jsonNumber(value int64) string {
	encoded, _ := json.Marshal(value)

	return string(encoded)
}

func callConnector(t *testing.T, connectorUrl string, authorization string) *http.Response {
	t.Helper()

	request, _ := http.NewRequest(http.MethodPost, connectorUrl+"/mcp", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	response, requestError := http.DefaultClient.Do(request)
	require.NoError(t, requestError)
	t.Cleanup(func() { _ = response.Body.Close() })

	return response
}

func TestACallWithoutAValidConnectorAuthorizationIsTurnedAwayAndPointedAtTheMetadata(t *testing.T) {
	testCases := []struct {
		name          string
		authorization string
		judgement     string
	}{
		{"沒有帶授權", "", liveJudgementFor(protectedResourceUrl)},
		{"不是 Bearer", "Basic amFtZXM6c2VjcmV0", liveJudgementFor(protectedResourceUrl)},
		{"交易服務判定失效", "Bearer stale", `{"active":false}`},
		{"發給別的服務", "Bearer elsewhere", liveJudgementFor("https://elsewhere.example.com/mcp")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			connectorStandIn, reachedUserId := guardedConnector(t, judging(testCase.judgement))

			response := callConnector(t, connectorStandIn.URL, testCase.authorization)

			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
			assert.Equal(t, `Bearer resource_metadata="`+resourceMetadataUrl+`"`,
				response.Header.Get("WWW-Authenticate"))
			assert.Empty(t, *reachedUserId)
		})
	}
}

func TestAValidConnectorAuthorizationReachesTheConnectorAsItsUser(t *testing.T) {
	testCases := []struct {
		name     string
		audience string
	}{
		{"發給這個外掛", protectedResourceUrl},
		{"只差結尾斜線", protectedResourceUrl + "/"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			connectorStandIn, reachedUserId := guardedConnector(t, judging(liveJudgementFor(testCase.audience)))

			response := callConnector(t, connectorStandIn.URL, "Bearer live")

			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "42", *reachedUserId)
		})
	}
}

func TestATradingServiceThatCannotBeReachedIsNotAnsweredAsARejectedAuthorization(t *testing.T) {
	connectorStandIn, _ := guardedConnector(t, func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	})

	response := callConnector(t, connectorStandIn.URL, "Bearer live")
	body, _ := io.ReadAll(response.Body)

	assert.NotEqual(t, http.StatusUnauthorized, response.StatusCode)
	assert.Empty(t, response.Header.Get("WWW-Authenticate"))
	assert.Contains(t, string(body), "連不到交易服務")
}

func TestTheMetadataSaysWhichResourceThisIsAndWhereToGetAnAuthorization(t *testing.T) {
	connectorStandIn, _ := guardedConnector(t, judging(`{"active":false}`))

	response, requestError := http.Get(connectorStandIn.URL + "/.well-known/oauth-protected-resource")
	require.NoError(t, requestError)
	t.Cleanup(func() { _ = response.Body.Close() })
	body, _ := io.ReadAll(response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.JSONEq(t, `{
		"resource": "`+protectedResourceUrl+`",
		"authorization_servers": ["`+authorizationServerUrl+`"],
		"bearer_methods_supported": ["header"]}`, string(body))
}
