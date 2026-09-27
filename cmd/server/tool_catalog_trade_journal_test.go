package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var everyTradeJournalAbility = []string{
	"trading_record_contract_trade",
	"trading_add_contract_trade_fill",
	"trading_update_contract_trade_fill",
	"trading_remove_contract_trade_fill",
	"trading_update_contract_trade_plan",
	"trading_add_contract_trade_note",
	"trading_write_contract_trade_review",
	"trading_set_contract_trade_setup_tags",
	"trading_list_contract_trades",
	"trading_get_contract_trade",
	"trading_delete_contract_trade",
	"trading_get_contract_trade_statistics",
	"trading_compare_contract_trades_with_backtest",
	"trading_get_trade_journal_settings",
	"trading_save_trade_journal_settings",
	"trading_list_trade_tags",
	"trading_create_trade_tag",
	"trading_rename_trade_tag",
	"trading_delete_trade_tag",
}

func buildTradeJournalRequest(
	t *testing.T,
	abilityName string,
	boxes map[string]json.RawMessage,
) vo.TradingServiceRequestVo {
	t.Helper()

	request, buildError := apiToolNamed(t, abilityName).BuildRequest(domains.NewToolArgumentsDomain(boxes))
	require.NoError(t, buildError)

	return request
}

func TestTheTagAndFeeRateAbilitiesAskTheRightWayAtTheRightAddress(t *testing.T) {
	testCases := []struct {
		abilityName  string
		expectedVerb vo.RequestVerb
		expectedPath string
	}{
		{"trading_get_trade_journal_settings", vo.RequestVerbRead, "/users/me/trade-journal-settings"},
		{"trading_save_trade_journal_settings", vo.RequestVerbReplace, "/users/me/trade-journal-settings"},
		{"trading_list_trade_tags", vo.RequestVerbRead, "/users/me/trade-tags"},
		{"trading_create_trade_tag", vo.RequestVerbSubmit, "/users/me/trade-tags"},
		{"trading_rename_trade_tag", vo.RequestVerbReplace, "/users/me/trade-tags/7"},
		{"trading_delete_trade_tag", vo.RequestVerbRemove, "/users/me/trade-tags/7"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			filledIn := everyBoxFilledIn()
			filledIn["id"] = json.RawMessage(`"7"`)

			request := buildTradeJournalRequest(t, testCase.abilityName, filledIn)

			assert.Equal(t, testCase.expectedVerb, request.Verb)
			assert.Equal(t, testCase.expectedPath, request.Path)
		})
	}
}

func TestRecordingATradeSendsTheFirstEntryFillNested(t *testing.T) {
	request := buildTradeJournalRequest(t, "trading_record_contract_trade", map[string]json.RawMessage{
		"symbol":         json.RawMessage(`"BTCUSDT"`),
		"direction":      json.RawMessage(`"long"`),
		"leverage":       json.RawMessage(`"10"`),
		"firstEntryFill": json.RawMessage(`{"kind":"entry","price":"97905","quantity":"0.03"}`),
	})

	assert.JSONEq(t,
		`{"symbol":"BTCUSDT","direction":"long","leverage":"10","firstEntryFill":{"kind":"entry","price":"97905","quantity":"0.03"}}`,
		string(request.Body))
}

func TestRecordingATradeRequiresOnlyWhatTheTradingServiceRequires(t *testing.T) {
	ability := abilityNamed(t, "trading_record_contract_trade")

	for boxName, isRequired := range map[string]bool{
		"symbol": true, "direction": true, "firstEntryFill": true,
		"leverage": false, "tradingStrategyId": false, "setupTagIds": false,
		"plan": false,
	} {
		box, isDeclared := boxNamed(ability, boxName)

		require.True(t, isDeclared, boxName)
		assert.Equal(t, isRequired, box.IsRequired, boxName)
	}
}

