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
	// 永續合約（自成一條路，與現貨互不相干）
	"trading_create_contract_k_candle":               false,
	"trading_list_contract_k_candles":                false,
	"trading_get_contract_k_candle":                  false,
	"trading_update_contract_k_candle":               false,
	"trading_delete_contract_k_candle":               false,
	"trading_backfill_contract_k_candles":            false,
	"trading_sync_contract_k_candle_history":         false,
	"trading_get_contract_k_candle_history_sync":     false,
	"trading_list_contract_trading_symbols":          false,
	"trading_add_to_contract_watchlist":              false,
	"trading_remove_from_contract_watchlist":         false,
	"trading_list_contract_funding_rate_settlements": false,
	"trading_list_contract_position_statistics":      false,
	"trading_get_contract_maintenance_margin_tiers":  false,
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

// There is one set of rules this service replays, so no ability offers a box for
// choosing one — not the two that write a trading strategy, and not either replay.
//
// Asserted as an absence across all four, because a box an assistant can see is a box
// it will fill in: it would pick a mode, watch the whole call be refused, and try
// another spelling. The refusal is the trading service's; the wasted round trip is
// this catalogue's.
func TestNoAbilityOffersATradingModeToChoose(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_trading_strategy",
		"trading_update_trading_strategy",
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)

			_, isDeclared := boxNamed(ability, "tradingMode")
			assert.False(t, isDeclared, "這個服務只重演現貨，沒有模式可挑")

			// And the old advice is gone with the box. A description still naming
			// the spellings would have the assistant telling somebody to go and
			// change a setting that no longer exists — advice that reads perfectly
			// sensibly and cannot be carried out.
			//
			// **Every box's description is read as well as the ability's own.** The
			// advice lived almost entirely in box descriptions, so scanning only the
			// ability's own text leaves this assertion true of three of these four
			// before the change — green, and proving nothing.
			for _, removedSpelling := range []string{
				"longShort", "leveragedLong", "shortOnly",
			} {
				assert.NotContains(t, ability.Description, removedSpelling)
				for _, box := range ability.Parameters {
					assert.NotContainsf(t, box.Description, removedSpelling,
						"那三個拼法還留在 %s 這一格的說明裡", box.Name)
				}
			}
		})
	}
}

// The two writing abilities still ask for everything they always asked for.
//
// It is the counterweight to the absence assertion above: that one alone stays green
// on an ability whose parameter list has been emptied altogether, and an assistant
// handed a tool with no boxes cannot build anything at all.
func TestWritingATradingStrategyStillAsksForEverythingItAlwaysDid(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_trading_strategy",
		"trading_update_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)

			for _, boxName := range []string{
				"name", "signalSources", "buyCondition", "sellCondition",
			} {
				box, isDeclared := boxNamed(ability, boxName)

				require.Truef(t, isDeclared, "少了這一格就拼不出一份交易策略：%s", boxName)
				assert.NotEmptyf(t, box.Description, "這一格沒有說明：%s", boxName)
			}
		})
	}
}

// Everything the two replays want is shared, so there is one list and not two copies
// of the replay conditions.
func TestBothReplaysReadTheSameConditionsOutOfOneList(t *testing.T) {
	scriptReplay := abilityNamed(t, "trading_backtest_strategy_script")
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

// This box is the only one in the catalogue whose contents the trading service does
// not police: the plan arrives as opaque JSON, so what the connector forwards is
// whatever the assistant put inside it. Figures with no capital read as no plan at
// all, and the cost lands on whoever reads the message — so its description is the
// only guard there is, and every sentence in it is asserted rather than trusted.
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

			// There is no borrowing to describe, and the description says so rather
			// than leaving the absence to be discovered. An assistant that finds no
			// leverage box reads it as "not supported yet" and looks for another way
			// to express it; told there is none, it stops looking.
			assert.Contains(t, positionPlan.Description, "沒有槓桿這一項")
			assert.NotContains(t, positionPlan.Description, "leveragedLong")

			// **And the key is absent from the shape**, which is the half that
			// matters: this plan travels as opaque JSON, so the assistant copies
			// this example verbatim and the connector forwards whatever is in it.
			// A sentence saying "no leverage" beside an example that still shows
			// `"leverage":"3"` is a sentence nobody reads.
			assert.NotContains(t, positionPlan.Description, `"leverage"`)
			assert.NotContains(t, positionPlan.Description, `\"leverage\"`)

			// And the four figures named, so the assistant knows what to put in it.
			for _, figure := range []string{
				"capital", "sizingMode", "sizingValue",
				"stopLossPercentage", "takeProfitPercentage",
			} {
				assert.Contains(t, positionPlan.Description, figure)
			}
		})
	}
}

