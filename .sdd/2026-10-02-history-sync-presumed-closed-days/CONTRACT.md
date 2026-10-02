# Contract Traceability Matrix — 歷史同步的推定休市天數

Contract: PRD.md
Design map: ARCH.md
Implementation: `cmd/server/tool_catalog_market.go`
Oracle: Acceptance Criteria (4 clauses)

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-1 | 說明列出推定休市天數並說出它的意思 | 「看同步走到哪」的說明列出 `presumedClosedDayCount`、說它是被跳過的平日數（來源說沒資料）、說它不算來源不答話 | `tool_catalog_market.go` `trading_get_k_candle_history_sync` | `tool_catalog_market_history_sync_test.go` 前三句 `mustSay` | asserts-oracle（定義與名稱在同一個相鄰字串；植入「拿掉不算來源不答話」會紅） | produces-oracle | ✅ conforms |
| AC-2 | 說明分開推定休市天數與略過根數 | 說兩者是兩件事、略過根數是來源答了但某幾根不合格 | 同上 | 第四句 `mustSay`（植入「把略過根數講成沒資料的天數」會紅） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-3 | 發起同步的說明講假日 | 國定假日被跳過、記在推定休市天數、不會讓同步停下 | `trading_sync_k_candle_history` | `mustSay`（植入「會讓同步停下」會紅） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-4 | 合約那邊不提推定休市 | 兩個合約歷史同步說明裡沒有推定休市天數 | 未改動 `tool_catalog_contract.go` | `mustNotSay` | asserts-oracle | produces-oracle | ✅ conforms |

## Orphans (code with no clause)

無。

## Summary

- Conforms: 4/4 clauses ✅ (100%)
- Violations / Mis-asserted / Partial / Gaps / Unclear: —
- Orphans: 0

Note: static conformance audit against the Acceptance Criteria — it judges test assertions and code paths against the spec's expected outcome, not by running the full suite.
