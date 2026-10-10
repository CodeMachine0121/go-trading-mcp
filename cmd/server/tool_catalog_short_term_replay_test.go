package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// replayAbilityNames are the four abilities that replay, spot and contract, a script
// and a trading strategy.
var replayAbilityNames = []string{
	"trading_backtest_strategy_script",
	"trading_backtest_trading_strategy",
	"trading_backtest_contract_strategy_script",
	"trading_backtest_contract_trading_strategy",
}

func replayArgumentsWith(extra map[string]json.RawMessage) domains.ToolArgumentsDomain {
	arguments := map[string]json.RawMessage{
		"id":             json.RawMessage(`11`),
		"script":         json.RawMessage(`"package main"`),
		"symbol":         json.RawMessage(`"BTCUSDT"`),
		"startTime":      json.RawMessage(`"2026-01-01T00:00:00Z"`),
		"endTime":        json.RawMessage(`"2026-02-01T00:00:00Z"`),
		"initialCapital": json.RawMessage(`"10000"`),
	}
	for name, value := range extra {
		arguments[name] = value
	}

	return domains.NewToolArgumentsDomain(arguments)
}

func TestEveryReplayForwardsTheFillTimingAndTheValidationStart(t *testing.T) {
	for _, abilityName := range replayAbilityNames {
		t.Run(abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, abilityName).BuildRequest(replayArgumentsWith(map[string]json.RawMessage{
				"fillTiming":          json.RawMessage(`"nextOpen"`),
				"validationStartTime": json.RawMessage(`"2026-01-21T00:00:00Z"`),
			}))

			require.NoError(t, buildError)
			assert.Contains(t, string(request.Body), `"fillTiming":"nextOpen"`)
			assert.Contains(t, string(request.Body), `"validationStartTime":"2026-01-21T00:00:00Z"`)
		})
	}
}

func TestAReplayWithNeitherSendsWhatItSentBefore(t *testing.T) {
	for _, abilityName := range replayAbilityNames {
		request, buildError := apiToolNamed(t, abilityName).BuildRequest(replayArgumentsWith(nil))

		require.NoError(t, buildError)
		assert.NotContains(t, string(request.Body), "fillTiming", abilityName)
		assert.NotContains(t, string(request.Body), "validationStartTime", abilityName)
	}
}

func TestOnlyWhatIsReplayedOnTheSpotOrLookedAtWithAWaitWaitsLonger(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10*time.Second, 120*time.Second) {
		request, buildError := apiTool.BuildRequest(replayArgumentsWith(nil))
		if buildError != nil {
			continue
		}

		waitsLonger := apiTool.Name() == "trading_compare_contract_trades_with_backtest" ||
			apiTool.Name() == "trading_compare_spot_trades_with_backtest" ||
			apiTool.Name() == "trading_get_backtest_run"

		if waitsLonger {
			assert.Equal(t, 120*time.Second, request.ResponseWaitLimit, apiTool.Name())
			continue
		}
		assert.Zero(t, request.ResponseWaitLimit, "只有當場重演與看一筆回測輪次該等比較久：%s", apiTool.Name())
	}
}

func TestOnlyLookingAtABacktestRunCondensesWhatItHandsOn(t *testing.T) {
	equityPoints := make([]string, 0, 300)
	for pointIndex := range 300 {
		equityPoints = append(equityPoints, fmt.Sprintf(`{"equity":"%d"}`, pointIndex))
	}
	longResult := `{"summary":{},"equityCurve":[` + strings.Join(equityPoints, ",") + `]}`
	succeeded := vo.TradingServiceResponseVo{Outcome: vo.TradingServiceSucceeded, Content: longResult}

	for _, apiTool := range apiToolCatalog(10*time.Second, 120*time.Second) {
		relayed := apiTool.Relayed(succeeded).Content
		if apiTool.Name() == "trading_get_backtest_run" {
			assert.Contains(t, relayed, `"equityCurvePointTotalCount":300`)
			continue
		}
		assert.Equal(t, longResult, relayed, "只有看一筆回測輪次會精簡：%s", apiTool.Name())
	}
}

func TestEveryReplayTeachesShortTermReplaying(t *testing.T) {
	for _, abilityName := range replayAbilityNames {
		description := apiToolNamed(t, abilityName).ToDefinitionDto().Description

		for _, taught := range []string{
			"fillTiming 用 nextOpen",
			"只拿 inSample 調參數，只拿 validation 判斷這支策略好不好",
			"profitFactor", "expectancy", "averageHoldingSeconds", "maximumConsecutiveLossCount", "costToGrossProfitRatio",
			"一筆都沒虧時是 null", "價差本身沒賺時是 null",
			"不要數交易明細",
		} {
			assert.Contains(t, description, taught, abilityName)
		}
	}
}

func TestTheSpotScriptReplayNoLongerSaysItHasNoEquityCurve(t *testing.T) {
	description := apiToolNamed(t, "trading_backtest_strategy_script").ToDefinitionDto().Description

	assert.NotContains(t, description, "沒有資金曲線")
}
