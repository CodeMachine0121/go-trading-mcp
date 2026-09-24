package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// positionPlanDescriptionOf is what a bot-writing ability says about the position plan.
func positionPlanDescriptionOf(t *testing.T, abilityName string) string {
	t.Helper()

	positionPlan, isDeclared := boxNamed(abilityNamed(t, abilityName), "positionPlan")
	require.True(t, isDeclared)

	return positionPlan.Description
}

// Writing a bot says a contract bot's suggestion follows the venue's rules: stepped,
// ticked, with an estimated liquidation price and a funding estimate.
func TestWritingABotSaysAContractSuggestionFollowsTheVenuesRules(t *testing.T) {
	for _, abilityName := range []string{"trading_create_strategy_bot", "trading_update_strategy_bot"} {
		t.Run(abilityName, func(t *testing.T) {
			description := positionPlanDescriptionOf(t, abilityName)

			for _, phrase := range []string{
				"合約機器人的建議照交易所規則算", "照數量步進往下取整", "保證金與名目照取整後的數量重算",
				"照價格跳動單位取最接近的一檔",
				"預估強平價", "名目所在那一級", "沒有分級時用最小那一級估算",
				"資金費率成本估算", "最近一次結算的費率 × 名目", "標明是估算", "還沒有結算紀錄就不估",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

// It names the three ways the venue refuses an order, and what to change.
func TestWritingABotNamesWhenTheVenueRefusesAContractSuggestion(t *testing.T) {
	description := positionPlanDescriptionOf(t, "trading_create_strategy_bot")

	for _, phrase := range []string{
		"交易所不收的那一筆只說不收與原因", "低於最小下單量", "名目低於最小名目",
		"那一級最高槓桿低於這台機器人的槓桿", "調整 capital、sizingMode／sizingValue 或 leverage",
	} {
		assert.Contains(t, description, phrase)
	}
}

// A contract with no trading specification still gets a suggestion, unrounded.
func TestWritingABotSaysWhatHappensWithoutATradingSpecification(t *testing.T) {
	description := positionPlanDescriptionOf(t, "trading_create_strategy_bot")

	assert.Contains(t, description, "合約標的還沒有交易規格時**照舊建議，但不取整、不估強平價")
}

// The liquidation warning is the stop against the estimated liquidation price; the
// rough rule is only the fallback.
func TestWritingABotComparesTheStopWithTheEstimatedLiquidationPrice(t *testing.T) {
	description := positionPlanDescriptionOf(t, "trading_create_strategy_bot")

	assert.Contains(t, description, "**止損比預估強平價還遠時**，訊息會警告還沒到止損就會先被強制平倉")
	assert.Contains(t, description, "這時強平警告退回粗略判斷（止損距離 × 槓桿達到 100%）")
	// The rough rule is no longer stated as the rule.
	assert.NotContains(t, description, "止損距離乘上槓桿達到 100% 時，訊息會警告")
}

// Reading a bot's rounds names the three figures a contract round's record carries.
func TestReadingABotsRoundsNamesTheContractRunDetails(t *testing.T) {
	for _, abilityName := range []string{"trading_list_strategy_bot_runs", "trading_run_strategy_bot_now"} {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"合約機器人的執行紀錄多帶三樣**：那一輪有建議部位時",
				"suggestedDirection（long 做多／short 做空）", "suggestedLeverage", "suggestedNotional（名目）",
				"現貨機器人的紀錄沒有這三樣，交易所不收那一筆的那一輪也沒有",
				"suggestedStake 是**保證金**", "**取整後**的數字，與當時送出的訊息一致",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

// The spot sentences stay word for word.
func TestWritingABotKeepsTheSpotSentences(t *testing.T) {
	description := positionPlanDescriptionOf(t, "trading_create_strategy_bot")

	assert.Contains(t, description, "**現貨機器人給大於一倍的槓桿會被拒絕**：現貨是拿現金換東西，沒有人借錢給你")
	assert.Contains(t, description, "sizingMode 三選一：allIn 全押（不必給 sizingValue）、percentage 押資金的百分之幾、")
}
