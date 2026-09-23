package application_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// replayStrategyScript is a replay ability, which waits longer and condenses.
func replayStrategyScript() domains.ApiToolDomain {
	return domains.NewApiToolDomain(
		"trading_backtest_strategy_script", "重演", vo.RequestVerbSubmit, "/backtests", true,
	).Waiting(120 * time.Second).CondensingReplayResults()
}

func aLongReplayResult() string {
	equityPoints := make([]string, 0, 1000)
	for pointIndex := range 1000 {
		equityPoints = append(equityPoints, fmt.Sprintf(`{"equity":"%d"}`, pointIndex))
	}

	return `{"summary":{"finalEquity":"1"},"equityCurve":[` + strings.Join(equityPoints, ",") + `]}`
}

func TestAReplayWaitsLongerAndReachesTheAssistantCondensed(t *testing.T) {
	connector := newConnector(t, replayStrategyScript())
	connector.signInOn(t, aConnection, aLiveGrant("james-token"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "james-token").
		DoAndReturn(func(_ context.Context, request vo.TradingServiceRequestVo, _ string) (vo.TradingServiceResponseVo, error) {
			assert.Equal(t, 120*time.Second, request.ResponseWaitLimit)

			return succeededWith(aLongReplayResult()), nil
		})

	resultDto := connector.call("trading_backtest_strategy_script", aConnection)

	require.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, `"equityCurvePointTotalCount":1000`)
	assert.Contains(t, resultDto.Content, `"summary":{"finalEquity":"1"}`)
}

func TestAReplayAnsweredAfterARenewedSigningInIsCondensedToo(t *testing.T) {
	connector := newConnector(t, replayStrategyScript())
	connector.signInOn(t, aConnection, aLiveGrant("stale"))
	gomock.InOrder(
		connector.tradingService.EXPECT().Send(gomock.Any(), gomock.Any(), "stale").Return(notRecognized(), nil),
		connector.tradingService.EXPECT().RenewSession(gomock.Any(), "stale-refresh").Return(aLiveGrant("fresh"), nil),
		connector.tradingService.EXPECT().Send(gomock.Any(), gomock.Any(), "fresh").Return(succeededWith(aLongReplayResult()), nil),
	)

	resultDto := connector.call("trading_backtest_strategy_script", aConnection)

	require.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, `"equityCurvePointTotalCount":1000`)
}

func TestARefusedReplayReachesTheAssistantInTheTradingServicesWords(t *testing.T) {
	connector := newConnector(t, replayStrategyScript())
	connector.signInOn(t, aConnection, aLiveGrant("james-token"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "james-token").
		Return(refusedWith(`{"message":"驗證起點必須落在期間之內","field":"validationStartTime"}`), nil)

	resultDto := connector.call("trading_backtest_strategy_script", aConnection)

	assert.Equal(t, dto.ToolOutcomeRefusedByTradingService, resultDto.Outcome)
	assert.Equal(t, `{"message":"驗證起點必須落在期間之內","field":"validationStartTime"}`, resultDto.Content)
}
