package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// everyAbilityTheTradingServiceOffers is this connector's promise written down.
//
// The trading service's own chat assistant is deliberately absent, and this list is
// where that decision is enforced: adding it back turns this test red, which is the
// point. One AI must not be able to spend another AI's budget on its own initiative.
//
// It is a literal list rather than something derived from the catalogue, which is the
// only way it can catch anything: a list read out of the thing it is checking agrees
// with it by construction. When the trading service gains an endpoint, this list gains
// a line and goes red until the catalogue gains one too — which is exactly the moment
// somebody should be told.
var everyAbilityTheTradingServiceOffers = map[string]bool{
	// 系統
	"trading_health": false,
	// 帳號（登入、登出、續用是外掛自己的，不在這份轉達清單裡）
	"trading_register_user":    false,
	"trading_get_current_user": true,
	"trading_change_password":  true,
	// 行情
	"trading_create_k_candle":           false,
	"trading_list_k_candles":            false,
	"trading_get_k_candle_series":       false,
	"trading_get_k_candle":              false,
	"trading_update_k_candle":           false,
	"trading_delete_k_candle":           false,
	"trading_backfill_k_candles":        false,
	"trading_sync_k_candle_history":     false,
	"trading_get_k_candle_history_sync": false,
	"trading_peek_live_k_candle":        false,
	// 交易標的
	"trading_list_trading_symbols":  false,
	"trading_add_to_watchlist":      false,
	"trading_remove_from_watchlist": false,
	// 指標
	"trading_calculate_indicator": true,
	// 策略腳本與市集
	"trading_create_strategy_script":   true,
	"trading_list_strategy_scripts":    true,
	"trading_get_strategy_script":      true,
	"trading_update_strategy_script":   true,
	"trading_delete_strategy_script":   true,
	"trading_publish_strategy_script":  true,
	"trading_withdraw_strategy_script": true,
	"trading_browse_marketplace":       true,
	"trading_adopt_strategy_script":    true,
	"trading_abandon_strategy_script":  true,
	// 交易策略
	"trading_create_trading_strategy": true,
	"trading_list_trading_strategies": true,
	"trading_get_trading_strategy":    true,
	"trading_update_trading_strategy": true,
	"trading_delete_trading_strategy": true,
	// 重演
	"trading_backtest_strategy_script":  true,
	"trading_backtest_trading_strategy": true,
	// 策略機器人
	"trading_create_strategy_bot":    true,
	"trading_list_strategy_bots":     true,
	"trading_get_strategy_bot":       true,
	"trading_update_strategy_bot":    true,
	"trading_delete_strategy_bot":    true,
	"trading_start_strategy_bot":     true,
	"trading_stop_strategy_bot":      true,
	"trading_list_strategy_bot_runs": true,
	"trading_run_strategy_bot_now":   true,
	// 通知
	"trading_get_telegram_delivery":      true,
	"trading_save_telegram_delivery":     true,
	"trading_remove_telegram_delivery":   true,
	"trading_send_telegram_test_message": true,
}

func TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse(t *testing.T) {
	requiresSignInPerNames := map[string]bool{}
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		requiresSignInPerNames[apiTool.Name()] = apiTool.RequiresSignIn()
	}

	assert.Equal(t, everyAbilityTheTradingServiceOffers, requiresSignInPerNames)
}

func TestNoTwoAbilitiesShareAName(t *testing.T) {
	catalog := apiToolCatalog(10 * time.Second)

	seenNames := map[string]bool{}
	for _, apiTool := range catalog {
		require.False(t, seenNames[apiTool.Name()],
			"兩件事同名時，後放的會安靜地蓋掉先放的：%s", apiTool.Name())
		seenNames[apiTool.Name()] = true
	}
}

