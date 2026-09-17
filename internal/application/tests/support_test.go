package application_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface/mocks"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/persistence"
	"go.uber.org/mock/gomock"
)

// now is the one instant every test reasons from, so that "expired" is a statement
// about the data rather than about how long the test took to run.
var now = time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)

const (
	aConnection       = "connection-a"
	anotherConnection = "connection-b"
)

// connector is everything under test wired the way it is wired in production: the real
// services, the real domain models, the real store — and mocks only where this
// connector stops and the world begins.
//
// Testing through this rather than through the services directly is deliberate. It
// means a test of "signing in again works" exercises the identity keeping, the expiry
// arithmetic and the request building all at once, which is where the mistakes
// actually are.
type connector struct {
	apiTools       *application.ApiToolApplication
	authentication *application.AuthenticationApplication
	tradingService *mocks.MockITradingServiceProxy
}

func newConnector(t *testing.T, apiTools ...domains.ApiToolDomain) *connector {
	t.Helper()

	controller := gomock.NewController(t)
	tradingServiceProxy := mocks.NewMockITradingServiceProxy(controller)

	clock := mocks.NewMockIClock(controller)
	clock.EXPECT().Now().Return(now).AnyTimes()

	authenticationService := service.NewAuthenticationService(
		tradingServiceProxy, persistence.NewSignedInSessionRepository(), clock)

	return &connector{
		apiTools: application.NewApiToolApplication(
			service.NewApiToolService(apiTools, authenticationService, tradingServiceProxy)),
		authentication: application.NewAuthenticationApplication(authenticationService),
		tradingService: tradingServiceProxy,
	}
}

// grantedUntil is a pair of proofs with the two expiries a test cares about.
func grantedUntil(accessExpiry time.Time, refreshExpiry time.Time, accessToken string) vo.SessionGrantVo {
	return vo.SessionGrantVo{
		Outcome: vo.TradingServiceSucceeded,
		Tokens: vo.NewTokenPairVo(
			accessToken, accessExpiry, accessToken+"-refresh", refreshExpiry),
	}
}

// aLiveGrant is the ordinary case: signed in just now, good for a while.
func aLiveGrant(accessToken string) vo.SessionGrantVo {
	return grantedUntil(now.Add(15*time.Minute), now.Add(30*24*time.Hour), accessToken)
}

// anExpiredGrant is signed in, but the short half has run out while the long one has not.
func anExpiredGrant(accessToken string) vo.SessionGrantVo {
	return grantedUntil(now.Add(-time.Minute), now.Add(30*24*time.Hour), accessToken)
}

// aStaleGrant is past saving: both halves are gone.
func aStaleGrant(accessToken string) vo.SessionGrantVo {
	return grantedUntil(now.Add(-time.Hour), now.Add(-time.Minute), accessToken)
}

// listStrategyScripts is an ability that must know who is asking.
func listStrategyScripts() domains.ApiToolDomain {
	return domains.NewApiToolDomain(
		"trading_list_strategy_scripts", "列出策略腳本",
		vo.RequestVerbRead, "/strategy-scripts", true)
}

// checkHealth is an ability that must not.
func checkHealth() domains.ApiToolDomain {
	return domains.NewApiToolDomain(
		"trading_health", "確認活著", vo.RequestVerbRead, "/health", false)
}

func noArguments() map[string]json.RawMessage {
	return map[string]json.RawMessage{}
}

func succeededWith(content string) vo.TradingServiceResponseVo {
	return vo.TradingServiceResponseVo{Outcome: vo.TradingServiceSucceeded, Content: content}
}

func refusedWith(content string) vo.TradingServiceResponseVo {
	return vo.TradingServiceResponseVo{Outcome: vo.TradingServiceRefused, Content: content}
}

func notRecognized() vo.TradingServiceResponseVo {
	return vo.TradingServiceResponseVo{
		Outcome: vo.TradingServiceIdentityNotRecognized, Content: "請重新登入"}
}

// domainsErrSignInRequired keeps the sentinel's import out of every test file that
// only needs to name it once.
func domainsErrSignInRequired() error {
	return domains.ErrSignInRequired
}

// domainsErrSignInExpired is the other half of that pair.
func domainsErrSignInExpired() error {
	return domains.ErrSignInExpired
}
