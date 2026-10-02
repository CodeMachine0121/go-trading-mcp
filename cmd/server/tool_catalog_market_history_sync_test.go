package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A history sync that finished with little stored must not read as broken, and the
// closed days it skipped must not be mistaken for skipped candles.
func TestTheSpotHistorySyncAbilitiesSayWhatAPresumedClosedDayIs(t *testing.T) {
	testCases := []struct {
		abilityName string
		mustSay     []string
		mustNotSay  []string
	}{
		{abilityName: "trading_sync_k_candle_history",
			mustSay: []string{
				"落在平日的國定假日會被跳過、記在 presumedClosedDayCount，不會讓同步停下",
				"例外是連續 15 個交易日都沒資料：那比任何休市都長，多半是來源不認得這個代號，同步會停下、原因寫在 fetchFailureReason。",
			}},
		{abilityName: "trading_get_k_candle_history_sync",
			mustSay: []string{
				"storedCount、skippedCount、presumedClosedDayCount。",
				"presumedClosedDayCount 是被跳過的平日數：來源說那一天沒有這個標的的資料",
				"那一天跳過、繼續問下一天，不算來源不答話。",
				"它與 skippedCount 是兩件事：skippedCount 是來源答了、但某幾根不合格。",
			}},
		// Contracts trade round the clock, so they have no closed day to skip.
		{abilityName: "trading_sync_contract_k_candle_history",
			mustNotSay: []string{"presumedClosedDayCount"}},
		{abilityName: "trading_get_contract_k_candle_history_sync",
			mustNotSay: []string{"presumedClosedDayCount"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			ability := abilityNamed(t, testCase.abilityName)
			described := ability.Description
			for _, parameter := range ability.Parameters {
				described += parameter.Description
			}

			for _, phrase := range testCase.mustSay {
				assert.Contains(t, described, phrase)
			}
			for _, phrase := range testCase.mustNotSay {
				assert.NotContains(t, described, phrase)
			}
		})
	}
}
