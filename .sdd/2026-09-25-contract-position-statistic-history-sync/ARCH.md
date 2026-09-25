# 告訴助理：合約歷史同步也補持倉統計 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-09-25-contract-position-statistic-history-sync/PRD.md`
**Tech context:** Go · MCP server · 能力目錄（`cmd/server/tool_catalog_*.go`）

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 只改五份能力說明的文字，讓助理讀到的與交易服務的行為一致。
- **Guiding principle:** 說明就是這個外掛對助理的契約；每一句「送出去才知道」的事，由 `TestTheContractAbilitiesSayWhatCannotBeDiscoveredBySending` 的 `mustSay` 釘住，另加一組 `mustNotSay` 擋住過時的說法。

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_contract.go` | **Modify** | 四份說明：同步、看進度、查持倉統計、手動補齊 |
| `cmd/server/tool_catalog_strategy.go` | **Modify** | 合約策略腳本「沒有值一律是零」那一句 |
| `cmd/server/tool_catalog_contract_test.go` | **Modify** | `mustSay` 補新句子；新增「不得再說」的測試 |
| `cmd/server/tool_catalog_strategy_test.go`（或同檔） | **Modify** | 合約腳本說明不再說只留三十天 |
| 能力名稱、路徑、參數 | **Not touched** | PRD 明定不變；既有的目錄測試照舊守住 |

## 3. New Classes / Modules

無。

## 4. Modified Components

| Component | Current role | Change needed |
| :--- | :--- | :--- |
| `trading_sync_contract_k_candle_history` 說明 | 說 K 線怎麼同步 | 加一段持倉統計：同一個回溯天數、先 K 線後持倉統計、只存沒有的、沒有檔案不算失敗、歷史資料庫不答話只停這一份 |
| `trading_get_contract_k_candle_history_sync` 說明 | 說兩種原因 | 加 `positionStatistic` 五項、與 K 線那組分開不加總、它自己的 `fetchFailureReason` |
| `trading_list_contract_position_statistics` 說明 | 說只留三十天、更早沒人錄 | 改為即時錄製從加入前三十天起、更早的用合約歷史同步補 |
| `trading_backfill_contract_k_candles` 說明 | 說只補 K 線 | 加一句：更早的持倉統計用合約歷史同步 |
| 合約策略腳本共用說明 | 說持倉統計只留三十天 | 改為只有錄到或同步過的那段才有值 |

## 5. Component Relationships

N/A（純文字）。

## 6. Extensibility & Handoff Notes

- **Most likely next requirement:** 交易服務又替同步多補一種資料。
- **Where it lands:** 同一份同步說明多一句、看進度說明多一組；`mustSay` 多一條。

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| US-01 同步的說明說出同一趟也補持倉統計 | 同步說明 + `mustSay` |
| US-01 沒有那一天的檔案不算失敗 | 同步說明 + `mustSay` |
| US-01 看進度的說明說出持倉統計那一組分開 | 看進度說明 + `mustSay` |
| US-01 歷史資料庫不答話不是失敗 | 看進度說明 + `mustSay` |
| US-02 查持倉統計的說明指向合約歷史同步 | 查持倉統計說明 + `mustSay` / `mustNotSay` |
| US-02 寫合約策略腳本的說明不再說只留三十天 | 策略說明 + `mustSay` / `mustNotSay` |
| US-02 手動補齊的說明指向合約歷史同步 | 手動補齊說明 + `mustSay` |
| US-02 能力清單不變 | 既有目錄測試（名稱、路徑、參數） |

## 8. Risks & Open Decisions

- 說明與交易服務要一起上線。
