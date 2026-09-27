package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var everySpotTradeJournalAbility = []string{
	"trading_record_spot_trade",
	"trading_add_spot_trade_fill",
	"trading_update_spot_trade_fill",
	"trading_remove_spot_trade_fill",
	"trading_update_spot_trade_plan",
	"trading_add_spot_trade_note",
	"trading_write_spot_trade_review",
	"trading_set_spot_trade_setup_tags",
	"trading_list_spot_trades",
	"trading_get_spot_trade",
	"trading_delete_spot_trade",
	"trading_get_spot_trade_statistics",
	"trading_compare_spot_trades_with_backtest",
}

func TestEverySpotAbilityAsksTheRightWayAtTheRightAddress(t *testing.T) {
	testCases := []struct {
		abilityName  string
		expectedVerb vo.RequestVerb
		expectedPath string
	}{
		{"trading_record_spot_trade", vo.RequestVerbSubmit, "/spot-trade-records"},
		{"trading_add_spot_trade_fill", vo.RequestVerbSubmit, "/spot-trade-records/7/fills"},
		{"trading_update_spot_trade_fill", vo.RequestVerbReplace, "/spot-trade-records/7/fills/3"},
		{"trading_remove_spot_trade_fill", vo.RequestVerbRemove, "/spot-trade-records/7/fills/3"},
		{"trading_update_spot_trade_plan", vo.RequestVerbReplace, "/spot-trade-records/7/plan"},
		{"trading_add_spot_trade_note", vo.RequestVerbSubmit, "/spot-trade-records/7/notes"},
		{"trading_write_spot_trade_review", vo.RequestVerbReplace, "/spot-trade-records/7/review"},
		{"trading_set_spot_trade_setup_tags", vo.RequestVerbReplace, "/spot-trade-records/7/setup-tags"},
		{"trading_list_spot_trades", vo.RequestVerbRead, "/spot-trade-records"},
		{"trading_get_spot_trade", vo.RequestVerbRead, "/spot-trade-records/7"},
		{"trading_delete_spot_trade", vo.RequestVerbRemove, "/spot-trade-records/7"},
		{"trading_get_spot_trade_statistics", vo.RequestVerbRead, "/spot-trade-records/statistics"},
		{"trading_compare_spot_trades_with_backtest", vo.RequestVerbRead, "/trading-strategies/7/spot-trade-comparison"},
	}
	require.Len(t, testCases, len(everySpotTradeJournalAbility))

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			filledIn := everyBoxFilledIn()
			filledIn["id"] = json.RawMessage(`"7"`)
			filledIn["fillId"] = json.RawMessage(`"3"`)

			request := buildTradeJournalRequest(t, testCase.abilityName, filledIn)

			assert.Equal(t, testCase.expectedVerb, request.Verb)
			assert.Equal(t, testCase.expectedPath, request.Path)
		})
	}
}

func TestRecordingASpotTradeSendsTheFirstBuyNested(t *testing.T) {
	request := buildTradeJournalRequest(t, "trading_record_spot_trade", map[string]json.RawMessage{
		"symbol":       json.RawMessage(`"2330"`),
		"firstBuyFill": json.RawMessage(`{"price":"1050","quantity":"1000"}`),
		"plan":         json.RawMessage(`{"plannedStopLossPrice":"1000"}`),
	})

	assert.JSONEq(t,
		`{"symbol":"2330","firstBuyFill":{"price":"1050","quantity":"1000"},"plan":{"plannedStopLossPrice":"1000"}}`,
		string(request.Body))
}

func TestASpotTradeOffersNoDirectionAndNoLeverage(t *testing.T) {
	for _, abilityName := range everySpotTradeJournalAbility {
		for _, forbiddenBox := range []string{"direction", "leverage", "liquidity"} {
			_, isDeclared := boxNamed(abilityNamed(t, abilityName), forbiddenBox)
			assert.False(t, isDeclared, "%s 不該有 %s", abilityName, forbiddenBox)
		}
	}
}

