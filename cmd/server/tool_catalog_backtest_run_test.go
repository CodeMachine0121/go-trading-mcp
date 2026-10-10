package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryReplayTeachesSubmittingThenLookingAtTheRun(t *testing.T) {
	for _, abilityName := range replayAbilityNames {
		t.Run(abilityName, func(t *testing.T) {
			description := apiToolNamed(t, abilityName).ToDefinitionDto().Description

			for _, taught := range []string{
				"它不等算完", "trading_get_backtest_run", "把輪次編號說給使用者",
				"會自動開始，不要重送", "你同時最多", "超過時間上限",
			} {
				assert.Contains(t, description, taught)
			}
			for _, noLongerSaid := range []string{"整體的時間上限", "交到你手上之前會先精簡", "回來的是成績單"} {
				assert.NotContains(t, description, noLongerSaid)
			}
		})
	}
}

func TestEveryReplayStillSendsWhatItSentBefore(t *testing.T) {
	expectedPaths := map[string]string{
		"trading_backtest_strategy_script":           "/backtests",
		"trading_backtest_trading_strategy":          "/trading-strategies/11/backtests",
		"trading_backtest_contract_strategy_script":  "/contract-backtests",
		"trading_backtest_contract_trading_strategy": "/trading-strategies/11/contract-backtests",
	}

	for _, abilityName := range replayAbilityNames {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)
			for _, requiredName := range []string{"symbol", "startTime", "endTime", "initialCapital"} {
				box, isDeclared := boxNamed(ability, requiredName)
				require.True(t, isDeclared, requiredName)
				assert.True(t, box.IsRequired, requiredName)
			}

			request, buildError := apiToolNamed(t, abilityName).BuildRequest(replayArgumentsWith(nil))
			require.NoError(t, buildError)
			assert.Equal(t, vo.RequestVerbSubmit, request.Verb)
			assert.Equal(t, expectedPaths[abilityName], request.Path)
		})
	}
}

func TestLookingAtABacktestRunCanWaitAWhile(t *testing.T) {
	ability := abilityNamed(t, "trading_get_backtest_run")
	idBox, hasID := boxNamed(ability, "id")
	require.True(t, hasID)
	assert.True(t, idBox.IsRequired)
	waitBox, hasWait := boxNamed(ability, "waitSeconds")
	require.True(t, hasWait)
	assert.False(t, waitBox.IsRequired)

	request, buildError := apiToolNamed(t, "trading_get_backtest_run").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"id":          json.RawMessage(`40`),
			"waitSeconds": json.RawMessage(`30`),
		}))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbRead, request.Verb)
	assert.Equal(t, "/backtest-runs/40", request.Path)
	assert.Equal(t, map[string]string{"waitSeconds": "30"}, request.Query)
	assert.Empty(t, request.Body)
}

func TestLookingAtABacktestRunWithoutWaitingSendsNoWait(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_get_backtest_run").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{"id": json.RawMessage(`40`)}))

	require.NoError(t, buildError)
	assert.Empty(t, request.Query)
}

func TestLookingAtABacktestRunTeachesHowToReadIt(t *testing.T) {
	description := abilityNamed(t, "trading_get_backtest_run").Description

	for _, taught := range []string{
		"queued", "running", "succeeded", "failed", "cancelled", "progressPercent",
		"最多 60", "超過時間上限", "被重新啟動打斷", "不是腳本壞了",
		"equityCurvePointTotalCount", "closedTradeTotalCount", "404", "前面還有在跑的",
	} {
		assert.Contains(t, description, taught)
	}
}

func TestListingBacktestRunsFiltersOnlyWhatWasAsked(t *testing.T) {
	ability := abilityNamed(t, "trading_list_backtest_runs")
	for _, optionalName := range []string{"status", "strategyScriptId", "tradingStrategyId", "limit"} {
		box, isDeclared := boxNamed(ability, optionalName)
		require.True(t, isDeclared, optionalName)
		assert.False(t, box.IsRequired, optionalName)
	}

	request, buildError := apiToolNamed(t, "trading_list_backtest_runs").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"status": json.RawMessage(`"failed"`),
			"limit":  json.RawMessage(`20`),
		}))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbRead, request.Verb)
	assert.Equal(t, "/backtest-runs", request.Path)
	assert.Equal(t, map[string]string{"status": "failed", "limit": "20"}, request.Query)

	for _, taught := range []string{"totalCount", "新到舊", "預設 20", "最多 200", "不要憑記憶重跑"} {
		assert.Contains(t, ability.Description, taught)
	}
}

func TestCancellingABacktestRunNamesTheRun(t *testing.T) {
	ability := abilityNamed(t, "trading_cancel_backtest_run")
	idBox, hasID := boxNamed(ability, "id")
	require.True(t, hasID)
	assert.True(t, idBox.IsRequired)

	request, buildError := apiToolNamed(t, "trading_cancel_backtest_run").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{"id": json.RawMessage(`42`)}))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbSubmit, request.Verb)
	assert.Equal(t, "/backtest-runs/42/cancel", request.Path)

	for _, taught := range []string{"排隊中", "進行中", "已經結束"} {
		assert.Contains(t, ability.Description, taught)
	}
}