func TestEveryAbilityTellsTheAssistantWhatItIsForAndWhatEachBoxMeans(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		definitionDto := apiTool.ToDefinitionDto()

		assert.NotEmpty(t, definitionDto.Description,
			"沒有說明的能力，助理只能靠名字猜：%s", definitionDto.Name)

		for _, parameter := range definitionDto.Parameters {
			assert.NotEmpty(t, parameter.Description,
				"沒有說明的欄位，助理會一直填錯：%s 的 %s", definitionDto.Name, parameter.Name)
			assert.NotEmpty(t, parameter.Kind,
				"沒有型別的欄位，助理送什麼都可能：%s 的 %s", definitionDto.Name, parameter.Name)
		}
	}
}

func TestOnlyWatchingAnAbilityStaysOnTheLine(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		request, buildError := apiTool.BuildRequest(noFilledInBoxes())
		if buildError != nil {
			continue
		}

		if apiTool.Name() == "trading_peek_live_k_candle" {
			assert.Equal(t, 10*time.Second, request.LiveUpdateWaitLimit)
			continue
		}

		assert.Zero(t, request.LiveUpdateWaitLimit,
			"只有「看一眼即時更新」該停在線上：%s", apiTool.Name())
	}
}

// TestNoAbilityLetsOneAssistantSpendAnothersBudget is a guard on a boundary, not a
// spelling check.
//
// The trading service has its own chat assistant, and relaying it would be trivial —
// three more entries in the catalogue. It is left out because an assistant that can
// call an assistant can run up a token bill with nobody in between deciding it was
// worth it, and each of those calls is minutes long and costs real money.
//
// The person can still use that assistant directly. What must not exist is a model's
// ability to reach for it unprompted, so this checks that no such entry has quietly
// come back.
func TestNoAbilityLetsOneAssistantSpendAnothersBudget(t *testing.T) {
	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		assert.NotContains(t, apiTool.Name(), "assistant",
			"外掛不代 AI 去問另一個 AI：%s", apiTool.Name())
		assert.NotContains(t, apiTool.Name(), "chat",
			"外掛不代 AI 去問另一個 AI：%s", apiTool.Name())
	}
}

// abilityNamed digs one ability out of the catalogue, so a test about one box does not
// have to walk the whole list in front of the reader.
func abilityNamed(t *testing.T, name string) dto.ToolDefinitionDto {
	t.Helper()

	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		if apiTool.Name() == name {
			return apiTool.ToDefinitionDto()
		}
	}

	require.FailNowf(t, "清單裡沒有這件能力", "%s", name)

	return dto.ToolDefinitionDto{}
}

// boxNamed is one declared box, and whether it was declared at all.
func boxNamed(definitionDto dto.ToolDefinitionDto, name string) (dto.ToolParameterDto, bool) {
	for _, parameter := range definitionDto.Parameters {
		if parameter.Name == name {
			return parameter, true
		}
	}

	return dto.ToolParameterDto{}, false
}

// An assistant picking a trading mode cannot see the person's broker and cannot see
// whether the market allows shorting. The sentence they typed is its only clue, so the
// description has to map that sentence onto one of the two spellings — otherwise it
// knows there are three and not which one was just described to it.
func TestWritingATradingStrategySaysWhichKindOfAccountItIsFor(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_trading_strategy",
		"trading_update_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			tradingMode, isDeclared := boxNamed(abilityNamed(t, abilityName), "tradingMode")

			require.True(t, isDeclared, "沒有這個欄位，助理就說不出這份規則是寫給哪一種帳戶的")
			assert.Contains(t, tradingMode.Description, "spot")
			assert.Contains(t, tradingMode.Description, "longShort")
			assert.Contains(t, tradingMode.Description, "leveragedLong")
			// What happens when it says nothing, and which one the person just
			// described. Both are things it can only learn here.
			assert.Contains(t, tradingMode.Description, "省略即 longShort")
			assert.Contains(t, tradingMode.Description, "不能放空")
			// The third one is the one it would otherwise never reach for: somebody
			// who only goes long reads as spot right up until the leverage comes up,
			// and spot is the answer that leaves their bot unsaveable. So the
			// situation has to be named, not only the spelling.
			assert.Contains(t, tradingMode.Description, "只做多、要上一點槓桿")

			// The venue does not settle it, and saying so is the only thing standing
			// between "我在幣安永續" and the default.
			//
			// Two of the three run there, so that sentence rules nothing out — and the
			// two mistakes are not equally survivable. Reaching for spot gets refused,
			// which the person sees. Reaching for long-short, or reaching for nothing at
			// all, turns every sell into a short they never asked for and returns a
			// report card that looks entirely reasonable. So the box has to say to ask.
			assert.Contains(t, tradingMode.Description, "場所不決定模式")
			assert.Contains(t, tradingMode.Description, "沒問出他放不放空之前不要猜")
			assert.Contains(t, tradingMode.Description, "反手做空")
			// Not required: saying nothing is a legitimate thing to do, and the
			// default belongs to the trading service rather than to this list.
			assert.False(t, tradingMode.IsRequired)
		})
	}
}

