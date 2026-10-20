# Contract Traceability Matrix — 一台機器人盯多個交易標的（外掛）

Contract: PRD.md · Design map: ARCH.md · Implementation: `cmd/server/`（`internal/` 未動）
Oracle: Acceptance Criteria (21 scenarios) + Business Rules (7) + NFR (2) = 30 clauses

Path abbreviations: `auto` = `cmd/server/tool_catalog_automation.go`, `con` = `cmd/server/tool_catalog_contract.go`,
`str` = `cmd/server/tool_catalog_strategy.go`, `MS` = `cmd/server/tool_catalog_multi_symbol_strategy_bot_test.go`,
`CB` = `cmd/server/tool_catalog_contract_strategy_bot_test.go`, `OF` = `cmd/server/tool_catalog_order_flow_test.go`,
`BA` = `cmd/server/tool_catalog_binance_auto_order_test.go`, `AT` = `internal/application/tests/api_tool_application_test.go`,
`AD` = `internal/domain/models/domains/api_tool_domain.go`（`BuildRequest`：只轉送宣告過且填了的名字，陣列原樣轉送）。

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev | Status |
|----|--------|------------------------|------|------|------------|------------|-----|--------|
| AC-01-1 | 建立盯兩個標的的機器人 | 送出 BTCUSDT、ETHUSDT；回覆的兩個監聽標的讀得出沒有上一次送出的信號 | auto:35 `symbols` 陣列；AD；auto:116 回覆說明 | MS TestWritingAStrategyBotForwardsTheSymbolListAsGiven；MS TestEveryAbility…（「沒送過就沒有」） | asserts-oracle | produces-oracle | shallow 0.59 / produces 0.92 → escalate | ✅ conforms（Jev 指出回覆那一半未釘，已補斷言） |
| AC-01-2 | 只盯一個標的 | 只送 BTCUSDT | auto:35；AD | MS Forwards…（`["BTCUSDT"]`） | asserts-oracle | produces-oracle | asserts 0.49 / produces 0.67 → escalate (low conf) | ✅ conforms |
| AC-01-3 | 正好 10 個標的 | 10 個原樣送出 | auto:35；AD | MS Forwards…（10 個） | asserts-oracle | produces-oracle | asserts 0.57 / produces 0.77 → escalate (low conf) | ✅ conforms |
| AC-01-4 | 超過 10 個標的 | 外掛不擋、原樣送出；拒絕原話帶回；說明事先寫最多 10 個 | auto:35；AD；既有拒絕轉交路徑 | MS Forwards…（11 個）、MS TestWritingAStrategyBotAsksForARequiredListOfSymbols；AT TestARefusalComesBackInTheTradingServicesOwnWords | asserts-oracle | produces-oracle | shallow 0.35 / produces 0.27 → escalate | ✅ conforms（拒絕轉交是既有、已有測試的共用路徑） |
| AC-01-5 | 重複的標的 | 原樣送出；說明寫大小寫不同視為重複 | auto:35；AD | MS Forwards…（BTCUSDT、btcusdt）、MS AsksForARequiredList…「大小寫不同視為同一個」 | asserts-oracle | produces-oracle | asserts 0.57 / not-implemented 0.24 → escalate | ✅ conforms（Jev 信心低；判定靠交易服務，外掛只轉送與說明） |
| AC-01-6 | 合約機器人有標的不在追蹤名單 | 說明要每個標的在追蹤名單、指向加入能力；拒絕原話帶回 | auto:35 | MS AsksForARequiredList…；CB TestWritingAStrategyBotSaysWhichKindItIs | asserts-oracle | produces-oracle | asserts 0.35 / unclear 0.32 → escalate | ✅ conforms |
| AC-01-7 | 總曝險提醒 | 建立說明：每個標的各自照整份、總曝險最多 N 倍 | auto:126 `strategyBotPerSymbolExposureNote` | MS TestWritingAStrategyBotSaysEachSymbolTakesTheWholePositionPlanAndReallyTrades | asserts-oracle | produces-oracle | asserts 0.87 / produces 0.93 → accept | ✅ conforms |
| AC-02-1 | 加一個標的 | 完整清單原樣送出；說明寫留下保留、新加空白 | auto:35 | MS AsksForARequiredList…、Forwards… | asserts-oracle | produces-oracle | asserts 0.04 / produces 0.21 → escalate (low conf) | ✅ conforms |
| AC-02-2 | 只改別的設定時照抄標的 | 修改說明：照抄、漏列即移除 | auto:202 | MS TestRewritingAStrategyBotSaysTheListIsCompleteAndWhyARemovalIsRefused | asserts-oracle | produces-oracle | asserts 0.65 / produces 0.63 → recheck | ✅ conforms |
| AC-02-3 | 移除一個空手的標的 | 只送 BTCUSDT；說明寫過去紀錄保留 | auto:35 | MS AsksForARequiredList…「過去的執行紀錄仍保留」 | asserts-oracle | produces-oracle | asserts 0.23 / produces 0.29 → escalate (low conf) | ✅ conforms |
| AC-02-4 | 移除還有機器人持倉的標的 | 整次被拒（持倉）；說明要先空手、不重送 | auto:203-206 | MS Rewriting… | asserts-oracle | produces-oracle | asserts 0.91 / produces 0.61 → recheck | ✅ conforms |
| AC-02-5 | 移除還有沒做完自動下單的標的 | 整次被拒（自動下單沒做完）；等它做完 | auto:203-206 | MS Rewriting… | asserts-oracle | produces-oracle | asserts 0.93 / produces 0.51 → escalate (low conf) | ✅ conforms |
| AC-02-6 | 舊的單一標的寫法不再送出 | 單一標的那一格不送出 | auto:27-45（未宣告 `symbol`）；AD | MS TestWritingAStrategyBotNeverForwardsASingleSymbol | asserts-oracle | produces-oracle | asserts 0.92 / produces 0.91 → accept | ✅ conforms |
| AC-03-1 | 一個標的停擺、另一個有倉 | 說明讓助理說出各標的持倉、停擺、其他照常、仍在跑 | auto:116-136 `strategyBotReplyNote`（七件能力） | MS TestEveryAbilityAnsweringWithABotSaysHowToReadEachWatchedSymbol | asserts-oracle | produces-oracle | asserts 0.57 / produces 0.21 → escalate (low conf) | ✅ conforms |
| AC-03-2 | 全部標的都停擺 | 讀機器人上的停擺原因、說整台停擺 | auto:121 | MS EveryAbility… | asserts-oracle | produces-oracle | asserts 0.81 / produces 0.21 → escalate (low conf) | ✅ conforms |
| AC-03-3 | 讓停擺的標的恢復 | 停止再啟動即恢復、持倉不動 | auto:123 | MS EveryAbility… | asserts-oracle | produces-oracle | asserts 0.47 / produces 0.39 → escalate (low conf) | ✅ conforms |
| AC-04-1 | 只看 ETHUSDT | 查詢帶 symbol=ETHUSDT | auto:247 `queryParameter("symbol")`；AD | MS TestReadingABotsRoundsCanAskForOneSymbol | asserts-oracle | produces-oracle | asserts 0.85 / produces 0.71 → recheck | ✅ conforms |
| AC-04-2 | 讀全部標的 | 不帶篩選；每筆寫著標的 | auto:241、247 | MS CanAskForOneSymbol（空查詢）、MS TestReadingABotsRoundsSaysEachRecordNamesItsSymbol | asserts-oracle | produces-oracle | asserts 0.97 / produces 0.44 → escalate (low conf) | ✅ conforms |
| AC-04-3 | 篩選的標的格式不對 | 說明寫會被拒絕；拒絕原話帶回 | auto:247-249；既有拒絕轉交路徑 | MS CanAskForOneSymbol「格式不對會被拒絕」；AT TestARefusalComesBack… | asserts-oracle | produces-oracle | asserts 0.84 / not-implemented 0.12 → escalate | ✅ conforms（拒絕轉交是共用路徑） |
| AC-05-1 | 移出追蹤名單前查哪些機器人盯著它 | 說明看監聽標的、只那個標的失明、其他照常 | con:218-220 | MS TestRemovingAContractFromTheWatchlistBlindsOnlyThatWatchedSymbol；CB TestRemovingAContractFromTheWatchlistWarnsOfTheBotsWatchingIt | asserts-oracle | produces-oracle | asserts 0.76 / produces 0.23 → escalate (low conf) | ✅ conforms |
| AC-05-2 | 自動下單的說明涵蓋每個標的 | 說明：每一個監聽標的都會真的下單 | auto:126-128 | MS EachSymbolTakesTheWholePositionPlanAndReallyTrades；BA TestEveryAbilityThatAnswersWithAStrategyBotSaysAutoOrderReallyTrades | asserts-oracle | produces-oracle | asserts 0.79 / produces 0.76 → recheck | ✅ conforms |
| BR-1 | 只收一串監聽標的、原樣轉送、不檢查 | 必填陣列；無外掛檢查 | auto:35；AD | MS AsksForARequiredList…、TestWritingAStrategyBotWithoutSymbolsIsNotSent、Forwards… | asserts-oracle | produces-oracle | not run | ✅ conforms |
| BR-2 | 修改是完整清單；移除被拒照原話、不重送 | 同 AC-02-* | auto:35、202-206 | MS Rewriting… | asserts-oracle | produces-oracle | not run | ✅ conforms |
| BR-3 | 讀機器人：每個標的欄位、整台 vs 標的停擺 | 同 AC-03-* | auto:116-136 | MS EveryAbility… | asserts-oracle | produces-oracle | not run | ✅ conforms |
| BR-4 | 啟動清狀態、立刻跑一輪每個標的都跑 | 說明含兩者 | auto:123-124 | MS EveryAbility… | asserts-oracle | produces-oracle | not run | ✅ conforms |
| BR-5 | 執行紀錄有標的、可篩、編號不重複、各 50 輪 | 說明與查詢欄位 | auto:241-249 | MS RoundsSaysEachRecord…、CanAskForOneSymbol | asserts-oracle | produces-oracle | not run | ✅ conforms |
| BR-6 | 部位規劃各自照整份、每個標的都真的下單 | 同 AC-01-7、AC-05-2 | auto:126 | MS EachSymbolTakes… | asserts-oracle | produces-oracle | not run | ✅ conforms |
| BR-7 | 合約標的行情停了只跳過那一個標的 | 讀紀錄說明：只有那個標的跳過、其他照常 | auto:111-113 | MS TestASkippedContractRoundIsOnlyThatSymbols | asserts-oracle | produces-oracle | not run | ✅ conforms |
| NFR-1 | 不保留單一標的寫法 | `symbol` 不宣告、不轉送 | auto:27-45 | MS NeverForwardsASingleSymbol、AsksForARequiredList…（未宣告） | asserts-oracle | produces-oracle | not run | ✅ conforms |
| NFR-2 | 自動下單仍只能在網頁切換 | 建立／修改不宣告 autoOrderEnabled | auto（未變） | BA TestWritingAStrategyBotNeverForwardsAutoOrder | asserts-oracle | produces-oracle | not run | ✅ conforms |