func TestEachAbilityRequiresOnlyWhatTheTradingServiceRequires(t *testing.T) {
	testCases := []struct {
		abilityName  string
		requiredness map[string]bool
	}{
		{"trading_add_contract_trade_fill", map[string]bool{
			"id": true, "kind": true, "price": true, "quantity": true, "filledAt": false, "liquidity": false, "fee": false,
		}},
		{"trading_update_contract_trade_fill", map[string]bool{"id": true, "fillId": true, "kind": true, "price": true, "quantity": true}},
		{"trading_update_contract_trade_plan", map[string]bool{
			"id": true, "plannedStopLossPrice": false, "plannedTakeProfitPrice": false, "entryReason": false, "confidence": false,
		}},
		{"trading_add_contract_trade_note", map[string]bool{"id": true, "content": true}},
		{"trading_write_contract_trade_review", map[string]bool{
			"id": true, "executionScore": true, "wentWell": false, "wentWrong": false, "nextTime": false, "mistakeTagIds": false,
		}},
		{"trading_set_contract_trade_setup_tags", map[string]bool{"id": true, "setupTagIds": true}},
		{"trading_list_contract_trades", map[string]bool{"status": false, "symbol": false, "period": false, "limit": false}},
		{"trading_get_contract_trade_statistics", map[string]bool{"period": false}},
		{"trading_save_trade_journal_settings", map[string]bool{"makerFeeRate": false, "takerFeeRate": false}},
		{"trading_create_trade_tag", map[string]bool{"kind": true, "name": true}},
		{"trading_rename_trade_tag", map[string]bool{"id": true, "name": true}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			ability := abilityNamed(t, testCase.abilityName)
			for boxName, isRequired := range testCase.requiredness {
				box, isDeclared := boxNamed(ability, boxName)

				require.True(t, isDeclared, boxName)
				assert.Equal(t, isRequired, box.IsRequired, boxName)
			}
		})
	}
}

func TestRecordingATradeTellsTheAssistantWhatToReportBack(t *testing.T) {
	description := abilityNamed(t, "trading_record_contract_trade").Description

	for _, phrase := range []string{"編號", "開倉均價", "計畫風險", "記下的時間"} {
		assert.Contains(t, description, phrase)
	}
}

func TestRecordingATradeForwardsTheTradingStrategyItFollowed(t *testing.T) {
	request := buildTradeJournalRequest(t, "trading_record_contract_trade", map[string]json.RawMessage{
		"symbol":            json.RawMessage(`"BTCUSDT"`),
		"direction":         json.RawMessage(`"long"`),
		"firstEntryFill":    json.RawMessage(`{"price":"97905","quantity":"0.03"}`),
		"tradingStrategyId": json.RawMessage(`5`),
	})

	assert.Contains(t, string(request.Body), `"tradingStrategyId":5`)
	for _, phrase := range []string{"只能是使用者自己的合約交易策略", "K 線交易策略會被拒絕", "不給即自行判斷"} {
		assert.Contains(t, boxDescription(t, "trading_record_contract_trade", "tradingStrategyId"), phrase)
	}
}

func TestLeavingOutTheFillTimeLeavesItToTheTradingService(t *testing.T) {
	request := buildTradeJournalRequest(t, "trading_add_contract_trade_fill", map[string]json.RawMessage{
		"id":       json.RawMessage(`"27"`),
		"kind":     json.RawMessage(`"entry"`),
		"price":    json.RawMessage(`"97960"`),
		"quantity": json.RawMessage(`"0.021"`),
	})

	assert.NotContains(t, string(request.Body), "filledAt")
	assert.NotContains(t, string(request.Body), "fee")
}

func TestGivenFillTimeIsForwarded(t *testing.T) {
	request := buildTradeJournalRequest(t, "trading_add_contract_trade_fill", map[string]json.RawMessage{
		"id":       json.RawMessage(`"27"`),
		"kind":     json.RawMessage(`"exit"`),
		"filledAt": json.RawMessage(`"2026-09-26T16:40:00+08:00"`),
		"price":    json.RawMessage(`"100420"`),
		"quantity": json.RawMessage(`"0.051"`),
	})

	assert.Contains(t, string(request.Body), `"filledAt":"2026-09-26T16:40:00+08:00"`)
}

