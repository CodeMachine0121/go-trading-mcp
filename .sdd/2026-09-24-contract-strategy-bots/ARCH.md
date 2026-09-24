# 助理養得起合約策略機器人 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-09-24-contract-strategy-bots/PRD.md`
**Tech context:** Go · MCP 外掛 · 能力清單即功能（`cmd/server/tool_catalog*.go` 宣告，`ApiToolDomain` 單一路徑轉達）

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 在策略機器人那一組能力上多一格 `marketDataKind`（建立／修改）、一格查詢用的 `marketDataKind`（列出），
  並改寫 `positionPlan`、`tradingStrategyId`、`symbol` 與幾段說明，讓助理分得清現貨機器人與合約機器人。
- **Guiding principle:** **照交易服務的入口決定外掛的形狀。** 交易服務兩種機器人共用 `/strategy-bots` 一組入口、靠一格種類分，
  外掛就一樣用一格——與策略腳本、交易策略的做法一致；交易服務另開入口時（合約 K 線、合約重演），外掛才另開帶 contract 的能力。
  另開 `trading_create_contract_strategy_bot` 會打同一個入口、要外掛替助理硬塞種類，而外掛的能力宣告刻意沒有「固定值」這種東西——
  加一個只為了這件事，等於在外掛裡長出一條交易服務沒有的規則。

---

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_automation.go` `strategyBotWriteParameters` | **Modify** | 加 `marketDataKind`（選填、body）；`tradingStrategyId` 與 `symbol` 說明寫出兩種各收什麼；`positionPlan` 說明加槓桿（只合約收）與合約建議的讀法，刪掉「沒有槓桿這一項」 |
| 同檔 `strategyBotApiTools` | **Modify** | `trading_create_strategy_bot` 說明提兩種機器人；`trading_update_strategy_bot` 說明寫出種類不給即保留；`trading_list_strategy_bots` 加查詢 `marketDataKind`；`trading_get_strategy_bot` 說明帶出種類與槓桿 |
| `cmd/server/tool_catalog_strategy.go` | **Modify** | `trading_create_trading_strategy` 說明與 `contractKCandleScriptNote` 改指向合約機器人 |
| `cmd/server/tool_catalog_contract_backtest_test.go` / `tool_catalog_test.go` | **Modify** | 舊的「只能是吃 K 線」斷言改成新說法 |
| 新測試 `cmd/server/tool_catalog_contract_strategy_bot_test.go` | **Add** | 釘住每一條 PRD 情境（說明句子、送出與不送出） |
| README 能力表 | **Modify** | 若有列機器人能力的說法則同步 |
| `internal/` 全部 | **Not touched** | 轉達路徑（`ApiToolDomain.BuildRequest`、`ToolArgumentsDomain.EncodedSubset`）已經「沒填就不送」、查詢欄位照送，不需要任何新行為 |
| 回應內容 | **Not touched** | 外掛原樣轉交機器人的回覆；交易服務把建議部位的鍵名改成小駝峰，外掛不解讀它，沒有東西依賴舊拼法 |

---

## 3. New Classes / Modules

無新型別。只新增一份測試檔。

---

## 4. Modified Components

| Component | Current role | Change needed |
| :--- | :--- | :--- |
| `strategyBotWriteParameters()` | 建立／修改共用的欄位 | 多 `bodyParameter("marketDataKind", string, …, false)`；三段說明改寫 |
| `trading_list_strategy_bots` | 無欄位 | `queryParameter("marketDataKind", string, …, false)` |
| `contractKCandleScriptNote` | 合約算式教學，最後一段說機器人掛不上 | 最後一段改成「吃合約行情的交易策略可以掛上合約機器人（marketDataKind 為 contractKCandle）」 |

---

## 5. Component Relationships

```mermaid
flowchart LR
    Assistant --> Catalog[apiToolCatalog]
    Catalog --> BotTools[strategyBotApiTools]
    BotTools --> Params[strategyBotWriteParameters]
    BotTools --> ApiTool[ApiToolDomain.BuildRequest]
    ApiTool --> TradingService[/strategy-bots]
```

---

## 6. Extensibility & Handoff Notes

- **Most likely next requirement:** 交易服務讓合約機器人多收別的設定（例如以標記價格當參考價、滑點）。
- **Where it lands:** `positionPlan` 的說明（外掛不拆開這一格，交易服務收什麼就原樣轉什麼），或 `strategyBotWriteParameters` 多一格。
- **How to add it:** 在清單補一格並在說明寫出會被什麼拒絕；轉達路徑不動。
- **Patterns applied & why:** 宣告式能力清單——新增是補一列，不是一條新分支。
- **Do not hardcode:** 不在外掛替助理填 `kCandle`；不在外掛檢查槓桿、名單或種類。
- **Known debt / deferred:** 啟動／停止／刪除／跑一輪／執行紀錄說明不分種類，因為交易服務對兩種一視同仁。

---

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| US-01 建一台合約機器人 | `marketDataKind` 欄位 + `tradingStrategyId`／`symbol` 說明 |
| US-01 沒說種類即現貨 | `EncodedSubset` 不送沒填的 + `marketDataKind` 說明 |
| US-01 種類不符／標的不在名單 由交易服務拒絕 | 既有轉達路徑（原話帶回）+ `symbol` 說明 |
| US-02 只改名字不帶種類／換種類 | 同一格 + `trading_update_strategy_bot` 說明 |
| US-03 給槓桿／不給槓桿／小於一／現貨給槓桿 | `positionPlan` 說明；`positionPlan` 原樣轉達 |
| US-04 只列合約／不指定 | `trading_list_strategy_bots` 查詢欄位 + 說明 |
| US-05 建交易策略說明／合約算式教學 | `tool_catalog_strategy.go` 兩處文字 |

---

## 8. Risks & Open Decisions

- **Risks / trade-offs:** 交易服務那一刀上線前，助理送出合約種類會被交易服務以舊規則拒絕；原話帶回，不會壞。
- **Open decisions (for implementation):** 無。
- **環境變數：** 無新增。
