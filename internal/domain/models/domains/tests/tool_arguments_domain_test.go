package domains_test

import (
	"encoding/json"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
)

func TestAValueReadsInAnAddressTheWayAPersonWroteIt(t *testing.T) {
	testCases := []struct {
		name         string
		rawValue     string
		expectedText string
	}{
		{"文字脫掉引號", `"BTCUSDT"`, "BTCUSDT"},
		{"數字原樣", `7`, "7"},
		{"是非原樣", `true`, "true"},
		{"帶引號的時間脫掉引號", `"2026-08-28T09:00:00Z"`, "2026-08-28T09:00:00Z"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			arguments := domains.NewToolArgumentsDomain(
				map[string]json.RawMessage{"value": json.RawMessage(testCase.rawValue)})

			assert.Equal(t, testCase.expectedText, arguments.PlainTextAt("value"))
		})
	}
}

func TestABoxNobodyFilledInIsEmptyRatherThanAnything(t *testing.T) {
	arguments := domains.NewToolArgumentsDomain(nil)

	assert.False(t, arguments.Has("symbol"))
	assert.Equal(t, "", arguments.PlainTextAt("symbol"))
}

func TestGatheringContentsTakesOnlyWhatWasNamedAndOnlyWhatWasFilledIn(t *testing.T) {
	arguments := domains.NewToolArgumentsDomain(map[string]json.RawMessage{
		"symbol":       json.RawMessage(`"BTCUSDT"`),
		"lookbackDays": json.RawMessage(`30`),
		"interval":     json.RawMessage(`"1h"`),
	})

	contents := arguments.EncodedSubset([]string{"symbol", "lookbackDays", "notFilledIn"})

	assert.JSONEq(t, `{"symbol":"BTCUSDT","lookbackDays":30}`, string(contents))
}

func TestNothingToGatherIsNothingAtAllRatherThanAnEmptySet(t *testing.T) {
	arguments := domains.NewToolArgumentsDomain(map[string]json.RawMessage{
		"symbol": json.RawMessage(`"BTCUSDT"`),
	})

	assert.Nil(t, arguments.EncodedSubset([]string{}))
	assert.Nil(t, arguments.EncodedSubset([]string{"notFilledIn"}))
}