// Replaying a whole set of rules has no trading mode to give: that set of rules keeps
// its own. A box here would be one the trading service ignores in silence, and an
// assistant has no way to tell it was ignored — it would go on believing it replayed a
// spot account while reading a report card built the other way.
func TestReplayingATradingStrategyHasNoTradingModeToGive(t *testing.T) {
	replay := abilityNamed(t, "trading_backtest_trading_strategy")

	_, isDeclared := boxNamed(replay, "tradingMode")
	assert.False(t, isDeclared, "填了會被交易服務安靜忽略的欄位，比沒有這個欄位更糟")

	// Taking the knob away leaves the assistant knowing only that it does not have
	// one. Where the knob actually is has to be said, or "my account cannot short"
	// gets answered with "I cannot do that".
	assert.Contains(t, replay.Description, "交易模式")
	assert.Contains(t, replay.Description, "trading_update_trading_strategy")
	assert.Contains(t, replay.Description, "不能放空")
}

// Replaying a bare script still asks the caller: there is no trading strategy there to
// ask. Everything else the two replays want is still shared, so the split is one box
// and not a second copy of the replay conditions.
func TestReplayingAStrategyScriptStillAsksWhichWayToTrade(t *testing.T) {
	scriptReplay := abilityNamed(t, "trading_backtest_strategy_script")

	tradingMode, isDeclared := boxNamed(scriptReplay, "tradingMode")
	require.True(t, isDeclared)
	assert.False(t, tradingMode.IsRequired)
	assert.NotEmpty(t, tradingMode.Description)

	strategyReplay := abilityNamed(t, "trading_backtest_trading_strategy")

	// Every condition the trading-strategy replay declares, apart from the identifier
	// in its address, the script replay declares too — and word for word, because both
	// read it out of the same shared list. Matching only the names would pass a second
	// copy of the list that had drifted a sentence.
	for _, strategyReplayBox := range strategyReplay.Parameters {
		if strategyReplayBox.Name == "id" {
			continue
		}

		scriptReplayBox, isDeclared := boxNamed(scriptReplay, strategyReplayBox.Name)

		require.True(t, isDeclared,
			"這一欄兩支回測都要，應該還是共用的那一份：%s", strategyReplayBox.Name)
		assert.Equal(t, strategyReplayBox, scriptReplayBox,
			"這一欄在兩支上不一樣，代表共用的那一份被抄成了兩份：%s", strategyReplayBox.Name)
	}
}