func TestTheFillTimeSaysNowIsTheDefaultAndHowToReportIt(t *testing.T) {
	for _, abilityName := range []string{"trading_add_contract_trade_fill", "trading_update_contract_trade_fill"} {
		filledAt, isDeclared := boxNamed(abilityNamed(t, abilityName), "filledAt")

		require.True(t, isDeclared, abilityName)
		assert.False(t, filledAt.IsRequired, abilityName)
		for _, phrase := range []string{"省略即交易服務以現在記下", "台北時間", "RFC3339"} {
			assert.Contains(t, filledAt.Description, phrase, abilityName)
		}
	}
	assert.Contains(t, boxDescription(t, "trading_record_contract_trade", "firstEntryFill"), "省略即交易服務以現在記下")
}

func TestTheOpeningPriceWarnsThatAReferencePriceIsNotOne(t *testing.T) {
	for _, abilityName := range []string{"trading_add_contract_trade_fill", "trading_update_contract_trade_fill"} {
		assert.Contains(t, boxDescription(t, abilityName, "price"), "參考價不是開倉價", abilityName)
		assert.Contains(t, boxDescription(t, abilityName, "price"), "先問他", abilityName)
	}
	assert.Contains(t, boxDescription(t, "trading_record_contract_trade", "firstEntryFill"), "參考價不是開倉價")
	assert.Contains(t, abilityNamed(t, "trading_record_contract_trade").Description, "參考價不是開倉價")
}

