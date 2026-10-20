package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWritingAStrategyBotForwardsTheCapitalAllocationsAsGiven(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		for _, givenPositionPlan := range []string{
			`{"capital":"1000","capitalAllocations":[{"symbol":"BTCUSDT","allocationPercentage":"60"},{"symbol":"ETHUSDT","allocationPercentage":"40"}]}`,
			`{"capital":"1000","capitalAllocations":[{"symbol":"BTCUSDT","allocationPercentage":"30"},{"symbol":"ETHUSDT","allocationPercentage":"30"}]}`,
			`{"capital":"1000","capitalAllocations":[{"symbol":"BTCUSDT","allocationPercentage":50},{"symbol":"ETHUSDT","allocationPercentage":50}]}`,
			`{"capital":"1000","capitalAllocations":[{"symbol":"BTCUSDT","allocationPercentage":"70"},{"symbol":"ETHUSDT","allocationPercentage":"40"}]}`,
			`{"capital":"1000","capitalAllocations":[{"symbol":"BTCUSDT","allocationPercentage":"33.333"}]}`,
			`{"capital":"1000","capitalAllocations":[{"symbol":"SOLUSDT","allocationPercentage":"10"}]}`,
		} {
			t.Run(abilityName+" "+givenPositionPlan, func(t *testing.T) {
				request, buildError := apiToolNamed(t, abilityName).BuildRequest(aBotWriteForm(map[string]json.RawMessage{
					"symbols":      json.RawMessage(`["BTCUSDT","ETHUSDT"]`),
					"positionPlan": json.RawMessage(givenPositionPlan),
				}))

				require.NoError(t, buildError)
				assert.Contains(t, string(request.Body), `"positionPlan":`+givenPositionPlan)
			})
		}
	}
}

func TestWritingAStrategyBotWithoutCapitalAllocationsSendsNone(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, abilityName).BuildRequest(aBotWriteForm(map[string]json.RawMessage{
				"symbols":      json.RawMessage(`["BTCUSDT","ETHUSDT","SOLUSDT"]`),
				"positionPlan": json.RawMessage(`{"capital":"1000"}`),
			}))

			require.NoError(t, buildError)
			assert.NotContains(t, string(request.Body), "capitalAllocations")
			assert.NotContains(t, string(request.Body), "allocationPercentage")
		})
	}
}

func TestWritingAStrategyBotSaysHowToGiveCapitalAllocations(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			description := positionPlanDescriptionOf(t, abilityName)

			for _, phrase := range []string{
				"**capital 是整台機器人的總金額**", "合約機器人的 capital 是**保證金**，不是名目",
				"從那個標的分到的資金算出來", "超過那個標的分到的資金就押不下",
				"**capitalAllocations（資金分配）選填**", `"capitalAllocations":[{"symbol":"BTCUSDT","allocationPercentage":"60"}`,
				"**不給就是平均分配**", "**每一個監聽標的都要給**（只給一部分會被拒絕）", "只能指名這台的監聽標的",
				"同一個標的不得給兩次", "**大於 0、不超過 100**", "**最多到小數點後兩位**", "**加起來不得超過 100**",
				"可以少於 100，剩下的不動用", "照原話告訴使用者",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}

func TestWritingAStrategyBotSaysTheCapitalIsTheBotsTotalSplitAcrossItsSymbols(t *testing.T) {
	for _, abilityName := range bothStrategyBotWrites {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"capital 是整台機器人的總金額，每個監聽標的只用它分到的那一份",
				"分到的資金＝capital × 那個標的的分配比例 ÷ 100",
				"BTCUSDT 分到 600、ETHUSDT 分到 400", "所有標的同時開滿也不超過總金額",
				"**合約機器人的總金額是保證金，不是名目**", "保證金最多 50、名目 250",
				"沒用到的份額**不借給**別的標的",
				"**不給 capitalAllocations 就是平均分配**", "無條件捨去到小數點後兩位",
				"3 個各 33.33%（總金額 1000 時各分到 333.3，0.01% 沒動用）", "7 個各 14.28%",
				"每個只分到 25、押不下 50", "這個標的分到的資金不足", "必要時調大總金額或改倉位大小",
				"每一個監聽標的都會真的在使用者的幣安帳戶上開倉平倉**，用的就是它分到的那一份",
			} {
				assert.Contains(t, description, phrase)
			}
			assert.NotContains(t, description, "各自照整份算")
			assert.NotContains(t, description, "N 倍")
		})
	}
}

func TestRewritingAStrategyBotSaysWhenTheAllocationIsKeptOrSplitAgain(t *testing.T) {
	description := abilityNamed(t, "trading_update_strategy_bot").Description

	for _, phrase := range []string{
		"監聽標的與原本完全相同（順序不論）就**保留原本的分配比例**",
		"只改名字、間隔或其他部位規劃時不必重帶",
		"**只要多了或少了任何一個標的，就重新平均分配**", "使用者自己調過的分配會不見",
		"加減標的前先跟使用者說清楚，或把新的分配比例一起送",
	} {
		assert.Contains(t, description, phrase)
	}
}

func TestEveryAbilityAnsweringWithABotSaysHowToReadEachSymbolsAllocation(t *testing.T) {
	for _, abilityName := range everyAbilityThatAnswersWithAStrategyBot {
		t.Run(abilityName, func(t *testing.T) {
			description := abilityNamed(t, abilityName).Description

			for _, phrase := range []string{
				"**allocationPercentage 分配比例**",
				"分到的資金＝positionPlan 的 capital × allocationPercentage ÷ 100",
				"回覆裡沒有這個數字，要自己算", "少於 100 的部分是沒動用的錢",
				"「資金分配 33.33%：總金額 1000 中的 333.3」", "100% 時訊息不多這一行",
			} {
				assert.Contains(t, description, phrase)
			}
		})
	}
}
