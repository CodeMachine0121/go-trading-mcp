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

var symbolScreeningAbilityNames = []string{
	"trading_list_symbol_screening_profiles",
	"trading_get_symbol_screening_profile",
	"trading_create_symbol_screening_profile",
	"trading_update_symbol_screening_profile",
	"trading_delete_symbol_screening_profile",
	"trading_submit_symbol_screening_run",
	"trading_list_symbol_screening_runs",
	"trading_get_symbol_screening_run",
	"trading_cancel_symbol_screening_run",
}

func screeningArguments(values map[string]string) domains.ToolArgumentsDomain {
	arguments := map[string]json.RawMessage{}
	for name, value := range values {
		arguments[name] = json.RawMessage(value)
	}

	return domains.NewToolArgumentsDomain(arguments)
}

func TestEveryScreeningAbilitySaysOnlyAdministratorsMayUseItAndNotToRetry(t *testing.T) {
	for _, abilityName := range symbolScreeningAbilityNames {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, taught := range []string{"只有管理員能用", "這是管理員功能", "不要重試", "營運者在系統外開通"} {
				assert.Contains(t, description, taught)
			}
		})
	}
}

func TestEveryScreeningAbilityAsksTheRightPlace(t *testing.T) {
	testCases := []struct {
		abilityName  string
		arguments    map[string]string
		expectedVerb vo.RequestVerb
		expectedPath string
	}{
		{"trading_list_symbol_screening_profiles", nil, vo.RequestVerbRead, "/symbol-screening-profiles"},
		{"trading_get_symbol_screening_profile", map[string]string{"id": `"4"`}, vo.RequestVerbRead,
			"/symbol-screening-profiles/4"},
		{"trading_create_symbol_screening_profile",
			map[string]string{"name": `"趨勢 V3"`, "strategyScriptId": `108`, "aggregationInterval": `"4h"`},
			vo.RequestVerbSubmit, "/symbol-screening-profiles"},
		{"trading_update_symbol_screening_profile",
			map[string]string{"id": `"4"`, "name": `"趨勢 V3"`, "strategyScriptId": `108`, "aggregationInterval": `"4h"`},
			vo.RequestVerbReplace, "/symbol-screening-profiles/4"},
		{"trading_delete_symbol_screening_profile", map[string]string{"id": `"4"`}, vo.RequestVerbRemove,
			"/symbol-screening-profiles/4"},
		{"trading_submit_symbol_screening_run", map[string]string{"profileId": `4`, "symbols": `"XRPUSDT"`},
			vo.RequestVerbSubmit, "/symbol-screening-runs"},
		{"trading_list_symbol_screening_runs", nil, vo.RequestVerbRead, "/symbol-screening-runs"},
		{"trading_get_symbol_screening_run", map[string]string{"id": `"9"`}, vo.RequestVerbRead,
			"/symbol-screening-runs/9"},
		{"trading_cancel_symbol_screening_run", map[string]string{"id": `"9"`}, vo.RequestVerbSubmit,
			"/symbol-screening-runs/9/cancel"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, testCase.abilityName).BuildRequest(screeningArguments(testCase.arguments))

			require.NoError(t, buildError)
			assert.Equal(t, testCase.expectedVerb, request.Verb)
			assert.Equal(t, testCase.expectedPath, request.Path)
		})
	}
}

func TestSubmittingAScreeningSendsThePastedListUntouched(t *testing.T) {
	twentySymbols := make([]string, 0, 20)
	for index := range 20 {
		twentySymbols = append(twentySymbols, fmt.Sprintf("COIN%dUSDT", index))
	}
	testCases := []struct {
		name   string
		pasted string
	}{
		{name: "case and a full-width comma", pasted: "xrpusdt，LINKUSDT"},
		{name: "exactly twenty", pasted: strings.Join(twentySymbols, ",")},
		{name: "a character the trading service refuses", pasted: "BTC-USDT"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			pastedJson, _ := json.Marshal(testCase.pasted)

			request, buildError := apiToolNamed(t, "trading_submit_symbol_screening_run").BuildRequest(
				screeningArguments(map[string]string{"profileId": `4`, "symbols": string(pastedJson)}))

			require.NoError(t, buildError)
			sent := map[string]json.RawMessage{}
			require.NoError(t, json.Unmarshal(request.Body, &sent))
			sentSymbols := ""
			require.NoError(t, json.Unmarshal(sent["symbols"], &sentSymbols))
			assert.Equal(t, testCase.pasted, sentSymbols)
			assert.JSONEq(t, `4`, string(sent["profileId"]))
		})
	}
}

func TestSubmittingAScreeningNeedsAProfileAndSymbols(t *testing.T) {
	ability := abilityNamed(t, "trading_submit_symbol_screening_run")

	for _, requiredName := range []string{"profileId", "symbols"} {
		box, isDeclared := boxNamed(ability, requiredName)
		require.True(t, isDeclared, requiredName)
		assert.True(t, box.IsRequired, requiredName)
	}
	for _, taught := range []string{"把篩選編號說給使用者", "不要重送", "照使用者貼的原樣"} {
		assert.Contains(t, ability.Description, taught)
	}
}

