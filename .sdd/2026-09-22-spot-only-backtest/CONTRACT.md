# Contract Traceability Matrix — 工具目錄只提供現貨重演

Contract: `.sdd/2026-09-22-spot-only-backtest/PRD.md`
Design map: `.sdd/2026-09-22-spot-only-backtest/ARCH.md`
Implementation: `cmd/server/tool_catalog*.go`
Oracle: Acceptance Criteria — 23 clauses (17 `AC-`, 4 `BR-`, 2 `NFR-`)

## Clauses

`T/` = `cmd/server/tool_catalog_test.go`。

### US-01 — 助手挑不到一個不存在的選項

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-1 | 建立交易策略的工具沒有這一格 | 參數裡沒有任何一格在問寫給哪一種帳戶 | `tool_catalog_strategy.go:tradingStrategyWriteParameters` | `T/:194` `NoAbilityOffersATradingModeToChoose` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-2 | 修改交易策略的工具與建立那一支一致 | 參數逐項相同，同樣沒有那一格 | 同上（兩支共用同一個函式） | 同上（迴圈涵蓋 update） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-3 | 重演一支策略腳本的工具沒有這一格 | 沒有一格在問這一次照哪一套規則交易 | `tool_catalog_strategy.go`（自有的那一格已移除） | 同上（迴圈涵蓋兩支重演） | asserts-oracle | produces-oracle | ✅ conforms |

### US-02 — 助手開不了一個借不到的槓桿

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-4 | 重演一支策略腳本不再問借幾倍 | 沒有借幾倍、沒有維持保證金率；其餘八格照舊 | `tool_catalog.go:backtestParameters`（12→10 格） | `T/:483` `NeitherReplayTakesALeverage`；`T/:293`、`T/:357`（止損與成本仍在） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-5 | 重演一份交易策略不再問借幾倍 | 同上 | 同上（兩支共用） | 同上（迴圈涵蓋兩支） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-6 | 機器人的建議部位剩四樣 | 說明列出四樣，沒有槓桿 | `tool_catalog_automation.go:strategyBotWriteParameters` | `T/:249` `WritingAStrategyBotSaysHowToSizeAPosition` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-7 | 修改機器人與建立那一支一致 | 說明逐字相同 | 同上（兩支共用同一個函式） | 同上（迴圈涵蓋 update） | asserts-oracle | produces-oracle | ✅ conforms |

### US-03 — 助手知道這個服務做什麼、不做什麼

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-8 | 目錄說出重演怎麼走倉位 | 說明寫著只做現貨，以及買入／賣出／空手時賣出各自的結果 | `tool_catalog_strategy.go:spotOnlyReplayNote` | `T/:541` `BothReplaysSayTheyOnlyEverTradeSpot` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-9 | 目錄說出這裡做不到什麼 | 說明寫著沒有交易模式可指定、開不了槓桿 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 目錄說出遇到那種要求該怎麼回 | 說明要助手直接說只重演現貨、不要換設定去湊 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-11 | 舊的挑模式指引整段不見 | 找不到那三個詞，也找不到那類建議 | 三個 catalog 檔的整段刪除 | `T/:194`（四支工具逐支斷言 NotContains 三個拼法） | asserts-oracle | produces-oracle | ✅ conforms |

### US-04 — 成績單的描述與成績單本身一致

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-12 | 不再提那一格已經不存在的數字 | 說明裡沒有那一格 | `liquidationReportCardNote` 刪除 | `T/:571` `NoReplayPromisesAWipeOutCount` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-13 | 還在的那幾格照舊被提到 | 止損出場筆數與累計成本照舊，措辭逐字相同 | `costedReportCardNote`（不動）；止損那一段（不動） | `T/:415` `BothReplaysSayHowToReadACostedReportCard`；`T/:342` `ReplayingAScriptSaysWhyTheStopCountMatters` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-14 | 進場成本照押下去的金額收 | 說明說照押下去的金額收，不再提曝險與倍數 | `tool_catalog.go:entryCostPercentage` | `T/:524` `TheEntryCostIsTakenOfWhatWasPutDown` | asserts-oracle | produces-oracle | ✅ conforms |