// This box is the only one in the catalogue that filling in wrongly does not get
// refused: leverage on a spot account is a legitimate figure, and four figures with no
// capital read as no plan at all. The trading service accepts both, and the cost lands
// on whoever reads the message — so its description is the only guard there is, and
// every sentence in it is asserted rather than trusted.
func TestWritingAStrategyBotSaysHowToSizeAPosition(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_strategy_bot",
		"trading_update_strategy_bot",
	} {
		t.Run(abilityName, func(t *testing.T) {
			positionPlan, isDeclared := boxNamed(abilityNamed(t, abilityName), "positionPlan")

			require.True(t, isDeclared, "沒有這個欄位，助理組不出一台會建議部位的機器人")
			// Nested, so that "all of it or none of it" is true of the shape.
			assert.Equal(t, string(vo.ToolParameterKindObject), positionPlan.Kind)
			// Not required: a bot that suggests nothing is an ordinary bot.
			assert.False(t, positionPlan.IsRequired)

			// The things that are accepted when wrong, and therefore have to be
			// said here.
			assert.Contains(t, positionPlan.Description, "整組可以不給")
			assert.Contains(t, positionPlan.Description, "capital 是這一組的開關")
			assert.Contains(t, positionPlan.Description, "字串給精確小數")
			assert.Contains(t, positionPlan.Description, "不給即 allIn")
			assert.Contains(t, positionPlan.Description, "百分點")

			// Leverage is no longer one of them — rules that cannot borrow now refuse
			// the whole bot. It still has to be described, for the opposite reason:
			// an assistant that only knows "this gets refused" fixes it by deleting
			// the leverage the person is actually running, rather than by naming the
			// mode that would have allowed it.
			assert.Contains(t, positionPlan.Description, "整台被拒絕")
			assert.Contains(t, positionPlan.Description, "leveragedLong")

			// And the five figures named, so the assistant knows what to put in it.
			for _, figure := range []string{
				"capital", "sizingMode", "sizingValue",
				"leverage", "stopLossPercentage", "takeProfitPercentage",
			} {
				assert.Contains(t, positionPlan.Description, figure)
			}
		})
	}
}

// Both replays take the two exit distances, so they live in the shared list rather
// than being added twice. The test for belonging there is one question — does that
// endpoint actually use it — and it is the same question that keeps the trading mode
// out of it.
func TestBothReplaysTakeTheSameTwoExitDistances(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			for _, boxName := range []string{"stopLossPercentage", "takeProfitPercentage"} {
				box, isDeclared := boxNamed(abilityNamed(t, abilityName), boxName)

				require.True(t, isDeclared,
					"沒有這一欄，助手算不出一組停損的代價：%s", boxName)
				// A replay that simulates no exits is the ordinary one, and was the
				// only one until now.
				assert.False(t, box.IsRequired)
				// Money and every other exact decimal in this catalogue travels as a
				// string.
				assert.Equal(t, string(vo.ToolParameterKindString), box.Kind)
			}
		})
	}
}

// Three things about these two boxes cannot be discovered before sending, and each
// one is wrong in its own way if guessed. Leaving them out simulates nothing, so an
// assistant guessing "a sensible default stop" believes it reconciled something it
// did not. The distances are measured from the entry fill, so the pair sitting on a
// bot — same names, measured from the latest price — is accepted here and answers a
// different question. And a candle reaching both levels counts as the stop, which is
// the only one of the three that moves the numbers the *worse* way: a surprise in the
// good direction gets read as good news, one in the bad direction gets read as a bug.
func TestTheExitDistancesSayWhatCannotBeDiscoveredBySending(t *testing.T) {
	replay := abilityNamed(t, "trading_backtest_strategy_script")

	stopLoss, isDeclared := boxNamed(replay, "stopLossPercentage")
	require.True(t, isDeclared)

	assert.Contains(t, stopLoss.Description, "不給就是完全不模擬止損")
	assert.Contains(t, stopLoss.Description, "進場價")
	assert.Contains(t, stopLoss.Description, "同名、不同事")
	assert.Contains(t, stopLoss.Description, "正好 100 可以")

	takeProfit, isDeclared := boxNamed(replay, "takeProfitPercentage")
	require.True(t, isDeclared)

	assert.Contains(t, takeProfit.Description, "一律算止損")
}

