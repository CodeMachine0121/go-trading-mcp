package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var everyAbilityThatWritesOrReadsATradingStrategy = []string{
	"trading_create_trading_strategy", "trading_get_trading_strategy", "trading_update_trading_strategy",
}

// A trading strategy only says buy, sell or hold; every ability touching one says where
// the direction went, so an assistant asked for "only long" sets it on a bot or a replay.
func TestEveryTradingStrategyAbilitySaysTheDirectionMovedToBotsAndReplays(t *testing.T) {
	for _, abilityName := range everyAbilityThatWritesOrReadsATradingStrategy {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"**交易策略沒有交易方向**", "讀回來也沒有交易模式",
				"**建立或修改時送 tradingMode 會整份被拒絕**",
				"tradingDirections", "trading_backtest_contract_trading_strategy 的 tradingMode",
				"**同一份交易策略可以同時養一台做多的機器人和一台做空的機器人**",
				"使用者說「這份策略只做多」時，把方向設在機器人或回測上",
			} {
				assert.Contains(t, description, phrase)
			}
			assert.NotContains(t, description, "記著自己的交易模式")
			assert.NotContains(t, description, "另外帶著它的交易模式")
		})
	}
}

// A bot's direction and notifications are forwarded exactly as given.
func TestWritingAStrategyBotForwardsItsDirectionAndNotificationsAsGiven(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		for _, given := range []struct{ box, value string }{
			{"tradingDirections", `["long"]`},
			{"tradingDirections", `["short"]`},
			{"tradingDirections", `["long","short"]`},
			{"tradingDirections", `[]`},
			{"notifications", `{"buy":false}`},
			{"notifications", `{"buy":true,"sell":true,"hold":true}`},
		} {
			t.Run(abilityName+" "+given.box+" "+given.value, func(t *testing.T) {
				request, buildError := apiToolNamed(t, abilityName).BuildRequest(aBotWriteForm(map[string]json.RawMessage{
					"marketDataKind": json.RawMessage(`"contractKCandle"`),
					given.box:        json.RawMessage(given.value),
				}))

				require.NoError(t, buildError)
				assert.Contains(t, string(request.Body), `"`+given.box+`":`+given.value)
			})
		}
	}
}

// Left out, nothing is sent: the trading service picks both directions and the default
// notifications on a create, and keeps what is stored on a rewrite.
func TestWritingAStrategyBotWithoutDirectionOrNotificationsSendsNeither(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, abilityName).BuildRequest(aBotWriteForm(nil))

			require.NoError(t, buildError)
			assert.NotContains(t, string(request.Body), "tradingDirections")
			assert.NotContains(t, string(request.Body), "notifications")
		})
	}
}

func TestWritingAStrategyBotDeclaresItsDirectionAndNotifications(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			ability := abilityNamed(t, abilityName)

			tradingDirections, isDeclared := boxNamed(ability, "tradingDirections")
			require.True(t, isDeclared)
			assert.False(t, tradingDirections.IsRequired)
			for _, phrase := range []string{
				"**只有合約機器人有**", "long（做多）或 short（做空），可以兩個都給", "**至少要一個**，給空陣列會被拒絕",
				"**現貨機器人給了會被拒絕**", "**建立時不給就是兩個都選（多空反手）**", "**修改時不給就是保留原本的**",
				"**使用者說了方向才給**",
			} {
				assert.Contains(t, tradingDirections.Description, phrase)
			}
			// No filled-in direction to copy: copied onto a create it makes a one-sided bot
			// nobody asked for, or gets a spot bot refused.
			assert.NotContains(t, tradingDirections.Description, `["long"]`)
			assert.NotContains(t, tradingDirections.Description, `["short"]`)

			notifications, isDeclared := boxNamed(ability, "notifications")
			require.True(t, isDeclared)
			assert.False(t, notifications.IsRequired)
			for _, phrase := range []string{
				"鍵是 buy（買入）、sell（賣出）、hold（持有）", "**只放使用者要改的那幾個**",
				"**建立時沒給的是買入開、賣出開、持有關**",
				"**修改時沒給的維持原本的**", `只想關買入通知就只送 {"buy":false}`,
			} {
				assert.Contains(t, notifications.Description, phrase)
			}
			// No complete filled-in object to copy: copied on a rewrite it would reset
			// switches the owner had changed.
			assert.NotContains(t, notifications.Description, `"sell":true`)
			assert.NotContains(t, notifications.Description, `"hold":false`)
		})
	}
}

