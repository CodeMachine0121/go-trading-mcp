package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// everyContractAbility is the contract line of the catalogue, by name.
var everyContractAbility = []string{
	"trading_create_contract_k_candle",
	"trading_list_contract_k_candles",
	"trading_get_contract_k_candle_series",
	"trading_get_contract_k_candle",
	"trading_update_contract_k_candle",
	"trading_delete_contract_k_candle",
	"trading_backfill_contract_k_candles",
	"trading_sync_contract_k_candle_history",
	"trading_get_contract_k_candle_history_sync",
	"trading_list_contract_trading_symbols",
	"trading_add_to_contract_watchlist",
	"trading_remove_from_contract_watchlist",
	"trading_list_contract_funding_rate_settlements",
	"trading_list_contract_position_statistics",
	"trading_get_contract_maintenance_margin_tiers",
}

// contractCandleFigureNames are every figure a contract candle carries.
var contractCandleFigureNames = []string{
	"open", "high", "low", "close", "volume", "quoteVolume", "takerBuyBaseVolume", "takerBuyQuoteVolume",
	"tradeCount", "markOpen", "markHigh", "markLow", "markClose",
	"indexOpen", "indexHigh", "indexLow", "indexClose",
	"premiumIndexOpen", "premiumIndexHigh", "premiumIndexLow", "premiumIndexClose",
}

// everyBoxFilledIn is a filled-in value for every box any ability declares, so that
// every ability builds a request and none is skipped for a box left empty.
func everyBoxFilledIn() map[string]json.RawMessage {
	filledIn := map[string]json.RawMessage{}
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		for _, parameter := range apiTool.ToDefinitionDto().Parameters {
			switch parameter.Kind {
			case "integer":
				filledIn[parameter.Name] = json.RawMessage(`1`)
			case "boolean":
				filledIn[parameter.Name] = json.RawMessage(`true`)
			case "array":
				filledIn[parameter.Name] = json.RawMessage(`[]`)
			case "object":
				filledIn[parameter.Name] = json.RawMessage(`{}`)
			default:
				filledIn[parameter.Name] = json.RawMessage(`"1"`)
			}
		}
	}

	return filledIn
}

// A contract ability never reaches a spot address, and a spot one never reaches a
// contract address — and the names say which is which. The same symbol names two
// instruments; asking the wrong side answers with plausible numbers for the wrong
// thing and no refusal.
func TestContractAbilitiesOnlyEverReachTheContractLine(t *testing.T) {
	isContractAbility := map[string]bool{}
	for _, abilityName := range everyContractAbility {
		isContractAbility[abilityName] = true
	}

	checkedContractAbilities := 0
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		request, buildError := apiTool.BuildRequest(domains.NewToolArgumentsDomain(everyBoxFilledIn()))
		require.NoError(t, buildError, apiTool.Name())

		assert.Equal(t, isContractAbility[apiTool.Name()], strings.HasPrefix(request.Path, "/contract-"),
			"%s 走到了 %s", apiTool.Name(), request.Path)
		assert.Equal(t, isContractAbility[apiTool.Name()], strings.Contains(apiTool.Name(), "contract"),
			"合約的能力名字要帶 contract，現貨的不帶：%s", apiTool.Name())
		if isContractAbility[apiTool.Name()] {
			checkedContractAbilities++
		}
	}
	assert.Equal(t, len(everyContractAbility), checkedContractAbilities)
}

// Writing a contract candle asks for every figure, the three price lines included: a
// contract candle missing any of them is refused, and an optional box is one an
// assistant leaves out.
func TestWritingAContractCandleAsksForEveryFigure(t *testing.T) {
	for _, abilityName := range []string{"trading_create_contract_k_candle", "trading_update_contract_k_candle"} {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)
			for _, boxName := range contractCandleFigureNames {
				box, isDeclared := boxNamed(ability, boxName)

				require.True(t, isDeclared, boxName)
				assert.True(t, box.IsRequired, "合約 K 線每一項都必填：%s", boxName)
			}
		})
	}
}

func filledInContractCandle() map[string]json.RawMessage {
	filledIn := map[string]json.RawMessage{
		"symbol":   json.RawMessage(`"BTCUSDT"`),
		"openTime": json.RawMessage(`"2026-09-23T08:00:00Z"`),
	}
	for _, figureName := range contractCandleFigureNames {
		filledIn[figureName] = json.RawMessage(`"1"`)
	}
	filledIn["tradeCount"] = json.RawMessage(`0`)
	filledIn["premiumIndexClose"] = json.RawMessage(`"-0.0005"`)

	return filledIn
}