// A report card with the exits simulated needs one more number read off it, and the
// tool description is the only place an assistant learns that.
func TestReplayingAScriptSaysWhyTheStopCountMatters(t *testing.T) {
	description := abilityNamed(t, "trading_backtest_strategy_script").Description

	assert.Contains(t, description, "stopLossExitCount")
	// The whole sentence, not the bare word. The wipe-out note added a second
	// `exitReason` to this description, so matching the word alone stopped proving
	// that *this* paragraph still points at it — the stop-count story could lose
	// its pointer entirely and this would stay green.
	assert.Contains(t, description, "每一筆交易自己也帶著 exitReason")
}

// The cost rates live in the shared list for the same reason the exit distances do:
// both replays take them, because a set of rules has no opinion about what its
// owner's broker charges. Sharing the list is what makes "the same on both" a fact
// rather than a rule somebody has to remember.
func TestBothReplaysTakeTheSameTwoCostRates(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			for _, boxName := range []string{"entryCostPercentage", "exitCostPercentage"} {
				box, isDeclared := boxNamed(abilityNamed(t, abilityName), boxName)

				require.True(t, isDeclared,
					"沒有這一欄，助手每一次重演都在比一個交易免費的世界：%s", boxName)
				// A replay that pays nothing is the ordinary one, and was the only
				// one until now.
				assert.False(t, box.IsRequired)
				assert.Equal(t, string(vo.ToolParameterKindString), box.Kind)
			}
		})
	}
}

// Guessing wrong about these two is not like guessing wrong about anything else in
// this catalogue. Every other bad guess comes back as a refusal or as numbers that
// look odd; a missing cost rate comes back as **a better report card**, and nothing
// anywhere says so.
//
// Worse, the bias grows with how often the strategy trades — which ruins the one job
// the assistant is asked to do most, putting two strategies side by side. So the
// description has to name that job, not merely describe the field.
func TestTheCostRatesSayWhatCannotBeDiscoveredBySending(t *testing.T) {
	replay := abilityNamed(t, "trading_backtest_strategy_script")

	entryCost, isDeclared := boxNamed(replay, "entryCostPercentage")
	require.True(t, isDeclared)

	assert.Contains(t, entryCost.Description, "不給就是完全不計手續費")
	assert.Contains(t, entryCost.Description, "偏樂觀")
	assert.Contains(t, entryCost.Description, "與交易次數成正比")
	assert.Contains(t, entryCost.Description, "不填費率等於沒有在比較")
	assert.Contains(t, entryCost.Description, "正好 100 可以")

	exitCost, isDeclared := boxNamed(replay, "exitCostPercentage")
	require.True(t, isDeclared)

	// The one rule these two break that the exit distances keep: a blank here is not
	// "no charge", it is "the same as the entry". An assistant carrying the exit
	// distances' rule across would believe it had set up a free exit.
	assert.Contains(t, exitCost.Description, "沿用 entryCostPercentage")
	assert.Contains(t, exitCost.Description, "不一樣")
	// It cannot see the person's broker, so the usual numbers have to be here.
	assert.Contains(t, exitCost.Description, "0.0855")
	assert.Contains(t, exitCost.Description, "0.3855")
	assert.Contains(t, exitCost.Description, "幣安")
}

// A costed report card carries one more number and quietly changes the meaning of two
// it already had. The last of those is the expensive one: an assistant still reading
// profit as the price move will call a round trip that earned less than its fees a
// winning trade — exactly the misreading this whole change exists to remove.
func TestBothReplaysSayHowToReadACostedReportCard(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			assert.Contains(t, description, "totalTransactionCost")
			assert.Contains(t, description, "被手續費吃掉")
			assert.Contains(t, description, "淨額")
			assert.Contains(t, description, "勝率")
		})
	}
}

