# 歷史同步的推定休市天數 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-10-02-history-sync-presumed-closed-days/PRD.md`
**Tech context:** Go · 外掛的能力目錄（`cmd/server/tool_catalog*.go`）

## 1. Design Goal & Guiding Principle

- **In one sentence:** 在 `trading_sync_k_candle_history` 與 `trading_get_k_candle_history_sync` 的說明加上推定休市天數（`presumedClosedDayCount`）的意思。
- **Guiding principle:** 說明跟著交易服務的回覆欄位名走，原樣用 `presumedClosedDayCount`，助理才對得上回覆裡看到的那一格。

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_market.go` | **Modify** | 兩個現貨歷史同步能力的說明 |
| 新測試 `cmd/server/tool_catalog_market_history_sync_test.go` | **Add** | 釘住說明裡的定義句（定義與名稱綁在同一個相鄰字串，避免刪掉定義仍綠） |
| `cmd/server/tool_catalog_contract.go` | **Not touched** | 合約沒有休市日 |
| `.sdd/UL-MAP.md` | **Modify** | 補「推定休市天數」一列（交易服務的能力） |

## 3. New Classes / Modules

無。

## 6. Extensibility & Handoff Notes

- 交易服務的輪次再多一個數字時，同一個說明、同一支測試的 `mustSay` 再加一句。

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| 說明列出推定休市天數並說出它的意思 | `trading_get_k_candle_history_sync` 說明 |
| 說明分開推定休市天數與略過根數 | 同上 |
| 發起同步的說明講假日 | `trading_sync_k_candle_history` 說明 |
| 發起同步的說明講連續太久的例外 | 同上 |
| 合約那邊不提推定休市 | 未改動合約說明；測試 `mustNotSay` |