### US-05 — 與重演無關的工具一支都不動

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-15 | 工具一支都沒少 | 支數與名稱逐項相同 | 未改動任何 `NewApiToolDomain` 的名稱、動詞或路由 | `T/:87` `TheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse`（未改動且綠） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-16 | 無關的工具逐字不變 | 參數與說明逐字相同 | `tool_catalog_market.go` 與 `internal/` 未改動 | 同上 + `T/:107`、`T/:123`、`T/:151`（皆未改動且綠） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-17 | *(隱含於 AC-15／16)* 機器人生命週期與通知不受影響 | 逐字相同 | 未改動 | `T/:435`、`T/:449`（未改動且綠） | asserts-oracle | produces-oracle | ✅ conforms |

### Section 4 — Core Business Rules

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| BR-1 | 從目錄上消失的三樣（表），且消失＝助手看不到也送不出 | 四類工具各自少掉對應的格；未宣告的欄位不會被轉送出去 | `backtestParameters`、`tradingStrategyWriteParameters`、`strategyBotWriteParameters` | `T/:483`、`T/:194`、`T/:249`，加上 `T/:502` `AFilledInLeverageNeverLeavesTheConnector`（**機械證明**：送了也出不去） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-2 | 目錄要主動說出的三件事 | 三句都在，且兩支重演讀同一份 | `spotOnlyReplayNote` | `T/:541`（三句逐句斷言 ＋ 斷言整段相同） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-3 | 說明只提成績單上真的有的格子 | 不提強平那一格；還在的照舊 | `liquidationReportCardNote` 刪除 | `T/:571`（含「仍描述 totalTransactionCost」的反向保險） | asserts-oracle | produces-oracle | ✅ conforms |
| BR-4 | 這個外掛不自己再擋一次那三個欄位 | `internal/` 沒有任何新的驗證 | `internal/` **0 檔案被動** | — | no-test | produces-oracle | 🟡 partial |

### Section 6 — Non-Functional Requirements

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| NFR-1 | 相容性：支數、名稱、路由不變，只有參數與說明變少 | 逐項相同 | 未改動任何工具宣告的前三個欄位 | `T/:87`、`T/:96` | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-2 | 一致性：目錄描述的每一件事，交易服務都做得到 | 目錄不再描述服務做不到的事 | 三個 catalog 檔 | `T/:194`、`T/:483`、`T/:571`（三個「不存在」的斷言合起來就是這一條） | asserts-oracle | produces-oracle | ✅ conforms |

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| — | 生產程式碼中 `tradingMode`／`leverage`／`maintenanceMargin`／`liquidation`／`曝險` 的命中數為 **0** | 無 orphan |

## Summary

- Conforms: **22/23** clauses ✅ (96%)
- Violations: 無
- Mis-asserted: 無
- Partial: **BR-4** 🟡 — 「這個外掛不自己再擋一次」是一條關於**沒有做什麼**的規則，
  它由 `internal/` 0 檔案被動來滿足。要用測試釘住它得寫一個「沒有新增驗證」的斷言，
  而那種斷言只會在下一次有人動 `internal/` 時變成噪音。**判斷：不補。**
- Gaps: 無
- Unclear: 無
- Orphans: 0

### 這一刀的驗收重點

這一刀改的**全是給助手讀的散文與一份參數清單**，兩者都沒有型別可以檢查。
所以測試以**缺席**為主要斷言對象——三個「那一格不在」、一個「那三個詞不在」、
一個「送了也出不去」。7 個 mutation 全數被殺，包含最容易漏的那一個：
**兩支重演各自抄一份說明**（改寫成看起來一樣的兩段文字，測試仍然紅）。

> **Ceiling.** 靜態一致性稽核：對照 PRD 推出的預期結果分別審測試與程式碼，
> 不靠跑整套測試下結論。這一刀無法用測試驗證的那一半，是**交易服務那一邊真的照這份目錄行為**——
> 那由上游切片的 `CONTRACT.md` 負責。