// The assistant's whole loop is build, replay, read the report card, adjust, go live —
// so its most natural next step is to take rules that backtested well and hang a stop
// on them. The replay can now count those exits, but only when asked, so this warning
// stays and points at how to ask rather than saying it cannot be done.
func TestPuttingABotLiveSaysHowToReconcileTheExits(t *testing.T) {
	description := abilityNamed(t, "trading_create_strategy_bot").Description

	// What it must no longer claim: that a replay never counts them.
	assert.NotContains(t, description, "從頭到尾不把止損止盈算進去")
	// What it must still refuse to let the assistant do.
	assert.Contains(t, description, "還沒有對過帳")
	// And the way out, named, because "go and reconcile it" without the box names is
	// advice the assistant cannot act on.
	assert.Contains(t, description, "stopLossPercentage")
	assert.Contains(t, description, "takeProfitPercentage")
}

// Every other box on a bot is untouched: this slice adds one and changes none.
func TestWritingAStrategyBotStillAsksForEverythingItAlwaysDid(t *testing.T) {
	create := abilityNamed(t, "trading_create_strategy_bot")

	for _, boxName := range []string{
		"name", "symbol", "tradingStrategyId", "triggerIntervalMinutes",
	} {
		box, isDeclared := boxNamed(create, boxName)

		require.True(t, isDeclared, boxName)
		assert.True(t, box.IsRequired, "這一欄本來就是必填：%s", boxName)
	}
}

// apiToolNamed is the ability itself rather than the shape an assistant reads, for
// the one question that cannot be asked of the shape: does a filled-in box actually
// leave the connector.
func apiToolNamed(t *testing.T, name string) domains.ApiToolDomain {
	t.Helper()

	for _, apiTool := range apiToolCatalog(10 * time.Second) {
		if apiTool.Name() == name {
			return apiTool
		}
	}

	require.FailNowf(t, "清單裡沒有這件能力", "%s", name)

	return domains.ApiToolDomain{}
}

// The leverage sits in the shared list for the reason the exit distances and the
// cost rates do: both replays take it, because a set of rules has no opinion about
// how much its owner is willing to borrow.
//
// That is also exactly where it parts company with the trading mode, which the
// second replay has no box for at all — one is a fact about the account, the other
// is a property of the rules being replayed.
func TestBothReplaysTakeTheSameLeverage(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			for _, boxName := range []string{"leverage", "maintenanceMarginRate"} {
				box, isDeclared := boxNamed(abilityNamed(t, abilityName), boxName)

				require.True(t, isDeclared,
					"沒有這一欄，助手交出的成績單講的是另一個帳戶：%s", boxName)
				// Borrowing nothing is the ordinary replay, and was the only one
				// until now.
				assert.False(t, box.IsRequired)
				assert.Equal(t, string(vo.ToolParameterKindString), box.Kind)
			}
		})
	}
}

// Both replays word it identically, because two wordings are two chances for only
// one of them to be improved — and then the same figure would mean two things.
func TestBothReplaysWordTheLeverageIdentically(t *testing.T) {
	for _, boxName := range []string{"leverage", "maintenanceMarginRate"} {
		t.Run(boxName, func(t *testing.T) {
			script, isDeclared := boxNamed(
				abilityNamed(t, "trading_backtest_strategy_script"), boxName)
			require.True(t, isDeclared)

			strategy, isDeclared := boxNamed(
				abilityNamed(t, "trading_backtest_trading_strategy"), boxName)
			require.True(t, isDeclared)

			assert.Equal(t, script.Description, strategy.Description)
		})
	}
}

// **This is the mechanical reason the slice exists.** The connector forwards only
// the names it declares, so a box nobody declared is a box the assistant can fill in
// and watch vanish — with no error anywhere, and a report card that looks fine.
func TestAFilledInLeverageActuallyLeavesTheConnector(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_backtest_strategy_script").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"symbol":                json.RawMessage(`"BTCUSDT"`),
			"startTime":             json.RawMessage(`"2026-09-01T00:00:00Z"`),
			"endTime":               json.RawMessage(`"2026-09-10T00:00:00Z"`),
			"initialCapital":        json.RawMessage(`"10000"`),
			"leverage":              json.RawMessage(`"5"`),
			"maintenanceMarginRate": json.RawMessage(`"0.5"`),
		}))

	require.NoError(t, buildError)
	assert.Contains(t, string(request.Body), `"leverage":"5"`)
	assert.Contains(t, string(request.Body), `"maintenanceMarginRate":"0.5"`)
}

