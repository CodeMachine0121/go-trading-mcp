package main

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aScriptWriteForm is the least a strategy script can be written with, plus whatever this case adds.
func aScriptWriteForm(extraBoxes map[string]json.RawMessage) domains.ToolArgumentsDomain {
	boxes := map[string]json.RawMessage{
		"id":     json.RawMessage(`"5"`),
		"name":   json.RawMessage(`"二十根均線"`),
		"script": json.RawMessage(`"package main"`),
	}
	for name, value := range extraBoxes {
		boxes[name] = value
	}

	return domains.NewToolArgumentsDomain(boxes)
}

// Creating a script forwards the usage only when one is given; left out, the trading
// service makes it a calculation script.
func TestCreatingAStrategyScriptForwardsTheUsageOnlyWhenGiven(t *testing.T) {
	testCases := []struct {
		name        string
		givenUsage  string
		mustSend    string
		mustNotSend string
	}{
		{name: "圖表用", givenUsage: `"chart"`, mustSend: `"usage":"chart"`},
		{name: "計算用", givenUsage: `"calculation"`, mustSend: `"usage":"calculation"`},
		{name: "不認得的用途照送，由交易服務拒絕", givenUsage: `"backtest"`, mustSend: `"usage":"backtest"`},
		{name: "沒說", mustNotSend: "usage"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			extraBoxes := map[string]json.RawMessage{}
			if testCase.givenUsage != "" {
				extraBoxes["usage"] = json.RawMessage(testCase.givenUsage)
			}

			request, buildError := apiToolNamed(t, "trading_create_strategy_script").BuildRequest(aScriptWriteForm(extraBoxes))

			require.NoError(t, buildError)
			if testCase.mustSend != "" {
				assert.Contains(t, string(request.Body), testCase.mustSend)
			}
			if testCase.mustNotSend != "" {
				assert.NotContains(t, string(request.Body), testCase.mustNotSend)
			}
		})
	}
}

func TestCreatingAStrategyScriptSaysWhatTheUsageIs(t *testing.T) {
	ability := abilityNamed(t, "trading_create_strategy_script")

	usage, isDeclared := boxNamed(ability, "usage")
	require.True(t, isDeclared)
	assert.False(t, usage.IsRequired)
	for _, phrase := range []string{"chart（圖表用）", "calculation（計算用）", "**不給就是 calculation**", "其他值會被拒絕"} {
		assert.Contains(t, usage.Description, phrase)
	}
	assert.Contains(t, ability.Description, "畫在圖上看的建立成 chart")
	assert.Contains(t, ability.Description, "之後要改用 trading_switch_strategy_script_usage")
}

// A rewrite never carries a usage: the trading service refuses a rewrite that does, so
// one the assistant sends anyway is dropped rather than forwarded to be refused.
func TestRewritingAStrategyScriptNeverCarriesAUsage(t *testing.T) {
	ability := abilityNamed(t, "trading_update_strategy_script")

	_, isDeclared := boxNamed(ability, "usage")
	assert.False(t, isDeclared)
	assert.Contains(t, ability.Description, "**改寫不動它的用途（usage），這裡也沒有那一格**")
	assert.Contains(t, ability.Description, "要改用途請用 trading_switch_strategy_script_usage")

	request, buildError := apiToolNamed(t, "trading_update_strategy_script").BuildRequest(
		aScriptWriteForm(map[string]json.RawMessage{"usage": json.RawMessage(`"chart"`)}))
	require.NoError(t, buildError)
	assert.Equal(t, "/strategy-scripts/5", request.Path)
	assert.NotContains(t, string(request.Body), "usage")
}

// Switching goes to the script's own usage address with nothing but the usage.
func TestSwitchingAStrategyScriptsUsageSendsOnlyTheUsage(t *testing.T) {
	for _, givenUsage := range []string{"chart", "calculation"} {
		t.Run(givenUsage, func(t *testing.T) {
			request, buildError := apiToolNamed(t, "trading_switch_strategy_script_usage").BuildRequest(
				domains.NewToolArgumentsDomain(map[string]json.RawMessage{
					"id":    json.RawMessage(`"5"`),
					"usage": json.RawMessage(`"` + givenUsage + `"`),
					"name":  json.RawMessage(`"不該送出"`),
				}))

			require.NoError(t, buildError)
			assert.Equal(t, vo.RequestVerbReplace, request.Verb)
			assert.Equal(t, "/strategy-scripts/5/usage", request.Path)
			assert.JSONEq(t, `{"usage":"`+givenUsage+`"}`, string(request.Body))
		})
	}
}

func TestSwitchingAStrategyScriptsUsageHoldsBackAMissingUsage(t *testing.T) {
	_, buildError := apiToolNamed(t, "trading_switch_strategy_script_usage").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{"id": json.RawMessage(`"5"`)}))

	require.ErrorIs(t, buildError, domains.ErrRequiredArgumentMissing)
	assert.ErrorContains(t, buildError, "usage")
}

