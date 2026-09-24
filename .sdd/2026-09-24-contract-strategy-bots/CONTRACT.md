# Contract Traceability Matrix — 2026-09-24-contract-strategy-bots

Contract: PRD.md
Design map: ARCH.md
Implementation: `cmd/server/tool_catalog_automation.go`, `cmd/server/tool_catalog_strategy.go`（轉達路徑 `internal/domain/models/domains/api_tool_domain.go` 不變）
Oracle: Acceptance Criteria (14 clauses) + Business Rules (2)

> Static conformance audit — it judges test assertions and code paths against the spec's expected outcome; it does not execute invented scenarios.

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-1 | 建一台合約機器人 | 送出內容帶「合約行情」種類；說明寫出合約機器人只能引用合約交易策略、標的要先在合約追蹤名單上 | tool_catalog_automation.go:30,33,36 | tool_catalog_contract_strategy_bot_test.go `TestWritingAStrategyBotForwardsTheKindOnlyWhenGiven`/given、`TestWritingAStrategyBotSaysWhichKindItIs` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-2 | 沒說種類即現貨 | 送出內容沒有種類；說明寫出不給就是現貨機器人 | automation.go:30；api_tool_domain.go `BuildRequest` 只送有填的 | `…ForwardsTheKindOnlyWhenGiven`/left out、`…SaysWhichKindItIs`（「建立時不給就是 kCandle」「現貨機器人」） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-3 | 種類與交易策略不符由交易服務拒絕 | 拒絕原話原樣帶回 | 既有轉達路徑 | application/tests `TestARefusalComesBackInTheTradingServicesOwnWords`；controller/tests `TestARefusalReachesTheAssistantInTheTradingServicesOwnWordsAndIsMarkedAsOne` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-4 | 標的不在合約追蹤名單上由交易服務拒絕 | 原話帶回；說明事先寫出要先加進合約追蹤名單 | automation.go:33 | `…SaysWhichKindItIs`（「必須已經在合約追蹤名單上」「trading_add_to_contract_watchlist」）＋ AC-3 的轉達測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-5 | 只改名字不帶種類 | 送出內容沒有種類；修改說明寫出不給即保留、換另一種會被拒絕 | automation.go:101 | `…ForwardsTheKindOnlyWhenGiven`/update left out、`TestRewritingAStrategyBotSaysTheKindIsKept` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-6 | 換種類由交易服務拒絕 | 送出內容帶「合約行情」；拒絕原話帶回 | automation.go:30 | `…ForwardsTheKindOnlyWhenGiven`/update given ＋ AC-3 轉達測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-7 | 給合約機器人槓桿 | 槓桿 5 在建議部位裡原樣送出；說明寫出只有合約收、不給即一倍、小於一或超過上限被拒 | automation.go:54 | `TestAContractBotsLeverageTravelsInsideItsPlan`、tool_catalog_test.go `TestWritingAStrategyBotSaysHowToSizeAPosition` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-8 | 不給槓桿 | 送出的建議部位裡沒有槓桿 | 轉達路徑原樣 | `TestAContractBotsLeverageTravelsInsideItsPlan`（left out） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-9 | 槓桿小於一 | 原話帶回 | 轉達路徑 | AC-3 轉達測試；說明「小於一會被拒絕」由 `…HowToSizeAPosition` 釘住 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 現貨機器人給槓桿 | 原話帶回；說明事先寫出現貨給大於一倍會被拒絕 | automation.go:54 起 | `…HowToSizeAPosition`（「現貨機器人給大於一倍的槓桿會被拒絕」）＋轉達測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-11 | 只列合約機器人 | 問交易服務時帶「只要合約行情」；說明寫出可以只列一種、每一台都帶著種類 | automation.go:86-89 | `TestListingStrategyBotsCanAskForOneKind` | asserts-oracle（修正後加上「只列其中一種」斷言） | produces-oracle | ✅ conforms |
| AC-12 | 不指定即全部 | 問的時候沒有帶種類 | automation.go:88 | `TestListingStrategyBotsCanAskForOneKind`（everyRequest.Query 為空） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-13 | 建交易策略的說明 | 寫出可以掛上合約機器人；不再寫「策略機器人目前掛不上它」 | tool_catalog_strategy.go:128-130 | `TestCreatingATradingStrategySaysAContractOneFitsAContractBot` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-14 | 合約算式教學 | 寫出可以掛上合約機器人；不再寫「策略機器人目前只跑 K 線」 | tool_catalog_strategy.go:177 起（最後一段） | contract_script_test.go `TestTheContractScriptNoteSaysWhereItCanGo` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-1 | 外掛只轉達：沒填不送、拒絕原話帶回 | 同左 | api_tool_domain.go `BuildRequest`／tool_arguments_domain.go `EncodedSubset` | domains/tests `TestBuildRequestLeavesOutBoxesThatWereNotFilledIn` ＋ 轉達測試 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-2 | 種類與槓桿選填，不替任何一格填預設值 | 兩格 IsRequired 為否；沒填不出現在送出內容 | automation.go:30,88 | `…SaysWhichKindItIs`（IsRequired false）、`…ForwardsTheKindOnlyWhenGiven`、`…LeverageTravelsInsideItsPlan` | asserts-oracle | produces-oracle | ✅ conforms |

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| automation.go:102 | 修改說明：「合約機器人的 positionPlan 是整組改寫，沒帶 leverage 就回到一倍」 | reconciled — 已補進 PRD §4 業務規則 |
| automation.go:68-71 | 建立說明：合約機器人照交易模式說做多／做空／平多／平空；止損止盈對帳用 trading_backtest_contract_trading_strategy | reconciled — 已補進 PRD §4 業務規則 |
| automation.go:54 | 形狀範例刻意不含 leverage，槓桿只在寫給合約機器人的那一句出現 | reconciled — 已補進 PRD §4 業務規則 |

## Summary

- Conforms: 16/16 clauses ✅ (100%，AC-11 的斷言已補上)
- Violations: —
- Mis-asserted: —（AC-11 已修正）
- Partial: —
- Gaps: —
- Unclear: —
- Orphans: 3，皆已補進 PRD 業務規則（無 Out of Scope 違規）