// Leaving both out sends what it always sent. Not a new rule — the connector has
// always forwarded only the boxes that were filled in — but it is the promise every
// replay made before this slice depends on, so it is asserted rather than assumed.
func TestLeavingTheLeverageOutSendsWhatItAlwaysSent(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_backtest_strategy_script").
		BuildRequest(domains.NewToolArgumentsDomain(map[string]json.RawMessage{
			"symbol":         json.RawMessage(`"BTCUSDT"`),
			"startTime":      json.RawMessage(`"2026-09-01T00:00:00Z"`),
			"endTime":        json.RawMessage(`"2026-09-10T00:00:00Z"`),
			"initialCapital": json.RawMessage(`"10000"`),
		}))

	require.NoError(t, buildError)
	assert.NotContains(t, string(request.Body), "leverage")
	assert.NotContains(t, string(request.Body), "maintenanceMarginRate")
}

// Six things about this box cannot be discovered by sending, and every one of them
// is wrong in the *invisible* direction: the report card comes back complete and
// plausible for an account that stopped existing halfway through.
func TestTheLeverageSaysWhatCannotBeDiscoveredBySending(t *testing.T) {
	leverage, isDeclared := boxNamed(
		abilityNamed(t, "trading_backtest_strategy_script"), "leverage")
	require.True(t, isDeclared)

	// What leaving it out means — the most common case, and the one an assistant
	// hits on every replay it has ever made.
	assert.Contains(t, leverage.Description, "不給、給 0 或給 1 都是不借錢")
	// That the replay keeps walking after the account is gone. This is the whole
	// slice: the trades after that point did not happen.
	assert.Contains(t, leverage.Description, "重演會照樣往下跑")
	// How far it can fall, said three ways because they answer three questions.
	// Each is asserted by a phrase that appears nowhere else in the description —
	// `100÷槓桿` alone is not enough, since the exact form contains the rule of
	// thumb and either one could quietly disappear behind the other.
	assert.Contains(t, leverage.Description, "大約 `(100÷槓桿)` 個百分點")
	assert.Contains(t, leverage.Description, "`100÷槓桿 − 維持保證金率`")
	assert.Contains(t, leverage.Description, "5 倍約 **19.5%**")
	// The one protection the assistant can actually add, which it would never guess.
	assert.Contains(t, leverage.Description, "stopLossPercentage")
	// Where the invisible damage becomes visible.
	assert.Contains(t, leverage.Description, "liquidationExitCount")
	// And last, the two cases that announce themselves by being refused. Half a
	// multiplier needs saying because the sentence above it names 0 and 1 as
	// harmless, which reads as "anything at or below one is a no-op" — and the
	// value between them is the one an assistant reaches for when it means half a
	// position.
	assert.Contains(t, leverage.Description, "介於 0 與 1 之間會整次被拒絕")
	// Which modes may borrow, rather than which one may not: an assistant told only
	// that spot is refused has no word for the account that is long-only and
	// borrowed, and will send the person back to spot with the leverage removed.
	assert.Contains(t, leverage.Description, "現貨（spot）借不到")
	assert.Contains(t, leverage.Description, "leveragedLong")
}