// A filled-in contract candle leaves the connector whole: the negative premium and a
// trade count of zero go out as they were given.
func TestAWrittenContractCandleCarriesItsPriceLines(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_create_contract_k_candle").
		BuildRequest(domains.NewToolArgumentsDomain(filledInContractCandle()))

	require.NoError(t, buildError)
	assert.Equal(t, "/contract-k-candles", request.Path)
	assert.Contains(t, string(request.Body), `"premiumIndexClose":"-0.0005"`)
	assert.Contains(t, string(request.Body), `"tradeCount":0`)
	for _, figureName := range contractCandleFigureNames {
		assert.Contains(t, string(request.Body), `"`+figureName+`":`)
	}
}

// A line left out never reaches the trading service: the connector stops it and says
// which box is missing.
func TestAContractCandleMissingALineNeverLeavesTheConnector(t *testing.T) {
	testCases := []struct {
		abilityName string
		leftOut     string
	}{
		{abilityName: "trading_create_contract_k_candle", leftOut: "indexOpen"},
		{abilityName: "trading_update_contract_k_candle", leftOut: "premiumIndexClose"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			filledIn := filledInContractCandle()
			delete(filledIn, testCase.leftOut)

			_, buildError := apiToolNamed(t, testCase.abilityName).
				BuildRequest(domains.NewToolArgumentsDomain(filledIn))

			assert.ErrorContains(t, buildError, testCase.leftOut)
		})
	}
}

// Each contract series is asked for one contract over one stretch, and a question
// without the contract never leaves.
func TestTheContractSeriesAreAskedForOneContractOverOneStretch(t *testing.T) {
	testCases := []struct {
		abilityName  string
		expectedPath string
		takesStretch bool
	}{
		{abilityName: "trading_list_contract_funding_rate_settlements",
			expectedPath: "/contract-funding-rate-settlements", takesStretch: true},
		{abilityName: "trading_list_contract_position_statistics",
			expectedPath: "/contract-position-statistics", takesStretch: true},
		{abilityName: "trading_get_contract_maintenance_margin_tiers",
			expectedPath: "/contract-maintenance-margin-tiers"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			filledIn := map[string]json.RawMessage{
				"symbol":    json.RawMessage(`"BTCUSDT"`),
				"startTime": json.RawMessage(`"2026-09-21T08:00:00Z"`),
				"endTime":   json.RawMessage(`"2026-09-23T08:00:00Z"`),
			}

			request, buildError := apiToolNamed(t, testCase.abilityName).
				BuildRequest(domains.NewToolArgumentsDomain(filledIn))
			delete(filledIn, "symbol")
			_, withoutSymbolError := apiToolNamed(t, testCase.abilityName).
				BuildRequest(domains.NewToolArgumentsDomain(filledIn))

			require.NoError(t, buildError)
			assert.Equal(t, testCase.expectedPath, request.Path)
			assert.Equal(t, "BTCUSDT", request.Query["symbol"])
			if testCase.takesStretch {
				assert.Equal(t, "2026-09-21T08:00:00Z", request.Query["startTime"])
				assert.Equal(t, "2026-09-23T08:00:00Z", request.Query["endTime"])
			}
			assert.ErrorContains(t, withoutSymbolError, "symbol")
		})
	}
}

func TestJoiningTheContractWatchlistNamesTheContract(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_add_to_contract_watchlist").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"symbol": json.RawMessage(`"BTCUSDT"`),
		}))

	require.NoError(t, buildError)
	assert.Equal(t, "/contract-watchlist", request.Path)
	assert.JSONEq(t, `{"symbol":"BTCUSDT"}`, string(request.Body))
}

// What an assistant cannot find out by sending, said where it reads before sending.
func TestTheContractAbilitiesSayWhatCannotBeDiscoveredBySending(t *testing.T) {
	testCases := []struct {
		abilityName string
		mustSay     []string
	}{
		// Who pays whom is the one fact about a funding rate that flips every result.
		{abilityName: "trading_list_contract_funding_rate_settlements",
			mustSay: []string{"費率為正時做多的人付給做空的人", "不要自行取整"}},
		// An empty past is not a fault: nobody recorded it.
		{abilityName: "trading_list_contract_position_statistics", mustSay: []string{"只留最近三十天"}},
		// An empty ladder is not a fault either: no account key.
		{abilityName: "trading_get_contract_maintenance_margin_tiers", mustSay: []string{"帳戶金鑰", "回空陣列", "不是錯誤"}},
		// null on an old candle is not zero, and there is a way to fill it in.
		{abilityName: "trading_list_contract_k_candles",
			mustSay: []string{"舊資料", "trading_sync_contract_k_candle_history"}},
		// The specification's rate is only the smallest tier.
		{abilityName: "trading_list_contract_trading_symbols",
			mustSay: []string{"最小那一級", "trading_get_contract_maintenance_margin_tiers"}},
		// A slow add is not a hung one, and the two venues' codes differ.
		{abilityName: "trading_add_to_contract_watchlist", mustSay: []string{"二十秒", "1000SHIBUSDT"}},
		// Contract runs are numbered apart from spot ones.
		{abilityName: "trading_get_contract_k_candle_history_sync", mustSay: []string{"合約自己那一串"}},
		// Leaving the watchlist loses nothing.
		{abilityName: "trading_remove_from_contract_watchlist",
			mustSay: []string{"只停止追蹤", "一筆都不刪", "現貨那邊完全不受影響"}},
		// Two ways of asking for an interval cannot both be given.
		{abilityName: "trading_get_contract_k_candle_series",
			mustSay: []string{"interval 與 displayableCandleCount 兩者只能給一個", "null"}},
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
		})
	}
}

