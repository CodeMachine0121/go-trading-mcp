# Contract Verification — 管理員標的篩選（外掛）

**Contract:** `.sdd/2026-10-19-admin-symbol-screening/PRD.md`（Section 3 Acceptance Criteria）
**Implementation:** `cmd/server/tool_catalog_symbol_screening.go`、`cmd/server/tool_catalog.go`；轉達路徑沿用 `ApiToolDomain.BuildRequest`／`Relayed` 與 `TradingServiceProxy.responseOf`
**Ceiling:** static conformance audit — 對照規格的預期結果判斷測試斷言與程式路徑，不執行自創情境。

## Clauses

| ID | Clause | Oracle（只由規格推得） | Code | Test | Test audit | Code audit | Jev | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-1 | 管理員透過助理送出篩選 | 帶回一筆排隊中的篩選與編號；說明要求說出編號 | `tool_catalog_symbol_screening.go:104` | `TestEveryScreeningAbilityAsksTheRightPlace`、`TestSubmittingAScreeningNeedsAProfileAndSymbols`、`TestEveryScreeningAbilityHandsBackWhatTheTradingServiceSaidWordForWord` | asserts-oracle | produces-oracle | asserts-oracle 0.73 / produces-oracle 0.41（low confidence） | ✅ conforms |
| AC-2 | 非管理員被擋下，助理照原話轉達 | 帶回「這是管理員功能」原話；說明要求照原話轉達、不重試 | `tool_catalog_symbol_screening.go:10`、`trading_service_proxy.go:241` | `TestEveryScreeningAbilitySaysOnlyAdministratorsMayUseItAndNotToRetry`、`TestEveryScreeningAbilityHandsBack…`、proxy 測試 403 → 拒絕 | asserts-oracle | produces-oracle | asserts-oracle 0.99 / produces-oracle 0.88 | ✅ conforms |
| AC-3 | 我是誰說出是不是管理員 | 我是誰的說明寫明會說出是不是管理員 | `tool_catalog.go:228` | `TestAskingWhoIAmSaysWhetherIAmAnAdministrator` | asserts-oracle | produces-oracle | asserts-oracle 0.83 / produces-oracle 0.68（recheck：重讀無誤） | ✅ conforms |
| AC-4 | 大小寫與全形逗號原樣送出 | 交易服務收到的清單一字不改 | `tool_catalog_symbol_screening.go:118` + `BuildRequest` | `TestSubmittingAScreeningSendsThePastedListUntouched` | asserts-oracle | produces-oracle | asserts-oracle 1.00 / produces-oracle 0.39（low confidence） | ✅ conforms |
| AC-5 | 正好二十個標的照常送出 | 二十個標的一字不改送出 | 同上 | 同上（exactly twenty） | asserts-oracle | produces-oracle | asserts-oracle 0.79 / produces-oracle 0.23（low confidence） | ✅ conforms |
| AC-6 | 不允許的字元由交易服務拒絕並原話帶回 | 帶回交易服務拒絕原話 | 同上 + `responseOf` | `…SendsThePastedListUntouched`、`…HandsBack…` | asserts-oracle | produces-oracle | asserts-oracle 0.95 / produces-oracle 0.92 | ✅ conforms |
| AC-7 | 說要等就把最多等幾秒帶給交易服務 | 要求看篩選 9、最多等 30 秒 | `tool_catalog_symbol_screening.go:136` + `BuildRequest` | `TestLookingAtAScreeningCanWaitAWhile` | asserts-oracle | produces-oracle | asserts-oracle 0.96 / produces-oracle 0.49（low confidence） | ✅ conforms |
| AC-8 | 沒說要等就立刻問 | 要求看篩選 9、沒有等 | 同上 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.97 / produces-oracle 0.79（recheck：重讀無誤） | ✅ conforms |
| AC-9 | 只給必要項目 | 只收到名稱、策略腳本、看盤間隔 | `tool_catalog_symbol_screening.go:26`、`:72` + `BuildRequest` | `TestCreatingAProfileSendsOnlyWhatTheUserSaid` | asserts-oracle | produces-oracle | asserts-oracle 0.93 / produces-oracle 0.99 | ✅ conforms |
| AC-10 | 明確給出等於預設的歷史天數 | 也收到歷史天數 730 | 同上 | 同上（斷言值 730） | asserts-oracle | produces-oracle | asserts-oracle 0.71 / produces-oracle 0.56（low confidence） | ✅ conforms |
| AC-11 | 看盤間隔太細由交易服務拒絕 | 帶回交易服務拒絕原話 | 同上 + `responseOf` | `…SendsOnlyWhatTheUserSaid`、`…HandsBack…` | asserts-oracle | produces-oracle | asserts-oracle 0.91 / produces-oracle 0.81（recheck：重讀無誤） | ✅ conforms |
| AC-12 | 看一筆篩選的說明教助理怎麼讀結論 | 說明解釋四種結論、兩道關卡、組合貢獻、推薦第一名、通過不是買賣建議 | `tool_catalog_symbol_screening.go:15` | `TestLookingAtAScreeningTeachesHowToReadTheVerdicts` | asserts-oracle | produces-oracle | asserts-oracle 0.85 / produces-oracle 0.96 | ✅ conforms |

**首輪發現（已修正）：** 第一輪 Jev 判 AC-2、AC-6、AC-11 的測試 shallow（篩選能力沒有直接驗證拒絕原話經過它被帶回），AC-10 只驗欄位名、沒驗 730。已補 `TestEveryScreeningAbilityHandsBackWhatTheTradingServiceSaidWordForWord` 並讓 AC-10 斷言值，兩者都經突變確認會紅；再審全數 conforms。首輪另有幾條 `not-implemented`，原因是只把宣告行交給 Jev、沒附共用的組請求程式 `ApiToolDomain.BuildRequest`，補上後一致。

## Orphans

| Behavior | Where | Classification |
| :--- | :--- | :--- |
| 列出／查看／修改／刪除篩選設定、列出／取消篩選的能力 | `tool_catalog_symbol_screening.go` | 屬於 PRD §1 Expected Outcome 的九件能力與 §4 Core Business Rules，沒有 Gherkin 情境；路由由 `TestEveryScreeningAbilityAsksTheRightPlace` 釘住 |

沒有任何行為落在 Out of Scope（開通管理員、外掛自己檢查規則／整理清單／補預設值、自動加入追蹤）。

## Summary

✅ 12 conforms · 🔴 0 violations · 🟠 0 mis-asserted · 🟡 0 partial · ❌ 0 gaps · ❔ 0 unclear · ⚠️ 1 orphan group（契約內的能力、無情境）
Conformance: 100%

Jev: ran on 12 clauses. Needs James（route escalate, low confidence, 但與我的判斷一致）：AC-1、AC-4、AC-5、AC-7、AC-10。