// The two boxes sit on the same call and both describe what a percentage is taken
// of, so they must not answer it differently. Borrowing multiplies what the entry
// charge is levied on, and the cost box is where an assistant looks for that.
func TestTheEntryCostAndTheLeverageAgreeOnWhatTheFeeIsTakenOf(t *testing.T) {
	entryCost, isDeclared := boxNamed(
		abilityNamed(t, "trading_backtest_strategy_script"), "entryCostPercentage")
	require.True(t, isDeclared)

	assert.Contains(t, entryCost.Description, "佔**曝險金額**的百分之幾")
	// And says so in a way that needs no cross-referencing: what it means when
	// nothing is borrowed, and what it costs when something is.
	assert.Contains(t, entryCost.Description, "沒開槓桿時曝險金額就是押注金額")
	assert.Contains(t, entryCost.Description, "開了 5 倍，同一個費率收的錢就是五倍")
}

// The maintenance margin's blank means something different from every blank beside
// it, and three groups sit together on the same call. Nothing but this sentence
// stops an assistant carrying the neighbours' rule across.
func TestTheMaintenanceMarginSaysItsBlankIsNotTheOthers(t *testing.T) {
	rate, isDeclared := boxNamed(
		abilityNamed(t, "trading_backtest_strategy_script"), "maintenanceMarginRate")
	require.True(t, isDeclared)

	assert.Contains(t, rate.Description, "不給不是關掉它，是用 0.5%")
	// The second half this test is named for, which it did not assert: *why* the
	// blank is not the neighbours'. Deleting the contrast clause left it green,
	// and that clause is the entire reason this box needs a sentence of its own.
	assert.Contains(t, rate.Description, "這與旁邊每一格的留白都不一樣")
	assert.Contains(t, rate.Description, "100÷槓桿")
}

// A borrowed replay puts one more number on the report card, and the tool
// description is the only place an assistant learns what it is for. Both replays say
// it, word for word, because they hand back the same report card.
func TestBothReplaysSayWhyTheWipeOutCountMatters(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			assert.Contains(t, description, "liquidationExitCount")
			// Why, not what: that a respectable return can belong to an account
			// emptied three times on the way is the part it cannot see.
			assert.Contains(t, description, "歸零過三次")
			// And it is the *same* paragraph, not a second one that happens to
			// mention the same two things. Asserting the words alone would let a
			// copy drift a sentence at a time while this stayed green, and then
			// the two replays would describe one report card two ways.
			assert.Contains(t, description, liquidationReportCardNote)
		})
	}
}

// Replaying a script is the one path where the assistant types the trading mode
// itself: there is no stored strategy to read it off, so this box is the only place
// it could learn what the spellings are.
//
// It used to name none of them — "省略即一直留在市場裡" and nothing else — which left
// an assistant guessing at a string, and guessing wrong is refused outright.
func TestReplayingAScriptSaysWhichModesItMayBeTold(t *testing.T) {
	tradingMode, isDeclared := boxNamed(
		abilityNamed(t, "trading_backtest_strategy_script"), "tradingMode")

	require.True(t, isDeclared, "沒有這個欄位，助理就說不出這一次要照哪一套規則重演")
	// Each spelling with what sets it apart, not the spelling alone: they appear in
	// the prose beside each other, so a bare substring check would stay green on a
	// box that offers only two of the three.
	for _, describedMode := range []struct {
		spelling    string
		description string
	}{
		{"longShort", "做得了空、也借得到錢"},
		{"spot", "做不了空、也借不到錢"},
		{"leveragedLong", "做不了空、但借得到錢"},
	} {
		assert.Contains(t, tradingMode.Description, describedMode.spelling)
		assert.Contains(t, tradingMode.Description, describedMode.description)
	}

	// What saying nothing means, and that a wrong guess is not quietly swallowed.
	assert.Contains(t, tradingMode.Description, "省略即 longShort")
	assert.Contains(t, tradingMode.Description, "整次被拒絕")
	// And that naming the venue has not answered the question. This replay is the
	// one where the assistant types the mode, so it is also the one where defaulting
	// silently reverses somebody into shorts.
	assert.Contains(t, tradingMode.Description, "場所不決定模式")
	assert.False(t, tradingMode.IsRequired)
}