func TestEachSpotAbilityRequiresOnlyWhatTheTradingServiceRequires(t *testing.T) {
	testCases := []struct {
		abilityName  string
		requiredness map[string]bool
	}{
		{"trading_record_spot_trade", map[string]bool{
			"symbol": true, "firstBuyFill": true, "plan": false, "tradingStrategyId": false, "setupTagIds": false,
		}},
		{"trading_add_spot_trade_fill", map[string]bool{
			"id": true, "kind": true, "price": true, "quantity": true, "filledAt": false, "fee": false,
		}},
		{"trading_update_spot_trade_fill", map[string]bool{
			"id": true, "fillId": true, "kind": true, "price": true, "quantity": true, "filledAt": false, "fee": false,
		}},
		{"trading_remove_spot_trade_fill", map[string]bool{"id": true, "fillId": true}},
		{"trading_update_spot_trade_plan", map[string]bool{
			"id": true, "plannedStopLossPrice": false, "plannedTakeProfitPrice": false, "entryReason": false, "confidence": false,
		}},
		{"trading_add_spot_trade_note", map[string]bool{"id": true, "content": true}},
		{"trading_write_spot_trade_review", map[string]bool{
			"id": true, "executionScore": true, "wentWell": false, "wentWrong": false, "nextTime": false, "mistakeTagIds": false,
		}},
		{"trading_set_spot_trade_setup_tags", map[string]bool{"id": true, "setupTagIds": true}},
		{"trading_list_spot_trades", map[string]bool{
			"status": false, "symbol": false, "market": false, "period": false, "limit": false,
		}},
		{"trading_get_spot_trade", map[string]bool{"id": true}},
		{"trading_delete_spot_trade", map[string]bool{"id": true}},
		{"trading_get_spot_trade_statistics", map[string]bool{"period": false}},
		{"trading_compare_spot_trades_with_backtest", map[string]bool{"id": true}},
	}
	require.Len(t, testCases, len(everySpotTradeJournalAbility))

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			ability := abilityNamed(t, testCase.abilityName)
			require.Len(t, ability.Parameters, len(testCase.requiredness), "宣告了沒被釘住的欄位")
			for boxName, isRequired := range testCase.requiredness {
				box, isDeclared := boxNamed(ability, boxName)

				require.True(t, isDeclared, boxName)
				assert.Equal(t, isRequired, box.IsRequired, boxName)
			}
		})
	}
}

