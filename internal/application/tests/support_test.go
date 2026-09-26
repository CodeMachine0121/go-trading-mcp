package application_test

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface/mocks"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
	"go.uber.org/mock/gomock"
)

const jamesAccessToken = "james-token"

type connector struct {
	apiTools       *application.ApiToolApplication
	tradingService *mocks.MockITradingServiceProxy
}

func newConnector(t *testing.T, apiTools ...domains.ApiToolDomain) *connector {
	t.Helper()

	tradingServiceProxy := mocks.NewMockITradingServiceProxy(gomock.NewController(t))

	return &connector{
		apiTools: application.NewApiToolApplication(
			service.NewApiToolService(apiTools, tradingServiceProxy)),
		tradingService: tradingServiceProxy,
	}
}

func listStrategyScripts() domains.ApiToolDomain {
	return domains.NewApiToolDomain(
		"trading_list_strategy_scripts", "列出策略腳本",
		vo.RequestVerbRead, "/strategy-scripts")
}

func checkHealth() domains.ApiToolDomain {
	return domains.NewApiToolDomain(
		"trading_health", "確認活著", vo.RequestVerbRead, "/health")
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