// Asking for the candles of one stretch, on each side, reaches that side's own
// address with the contract and the stretch — not merely "not the other side".
func TestReadingCandlesAsksTheRightLineForTheRightStretch(t *testing.T) {
	testCases := []struct {
		abilityName  string
		expectedPath string
	}{
		{abilityName: "trading_list_contract_k_candles", expectedPath: "/contract-k-candles"},
		{abilityName: "trading_get_contract_k_candle_series", expectedPath: "/contract-k-candles/series"},
		{abilityName: "trading_list_k_candles", expectedPath: "/k-candles"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, testCase.abilityName).
				BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
					"symbol":    json.RawMessage(`"BTCUSDT"`),
					"startTime": json.RawMessage(`"2026-09-23T08:00:00Z"`),
					"endTime":   json.RawMessage(`"2026-09-23T09:00:00Z"`),
				}))

			require.NoError(t, buildError)
			assert.Equal(t, testCase.expectedPath, request.Path)
			assert.Equal(t, "BTCUSDT", request.Query["symbol"])
			assert.Equal(t, "2026-09-23T08:00:00Z", request.Query["startTime"])
			assert.Equal(t, "2026-09-23T09:00:00Z", request.Query["endTime"])
		})
	}
}

// The series takes the spot series' two ways of asking for an interval, and hands
// on whichever one was given.
func TestTheContractSeriesAsksForAnIntervalTheWayTheSpotOneDoes(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_get_contract_k_candle_series").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"symbol":                 json.RawMessage(`"BTCUSDT"`),
			"startTime":              json.RawMessage(`"2026-09-16T00:00:00Z"`),
			"endTime":                json.RawMessage(`"2026-09-23T00:00:00Z"`),
			"displayableCandleCount": json.RawMessage(`100`),
		}))

	require.NoError(t, buildError)
	assert.Equal(t, "100", request.Query["displayableCandleCount"])
	for _, boxName := range []string{"interval", "displayableCandleCount"} {
		box, isDeclared := boxNamed(abilityNamed(t, "trading_get_contract_k_candle_series"), boxName)
		require.True(t, isDeclared, boxName)
		assert.False(t, box.IsRequired, "兩種說法都不給時由系統挑，所以兩格都不是必填：%s", boxName)
	}
}

// Every contract ability that reads data says what will get it refused, so an
// assistant learns it before sending rather than by being refused.
func TestEveryContractReadSaysWhatWillGetItRefused(t *testing.T) {
	testCases := []struct {
		abilityName string
		mustSay     []string
	}{
		{abilityName: "trading_list_contract_funding_rate_settlements",
			mustSay: []string{"會被拒絕的情況", "結束早於開始", "單次筆數上限"}},
		{abilityName: "trading_list_contract_position_statistics",
			mustSay: []string{"會被拒絕的情況", "結束早於開始", "單次筆數上限"}},
		{abilityName: "trading_list_contract_k_candles", mustSay: []string{"結束早於開始", "上限"}},
		{abilityName: "trading_get_contract_k_candle_series", mustSay: []string{"超過單次上限會被拒絕"}},
		{abilityName: "trading_delete_contract_k_candle", mustSay: []string{"404"}},
		{abilityName: "trading_get_contract_k_candle", mustSay: []string{"404"}},
		{abilityName: "trading_get_contract_maintenance_margin_tiers", mustSay: []string{"代號留白會被拒絕"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.abilityName, func(t *testing.T) {
			description := abilityNamed(t, testCase.abilityName).Description

			for _, phrase := range testCase.mustSay {
				assert.Contains(t, description, phrase)
			}
		})
	}
}
