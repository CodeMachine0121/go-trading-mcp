# Contract Traceability Matrix — 合約訂單流（足跡）外掛說明

Contract: PRD.md · Design map: ARCH.md · Implementation: `cmd/server/` · Oracle: 5 clauses

> Static conformance audit; each test was falsified by removing the wording it pins.

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-1 | 四種參數種類 | 說明列出四種，各自說出是什麼與怎麼讀 | `tool_catalog.go` 參數說明 | `TestEveryParameterBoxNamesTheFourKindsTheSameWay` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-2 | 訂單流價位寬度只給合約 | 只給 contractKCandle、至多一個、大於零、最細 0.00000001 | 同上 | 同上（種類與定義綁成一句） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-3 | 原樣轉送 | 交易服務收到的參數一字不差 | 既有轉送（這次沒改） | `TestAnOrderFlowPriceStepIsForwardedUntouched` | asserts-oracle | produces-oracle | ✅ conforms（轉送是既有行為，這支測試釘住它，不是由這次的改動證偽的） |
| AC-4 | 足跡的每一項 | Footprint 四項、PriceLevel 四項 | `contractKCandleScriptNote` | `TestTheContractScriptNoteNamesTheFootprint` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-5 | 宣告了才有與不落地 | 沒宣告是空的；不存進資料庫、記憶中暫留，重啟或換分身會重拿 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-6 | 上限的算法與出路（US-03） | 日曆日、含回看與暖機、預設 200 根暖機與如何取代、交易策略取最大回看、出路、機器人會被停下 | `orderFlowReachNote`（合約算式說明、兩種合約重演、建立與修改機器人） | `TestTheOrderFlowReachIsExplainedWhereverItCanRefuse` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-7 | 每一格參數說明都一樣（US-03） | 六個 parameters 格都列四種；只有執行的那幾格提 parameterValues | `parameterKindsNote` + `parameterValuesCheckedNote` | `TestEveryParameterBoxNamesTheFourKindsTheSameWay`、`TestOnlyARunsParameterBoxTalksAboutThisTimesValues` | asserts-oracle | produces-oracle | ✅ conforms |

## Summary

- Conforms: 7/7 ✅ (100%) · Orphans: 0