// The direction decides what buy and sell mean, and three things refuse a change of it.
func TestWritingAStrategyBotSaysWhatItsDirectionMeansAndWhenItCannotChange(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"只有做多 → 買入＝做多、賣出＝平多", "只有做空 → 賣出＝做空、買入＝平空", "兩個都選 → 多空反手",
				"同一個賣出結論在做多那台是平多、在做空那台是做空",
				"reason 為 tradingDirectionAutoOrderOn 是**這台機器人的自動下單開著**",
				"watchedSymbolHoldsPosition／watchedSymbolAutoOrderOutstanding",
				"照原話說出是哪個標的，要先讓它空手或等它做完",
				"**交易方向沒變時不受這兩道限制**",
				"**自動下單開著的合約機器人，這個外掛改不了它的交易方向**",
				"網頁上那台機器人的詳細頁", "**不要重送同一份**",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

// Silencing a conclusion only silences its message: the signal, the duplicate guard and
// the auto order all still act on it.
func TestWritingAStrategyBotSaysANotificationOnlySilencesTheMessage(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"**結論通知只管寄不寄那一則輪次訊息**", "照樣擋掉之後重複的結論", "**開著自動下單時照樣下單**",
				"寄一則「回到持有」，接下來連續的持有不再寄", "衝突不算持有、也不打斷連續的持有",
				"剛啟動的第一個結論是持有也會寄", "「回到持有」那一則**沒有建議部位、不觸發自動下單**",
				"三個都關也可以：機器人照跑", "已啟動、已停止這類機器人自己的動靜照寄",
				"改通知與改其他設定一樣，機器人須已停止",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestCreatingAContractBotSaysItReadsItsOwnDirection(t *testing.T) {
	description := abilityNamed(t, "trading_create_strategy_bot").Description

	assert.Contains(t, description, "每一輪照**這台機器人自己的交易方向**（tradingDirections）")
	assert.NotContains(t, description, "照那份交易策略的交易模式")
}

func TestRewritingAStrategyBotSaysTheDirectionAndNotificationsAreKept(t *testing.T) {
	description := abilityNamed(t, "trading_update_strategy_bot").Description

	for _, phrase := range []string{
		"**例外有五樣，不給就是保留原本的**", "tradingDirections（交易方向，不給不會回到多空反手）",
		"notifications（結論通知，沒給的那幾個維持原本的）", "**只有已停止的機器人改得動**",
	} {
		assert.Contains(t, description, phrase)
	}
}

func TestEveryAbilityAnsweringWithABotSaysHowToReadItsDirectionAndNotifications(t *testing.T) {
	for _, abilityName := range everyAbilityThatAnswersWithAStrategyBot {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"**合約機器人帶著 tradingDirections（交易方向）**", `["long"] 只做多`, `["short"] 只做空`,
				`["long","short"] 多空反手`, "現貨機器人沒有這一格",
				"**每一台機器人都帶著 notifications**（buy 買入通知、sell 賣出通知、hold 持有通知",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

// The instructions say it too, so an assistant asked to flip a live bot never promises it.
func TestTheInstructionsSayADirectionCannotChangeWhileAutoOrderIsOn(t *testing.T) {
	instructions := assistantSessionOn(t, assembledConnector(t)).InitializeResult().Instructions

	assert.Contains(t, instructions, "**自動下單開著的合約機器人，這個外掛改不了它的交易方向**")
	assert.Contains(t, instructions, "結論通知不受這個限制，這個外掛改得動")
}

// Screening profiles say their own trading mode, unchanged by trading strategies losing theirs.
func TestAScreeningProfileStillSaysItsOwnTradingMode(t *testing.T) {
	for _, abilityName := range []string{
		"trading_create_symbol_screening_profile", "trading_update_symbol_screening_profile",
	} {
		t.Run(abilityName, func(t *testing.T) {
			tradingMode, isDeclared := boxNamed(abilityNamed(t, abilityName), "tradingMode")

			require.True(t, isDeclared)
			assert.Contains(t, tradingMode.Description, "longShort／longOnly／shortOnly")
		})
	}
}