func TestEveryAbilitySaysWhatWillGetItRefused(t *testing.T) {
	testCases := []struct {
		abilityName     string
		requiredPhrases []string
	}{
		{"trading_record_contract_trade", []string{
			"同一標的、同一方向已有持倉中會被拒絕", "trading_add_contract_trade_fill", "沒有欄位可以指定擁有者",
			"依序", "中途被拒就停下",
		}},
		{"trading_add_contract_trade_fill", []string{"加倉（entry）", "減倉或平倉（exit）", "減倉超過目前持倉會被拒絕", "先平倉", "可以寫檢討", "已平倉的交易不能再加倉或減倉"}},
		{"trading_update_contract_trade_fill", []string{"只有持倉中的交易可以修正", "平倉後紀錄已鎖定", "加附註", "刪除整筆重記", "沒給的項目不會保留", "trading_get_contract_trade"}},
		{"trading_remove_contract_trade_fill", []string{"持倉中才可刪", "不能刪到沒有開倉紀錄", "trading_delete_contract_trade"}},
		{"trading_update_contract_trade_plan", []string{"平倉後計畫已鎖定", "trading_add_contract_trade_note", "沒給的項目會被清空", "trading_get_contract_trade"}},
		{"trading_add_contract_trade_note", []string{"任何狀態都能加", "附註不會改動原本的計畫"}},
		{"trading_write_contract_trade_review", []string{"只能在平倉後寫", "持倉中會被拒絕", "之後還能改", "沒給的項目會被清空", "trading_get_contract_trade"}},
		{"trading_set_contract_trade_setup_tags", []string{"整組取代", "空陣列即全部拿掉"}},
		{"trading_list_contract_trades", []string{"省略 limit 即最近 20 筆", "由新到舊", "說出總數", "status=closed"}},
		{"trading_get_contract_trade", []string{"不要說成 0", "找不到"}},
		{"trading_delete_contract_trade", []string{"刪除前必須先向使用者確認是哪一筆", "先列出請他指定", "不要猜", "一併刪除"}},
		{"trading_get_contract_trade_statistics", []string{"只計期間內平倉", "不是 0%", "沒設止損的交易不計入 R"}},
		{"trading_compare_contract_trades_with_backtest", []string{"需要時間", "實盤照常", "策略已刪除時說無法重演", "不要立刻重送"}},
		{"trading_get_trade_journal_settings", []string{"未設定"}},
		{"trading_save_trade_journal_settings", []string{"不得為負", "只用於合約日誌", "舊紀錄的手續費不變", "沒給的那一個會被清空", "trading_get_trade_journal_settings"}},
		{"trading_list_trade_tags", []string{"五個預設失誤標籤", "兩本日誌共用同一組"}},
		{"trading_create_trade_tag", []string{"同一類不能重名", "不同類可以"}},
		{"trading_rename_trade_tag", []string{"跟著改名"}},
		{"trading_delete_trade_tag", []string{"還貼在交易上的標籤不能刪", "現貨與合約的交易一起算", "說出還有幾筆"}},
	}
	require.Len(t, testCases, len(everyTradeJournalAbility))

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			description := abilityNamed(t, testCase.abilityName).Description
			for _, phrase := range testCase.requiredPhrases {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestTheDefaultsTheDescriptionsPromiseAreLeftToTheTradingService(t *testing.T) {
	listRequest := buildTradeJournalRequest(t, "trading_list_contract_trades", map[string]json.RawMessage{})
	assert.Empty(t, listRequest.Query)

	givenLimit := buildTradeJournalRequest(t, "trading_list_contract_trades", map[string]json.RawMessage{
		"status": json.RawMessage(`"closed"`),
		"limit":  json.RawMessage(`50`),
	})
	assert.Equal(t, map[string]string{"status": "closed", "limit": "50"}, givenLimit.Query)

	statisticsRequest := buildTradeJournalRequest(t, "trading_get_contract_trade_statistics", map[string]json.RawMessage{})
	assert.Empty(t, statisticsRequest.Query)
	assert.Contains(t, boxDescription(t, "trading_get_contract_trade_statistics", "period"), "省略即最近 30 天")
}

func TestNoTradeJournalAbilityLetsTheAssistantNameAnOwner(t *testing.T) {
	for _, abilityName := range everyTradeJournalAbility {
		for _, parameter := range abilityNamed(t, abilityName).Parameters {
			assert.NotContains(t, strings.ToLower(parameter.Name), "owner", abilityName)
			assert.NotContains(t, strings.ToLower(parameter.Name), "user", abilityName)
		}
	}
}

func TestOnlyTheBacktestComparisonWaitsLikeAReplay(t *testing.T) {
	for _, abilityName := range everyTradeJournalAbility {
		filledIn := everyBoxFilledIn()
		filledIn["id"] = json.RawMessage(`"7"`)

		request := buildTradeJournalRequest(t, abilityName, filledIn)

		if abilityName == "trading_compare_contract_trades_with_backtest" {
			assert.Equal(t, 120*time.Second, request.ResponseWaitLimit)
			continue
		}
		assert.Zero(t, request.ResponseWaitLimit, abilityName)
	}
}

func boxDescription(t *testing.T, abilityName string, boxName string) string {
	t.Helper()

	box, isDeclared := boxNamed(abilityNamed(t, abilityName), boxName)
	require.True(t, isDeclared, "%s 沒有 %s", abilityName, boxName)

	return box.Description
}

func TestRecordingATradeSaysAnUnlabelledFirstRecordOpensThePosition(t *testing.T) {
	description := boxDescription(t, "trading_record_contract_trade", "firstEntryFill")

	assert.Contains(t, description, `"kind":"entry"`)
	assert.Contains(t, description, "kind 省略即開倉")
}

func TestNoContractJournalAbilitySpeaksOfFills(t *testing.T) {
	for _, abilityName := range everyTradeJournalAbility {
		definition := abilityNamed(t, abilityName)
		assert.NotContains(t, definition.Description, "成交", abilityName)
		for _, parameter := range definition.Parameters {
			assert.NotContains(t, parameter.Description, "成交", "%s 的 %s", abilityName, parameter.Name)
		}
	}
}

func TestContractRecordsAreCalledByTheirPositionNames(t *testing.T) {
	kind := boxDescription(t, "trading_add_contract_trade_fill", "kind")
	for _, phrase := range []string{"開倉", "加倉", "減倉", "平倉"} {
		assert.Contains(t, kind, phrase)
	}
	assert.Contains(t, boxDescription(t, "trading_add_contract_trade_fill", "price"), "開倉價或平倉價")
}

func TestLeavingLiquidityOutIsSaidToMeanTaker(t *testing.T) {
	assert.Contains(t, boxDescription(t, "trading_add_contract_trade_fill", "liquidity"), "省略即吃單")
	assert.Contains(t, boxDescription(t, "trading_record_contract_trade", "firstEntryFill"), "liquidity 省略即吃單")
}