func TestSwitchingAStrategyScriptsUsageSaysWhenItIsRefusedAndWhatItKeeps(t *testing.T) {
	description := abilityNamed(t, "trading_switch_strategy_script_usage").Description

	for _, phrase := range []string{
		"**從市集加入的副本也能切**", "副本的算式仍然看不到、改不動", "回覆是切換後的那一支策略腳本",
		"**切成 chart 時，這一支正被任何一份交易策略或標的篩選設定用著就會被拒絕**",
		"reason 為 strategyScriptUsageInUse", "拒絕的話會說出是哪幾份在用",
		"請他先把它從那些交易策略或篩選設定換掉，**不要重送同一個切換**",
		"切成 calculation 沒有限制", "切成它原本的用途不算失敗，什麼也不變", "不是你的（或不存在的）回「找不到」",
		"trading_update_strategy_script 改寫內容時不能順便改用途",
	} {
		assert.Contains(t, description, phrase)
	}
}

// Every ability answering with a script, and every one setting its usage, says how to read it.
func TestEveryAbilityAboutAScriptsUsageSaysWhereEachUsageGoes(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_strategy_script", "trading_list_strategy_scripts",
		"trading_get_strategy_script", "trading_switch_strategy_script_usage",
	} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"**只有計算用的才能當交易策略的信號來源、標的篩選設定的腳本、直接回測指名的腳本**",
				"網頁圖表的指標選單**只列圖表用的**", "單純算一次指標兩種都可以",
				"**升級前就有的每一支（含市集副本）一律是計算用**", "「圖表上找不到我寫的指標」",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestTheMarketplaceSaysScriptsCarryTheirUsageIntoCopies(t *testing.T) {
	assert.Contains(t, abilityNamed(t, "trading_browse_marketplace").Description, "也帶著 usage")

	adopt := abilityNamed(t, "trading_adopt_strategy_script").Description
	assert.Contains(t, adopt, "**副本帶著加入當下那一支的用途（usage）**，之後作者切換用途不影響副本")
	assert.Contains(t, adopt, "trading_switch_strategy_script_usage")
}

func TestBothTradingStrategyWritesSayOnlyCalculationScriptsMayBeSignalSources(t *testing.T) {
	for _, abilityName := range []string{"trading_create_trading_strategy", "trading_update_trading_strategy"} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			assert.Contains(t, description, "**信號來源只能指名計算用（usage 為 calculation）的策略腳本**；圖表用的會整份被拒絕")
			assert.Contains(t, description, "用 trading_switch_strategy_script_usage 切成計算用")
		})
	}
}

func TestBothScriptReplaysSayTheNamedScriptMustBeACalculationOne(t *testing.T) {
	for _, abilityName := range []string{"trading_backtest_strategy_script", "trading_backtest_contract_strategy_script"} {
		t.Run(abilityName, func(t *testing.T) {
			strategyScriptID, isDeclared := boxNamed(abilityNamed(t, abilityName), "strategyScriptId")

			require.True(t, isDeclared)
			assert.Contains(t, strategyScriptID.Description, "**指名的必須是計算用（usage 為 calculation）**，圖表用的會被拒絕")
			assert.Contains(t, strategyScriptID.Description, "自帶一段還沒存的算式不受用途影響")
		})
	}
}

func TestBothScreeningProfileWritesSayTheScriptMustBeACalculationOne(t *testing.T) {
	for _, abilityName := range []string{"trading_create_symbol_screening_profile", "trading_update_symbol_screening_profile"} {
		t.Run(abilityName, func(t *testing.T) {
			strategyScriptID, _ := boxNamed(abilityNamed(t, abilityName), "strategyScriptId")

			assert.Contains(t, strategyScriptID.Description, "**是計算用（usage 為 calculation）**——圖表用的會被拒絕")
		})
	}
}

func TestBothIndicatorCalculationsSayEitherUsageCalculates(t *testing.T) {
	for _, abilityName := range []string{"trading_calculate_indicator", "trading_calculate_contract_indicator"} {
		t.Run(abilityName, func(t *testing.T) {
			strategyScriptID, _ := boxNamed(abilityNamed(t, abilityName), "strategyScriptId")

			assert.Contains(t, strategyScriptID.Description, "**圖表用、計算用的都能算**——單純算一次指標不看用途")
		})
	}
}

// Both boxes that take a usage spell the same two, and only the switch insists on one.
func TestEveryUsageBoxSpellsTheSameTwoUsages(t *testing.T) {
	for _, testCase := range []struct {
		abilityName string
		isRequired  bool
	}{
		{abilityName: "trading_create_strategy_script", isRequired: false},
		{abilityName: "trading_switch_strategy_script_usage", isRequired: true},
	} {
		t.Run(testCase.abilityName, func(t *testing.T) {
			usage, isDeclared := boxNamed(abilityNamed(t, testCase.abilityName), "usage")

			require.True(t, isDeclared)
			assert.Equal(t, testCase.isRequired, usage.IsRequired)
			assert.Contains(t, usage.Description, "二選一：chart（圖表用）或 calculation（計算用）")
			assert.Contains(t, usage.Description, "其他值會被拒絕")
		})
	}
}