func TestLookingAtAScreeningCanWaitAWhile(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     map[string]string
		expectedQuery map[string]string
	}{
		{name: "told to wait", arguments: map[string]string{"id": `"9"`, "waitSeconds": `30`},
			expectedQuery: map[string]string{"waitSeconds": "30"}},
		{name: "not told to wait", arguments: map[string]string{"id": `"9"`}, expectedQuery: map[string]string{}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, buildError := apiToolNamed(t, "trading_get_symbol_screening_run").BuildRequest(
				screeningArguments(testCase.arguments))

			require.NoError(t, buildError)
			assert.Equal(t, "/symbol-screening-runs/9", request.Path)
			if len(testCase.expectedQuery) == 0 {
				assert.Empty(t, request.Query)
			} else {
				assert.Equal(t, testCase.expectedQuery, request.Query)
			}
			assert.Equal(t, 120*time.Second, request.ResponseWaitLimit)
		})
	}
}

func TestCreatingAProfileSendsOnlyWhatTheUserSaid(t *testing.T) {
	testCases := []struct {
		name                string
		arguments           map[string]string
		expectedFields      []string
		expectedHistoryDays string
	}{
		{name: "only the essentials",
			arguments:      map[string]string{"name": `"趨勢 V3"`, "strategyScriptId": `108`, "aggregationInterval": `"4h"`},
			expectedFields: []string{"aggregationInterval", "name", "strategyScriptId"}},
		{name: "history days equal to the default are still sent",
			expectedHistoryDays: `730`,
			arguments: map[string]string{"name": `"趨勢 V3"`, "strategyScriptId": `108`, "aggregationInterval": `"4h"`,
				"historyDays": `730`},
			expectedFields: []string{"aggregationInterval", "historyDays", "name", "strategyScriptId"}},
		{name: "a fifteen-minute interval is left for the trading service to refuse",
			arguments:      map[string]string{"name": `"太細"`, "strategyScriptId": `108`, "aggregationInterval": `"15m"`},
			expectedFields: []string{"aggregationInterval", "name", "strategyScriptId"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, buildError := apiToolNamed(t, "trading_create_symbol_screening_profile").BuildRequest(
				screeningArguments(testCase.arguments))

			require.NoError(t, buildError)
			sent := map[string]json.RawMessage{}
			require.NoError(t, json.Unmarshal(request.Body, &sent))
			sentFields := []string{}
			for field := range sent {
				sentFields = append(sentFields, field)
			}
			assert.ElementsMatch(t, testCase.expectedFields, sentFields)
			if testCase.expectedHistoryDays != "" {
				assert.JSONEq(t, testCase.expectedHistoryDays, string(sent["historyDays"]))
			}
		})
	}
}

func TestBothProfileWritesTakeTheSameBoxes(t *testing.T) {
	createBoxes := abilityNamed(t, "trading_create_symbol_screening_profile").Parameters
	updateAbility := abilityNamed(t, "trading_update_symbol_screening_profile")

	for _, createBox := range createBoxes {
		updateBox, isDeclared := boxNamed(updateAbility, createBox.Name)
		require.True(t, isDeclared, createBox.Name)
		assert.Equal(t, createBox.IsRequired, updateBox.IsRequired, createBox.Name)
	}
	assert.Len(t, updateAbility.Parameters, len(createBoxes)+1)
	assert.Contains(t, updateAbility.Description, "整份改寫")
}

func TestLookingAtAScreeningTeachesHowToReadTheVerdicts(t *testing.T) {
	description := abilityNamed(t, "trading_get_symbol_screening_run").Description

	for _, taught := range []string{
		"passed", "rejected", "unfinished", "cancelled", "basicCondition", "portability",
		"gapReductionDays", "最長空窗縮短了幾天", "recommendationRank 1 是最值得考慮的", "不是買賣建議",
	} {
		assert.Contains(t, description, taught)
	}
}

func TestAskingWhoIAmSaysWhetherIAmAnAdministrator(t *testing.T) {
	description := abilityNamed(t, "trading_get_current_user").Description

	assert.Contains(t, description, "isAdmin")
	assert.Contains(t, description, "管理員")
}

func TestEveryScreeningAbilityHandsBackWhatTheTradingServiceSaidWordForWord(t *testing.T) {
	testCases := []struct {
		name     string
		response vo.TradingServiceResponseVo
	}{
		{name: "not an administrator", response: vo.TradingServiceResponseVo{
			Outcome: vo.TradingServiceRefused, Content: `{"message":"這是管理員功能"}`}},
		{name: "a symbol it cannot read", response: vo.TradingServiceResponseVo{
			Outcome: vo.TradingServiceRefused, Content: `{"message":"篩選的要求不正確: BTC-USDT 只能有英文字母與數字"}`}},
		{name: "an interval it cannot screen", response: vo.TradingServiceResponseVo{
			Outcome: vo.TradingServiceRefused, Content: `{"message":"篩選的要求不正確: 看盤間隔只能是每 4 小時或每天"}`}},
		{name: "a queued run", response: vo.TradingServiceResponseVo{
			Outcome: vo.TradingServiceSucceeded, Content: `{"id":9,"status":"queued","candidateCount":2}`}},
	}

	for _, abilityName := range symbolScreeningAbilityNames {
		for _, testCase := range testCases {
			t.Run(abilityName+" "+testCase.name, func(t *testing.T) {
				assert.Equal(t, testCase.response, apiToolNamed(t, abilityName).Relayed(testCase.response))
			})
		}
	}
}