// Both replays take the two exit distances, so they live in the shared list rather
// than being added twice. The test for belonging there is one question: does that
// endpoint actually use it.
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

// **This is the mechanical half of the slice.** The connector forwards only the names
// it declares, so taking the two boxes off the list is what actually stops a borrowed
// replay leaving here — a description that merely stopped mentioning them would leave
// an assistant able to send one and watch it work.
func TestNeitherReplayTakesALeverage(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			for _, boxName := range []string{"leverage", "maintenanceMarginRate"} {
				_, isDeclared := boxNamed(abilityNamed(t, abilityName), boxName)

				assert.False(t, isDeclared,
					"這個服務借不到錢，這一欄只會讓助手送出一個必定被拒絕的值：%s", boxName)
			}
		})
	}
}

// A multiplier the assistant sends anyway never leaves here. The connector forwards
// only what it declares, so an undeclared box is dropped — and the trading service
// would have refused it in any case.
func TestAFilledInLeverageNeverLeavesTheConnector(t *testing.T) {
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
	assert.NotContains(t, string(request.Body), "leverage")
	assert.NotContains(t, string(request.Body), "maintenanceMarginRate")
	// The conditions that are still real go out untouched, so this is not passing
	// because the body came back empty.
	assert.Contains(t, string(request.Body), `"initialCapital":"10000"`)
}

// The entry charge is taken of what was put down, because that is the whole of what a
// spot position has in the market. It used to say "exposure", which was the stake
// multiplied by a loan — and the loan is gone.
func TestTheEntryCostIsTakenOfWhatWasPutDown(t *testing.T) {
	entryCost, isDeclared := boxNamed(
		abilityNamed(t, "trading_backtest_strategy_script"), "entryCostPercentage")
	require.True(t, isDeclared)

	assert.Contains(t, entryCost.Description, "佔**押下去的金額**的百分之幾")
	assert.NotContains(t, entryCost.Description, "曝險金額")
	assert.NotContains(t, entryCost.Description, "槓桿")
}

// Both replays say what this service actually replays — and, just as importantly,
// what it does not. Word for word, because they replay the same thing.
//
// The second half is the part an assistant cannot work out from the boxes. A missing
// box reads as "not supported yet, try another way", and the way it tries is bending
// somebody's strategy into a shape that produces a report card for trades they cannot
// place. The note ends by telling it to say so instead.
func TestBothReplaysSayTheyOnlyEverTradeSpot(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			// What it does.
			assert.Contains(t, description, "重演只做現貨")
			assert.Contains(t, description, "空手時聽到賣出什麼都不做")
			// What it does not.
			assert.Contains(t, description, "沒有交易模式可以指定，也開不了槓桿")
			// And what to say to somebody who asks for it anyway.
			assert.Contains(t, description, "不要替他換一組設定去湊")

			// The *same* paragraph, not a second one that happens to mention the
			// same things. Asserting the sentences alone would let a copy drift a
			// sentence at a time while this stayed green, and then the two replays
			// would describe one service two ways.
			assert.Contains(t, description, spotOnlyReplayNote)
		})
	}
}

// The report card no longer carries a count of positions taken off because the money
// behind them ran out, so no description may promise one.
//
// Asserted as an absence, because this is how it would come back: a sentence about a
// figure the assistant then goes looking for, does not find, and answers around.
func TestNoReplayPromisesAWipeOutCount(t *testing.T) {
	for _, abilityName := range []string{
		"trading_backtest_strategy_script",
		"trading_backtest_trading_strategy",
	} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			assert.NotContains(t, description, "liquidationExitCount")
			// Forced liquidation may still be *named*, and is — but only in the
			// sentence saying this service does not do it. What must not survive is
			// a promise that the report card counts them.
			assert.NotContains(t, description, "成績單多一個 liquidationExitCount")
			assert.Contains(t, description, "不會有強制平倉這種出場")
			// The figure that is still on the report card is still described, so this
			// is not passing because the description lost its report-card paragraph
			// altogether.
			assert.Contains(t, description, "totalTransactionCost")
		})
	}
}