## Orphans

| Behavior | Where | Classification |
| :--- | :--- | :--- |
| 訂單流超過上限的說法改成「那個監聽標的停擺（orderFlowRefused），只盯一個或全部停擺時整台停擺」 | str `orderFlowReachNote`；OF | 不在 PRD 情境內，但屬 US-05「其他說明不再說機器人只盯一個標的」的同一類更正；交易服務 PRD 明定訂單流超過範圍只停那個標的。保留 |
| README 新增〈一台機器人盯多個交易標的〉一節 | README.md | 文件，非行為 |

## Summary

- ✅ 30 conforms · 🔴 0 violations · 🟠 0 mis-asserted · 🟡 0 partial · ❌ 0 gaps · ❔ 0 unclear · ⚠️ 1 orphan（保留）
- Conformance: 100%
- Jev ran on 21 AC clauses (BR/NFR not run — they restate AC clauses). Jev's escalations are mostly low-confidence agreements; its disagreements (AC-01-4/5/6, AC-04-3) treat the trading service's refusal as something the connector must implement — by design (PRD Out of Scope) the connector forwards unchanged and relays the refusal through the shared, already-tested path. AC-01-1's "shallow" flag was valid for the reply half and was fixed by pinning 「沒送過就沒有」.
- Ceiling: static conformance audit; it judges assertions and code paths against the spec, it does not execute invented scenarios.