func TestEverySpotAbilitySaysWhatWillGetItRefusedAndWhatToDo(t *testing.T) {
	testCases := []struct {
		abilityName     string
		requiredPhrases []string
	}{
		{"trading_record_spot_trade", []string{
			"只有先買後賣", "沒有做空、沒有槓桿", "不要送出", "先問現貨還是合約", "參考價不是成交價",
			"同一標的已有持有中會被拒絕", "trading_add_spot_trade_fill", "沒有欄位可以指定擁有者", "編號", "股數說出來請使用者確認",
			"多筆買進", "依序", "中途被拒就停下",
		}},
		{"trading_add_spot_trade_fill", []string{"賣出超過目前持有會被拒絕", "已平倉", "報酬率", "可以寫檢討", "已平倉的交易不能再加買賣", "中途被拒就停下"}},
		{"trading_update_spot_trade_fill", []string{"沒給的項目不會保留", "trading_get_spot_trade", "只有持有中的交易可以修正", "平倉後已鎖定", "刪除整筆重記"}},
		{"trading_remove_spot_trade_fill", []string{"持有中才可刪", "不能刪到沒有買進", "trading_delete_spot_trade"}},
		{"trading_update_spot_trade_plan", []string{"沒給的項目會被清空", "trading_get_spot_trade", "平倉後計畫已鎖定", "trading_add_spot_trade_note"}},
		{"trading_add_spot_trade_note", []string{"任何狀態都能加", "附註不會改動原本的計畫"}},
		{"trading_write_spot_trade_review", []string{"只能在平倉後寫", "持有中會被拒絕", "之後還能改", "沒給的項目會被清空", "與合約日誌共用"}},
		{"trading_set_spot_trade_setup_tags", []string{"整組取代", "空陣列即全部拿掉"}},
		{"trading_list_spot_trades", []string{"省略 limit 即最近 20 筆", "由新到舊", "說出總數", "status=closed", "market=taiwanStock"}},
		{"trading_get_spot_trade", []string{"沒有資金費用與強平價", "不要說成 0", "估算", "找不到"}},
		{"trading_delete_spot_trade", []string{"刪除前必須先向使用者確認是哪一筆", "先列出請他指定", "不要猜", "一併刪除"}},
		{"trading_get_spot_trade_statistics", []string{"台股與加密貨幣分成兩組", "金額不跨幣別加總", "以幾筆計", "不是 0%", "只計期間內平倉"}},
		{"trading_compare_spot_trades_with_backtest", []string{"重演不計成本", "需要時間", "實盤照常", "策略已刪除時說無法重演", "不要立刻重送"}},
	}
	require.Len(t, testCases, len(everySpotTradeJournalAbility))

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			description := abilityNamed(t, testCase.abilityName).Description
			for _, phrase := range testCase.requiredPhrases {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestSpotQuantitiesPricesAndFeesSayWhatTheAssistantMustKnow(t *testing.T) {
	for _, abilityName := range []string{"trading_add_spot_trade_fill", "trading_update_spot_trade_fill"} {
		for _, phrase := range []string{"台股以股計、必須是整數", "1,000 股", "請他確認"} {
			assert.Contains(t, boxDescription(t, abilityName, "quantity"), phrase, abilityName)
		}
		for _, phrase := range []string{"沒說即 0", "證交稅要併入手續費"} {
			assert.Contains(t, boxDescription(t, abilityName, "fee"), phrase, abilityName)
		}
		assert.Contains(t, boxDescription(t, abilityName, "price"), "參考價不是成交價", abilityName)
		for _, phrase := range []string{"省略即交易服務以現在記下", "台北時間", "RFC3339"} {
			assert.Contains(t, boxDescription(t, abilityName, "filledAt"), phrase, abilityName)
		}
	}

	firstBuyFill := boxDescription(t, "trading_record_spot_trade", "firstBuyFill")
	for _, phrase := range []string{"台股以股計、必須是整數", "沒說即 0", "參考價不是成交價", "省略即交易服務以現在記下", "kind 省略即買進"} {
		assert.Contains(t, firstBuyFill, phrase)
	}
}

func TestASpotPlanSaysWhichSideTheStopBelongsOn(t *testing.T) {
	testCases := []struct {
		abilityName string
		boxName     string
	}{
		{"trading_update_spot_trade_plan", "plannedStopLossPrice"},
		{"trading_record_spot_trade", "plan"},
	}

	for _, testCase := range testCases {
		assert.Contains(t, boxDescription(t, testCase.abilityName, testCase.boxName), "止損必須低於第一筆買進價", testCase.abilityName)
	}
}

func TestASpotTradeFollowsOnlyASpotStrategy(t *testing.T) {
	request := buildTradeJournalRequest(t, "trading_record_spot_trade", map[string]json.RawMessage{
		"symbol":            json.RawMessage(`"2330"`),
		"firstBuyFill":      json.RawMessage(`{"price":"1050","quantity":"1000"}`),
		"tradingStrategyId": json.RawMessage(`5`),
	})

	assert.Contains(t, string(request.Body), `"tradingStrategyId":5`)
}

func TestASpotStrategyBoxSaysOnlyASpotStrategyFits(t *testing.T) {
	for _, phrase := range []string{"只能是使用者自己的 K 線（現貨）交易策略", "合約交易策略會被拒絕", "不給即自行判斷"} {
		assert.Contains(t, boxDescription(t, "trading_record_spot_trade", "tradingStrategyId"), phrase)
	}
}

func TestSpotDefaultsAreLeftToTheTradingService(t *testing.T) {
	testCases := []struct {
		name           string
		abilityName    string
		givenBoxes     map[string]json.RawMessage
		expectedQuery  map[string]string
		absentFromBody []string
	}{
		{"listing sends no query of its own", "trading_list_spot_trades", map[string]json.RawMessage{}, map[string]string{}, nil},
		{"listing forwards only what was given", "trading_list_spot_trades", map[string]json.RawMessage{
			"market": json.RawMessage(`"taiwanStock"`),
			"status": json.RawMessage(`"closed"`),
		}, map[string]string{"market": "taiwanStock", "status": "closed"}, nil},
		{"statistics leave the period to the service", "trading_get_spot_trade_statistics", map[string]json.RawMessage{}, map[string]string{}, nil},
		{"a sell without fee or time leaves both to the service", "trading_add_spot_trade_fill", map[string]json.RawMessage{
			"id":       json.RawMessage(`"41"`),
			"kind":     json.RawMessage(`"sell"`),
			"price":    json.RawMessage(`"1120"`),
			"quantity": json.RawMessage(`"600"`),
		}, map[string]string{}, []string{"fee", "filledAt"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := buildTradeJournalRequest(t, testCase.abilityName, testCase.givenBoxes)

			if len(testCase.expectedQuery) == 0 {
				assert.Empty(t, request.Query)
			} else {
				assert.Equal(t, testCase.expectedQuery, request.Query)
			}
			for _, absentKey := range testCase.absentFromBody {
				assert.NotContains(t, string(request.Body), absentKey)
			}
		})
	}
	assert.Contains(t, boxDescription(t, "trading_get_spot_trade_statistics", "period"), "省略即最近 30 天")
}

func TestNoSpotAbilityLetsTheAssistantNameAnOwner(t *testing.T) {
	for _, abilityName := range everySpotTradeJournalAbility {
		for _, parameter := range abilityNamed(t, abilityName).Parameters {
			assert.NotContains(t, strings.ToLower(parameter.Name), "owner", abilityName)
			assert.NotContains(t, strings.ToLower(parameter.Name), "user", abilityName)
		}
	}
}

func TestOnlyTheSpotBacktestComparisonWaitsLikeAReplay(t *testing.T) {
	for _, abilityName := range everySpotTradeJournalAbility {
		filledIn := everyBoxFilledIn()
		filledIn["id"] = json.RawMessage(`"7"`)
		filledIn["fillId"] = json.RawMessage(`"3"`)

		request := buildTradeJournalRequest(t, abilityName, filledIn)

		if abilityName == "trading_compare_spot_trades_with_backtest" {
			assert.Equal(t, 120*time.Second, request.ResponseWaitLimit)
			continue
		}
		assert.Zero(t, request.ResponseWaitLimit, abilityName)
	}
}

func TestAFeeTheUserGaveIsForwardedAsGiven(t *testing.T) {
	request := buildTradeJournalRequest(t, "trading_add_spot_trade_fill", map[string]json.RawMessage{
		"id":       json.RawMessage(`"41"`),
		"kind":     json.RawMessage(`"sell"`),
		"price":    json.RawMessage(`"1120"`),
		"quantity": json.RawMessage(`"600"`),
		"fee":      json.RawMessage(`"1983"`),
	})

	assert.JSONEq(t, `{"kind":"sell","price":"1120","quantity":"600","fee":"1983"}`, string(request.Body))
}
