# 助理做得了短線重演 — Architecture Design

**Status:** Confirmed（使用者授權一律採 best practice）
**Source PRD:** `.sdd/2026-09-24-short-term-replays/PRD.md`
**Tech context:** Go · MCP SDK · Clean / Onion（能力清單在 `cmd/server/tool_catalog*.go`，行為住在 `domains/`）

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 四件重演能力多兩格與新說明、等得比較久、成功回覆交給助理前先精簡；其他能力一行不動。
- **Guiding principle:** **一件能力「怎麼被對待」是那件能力自己的屬性。** 已經有 `Watching(waitLimit)` 這種修飾：
  即時更新那件能力自己說它要怎麼等。這一刀照同一個形狀加兩個修飾——`Waiting(responseWaitLimit)`（等多久）與
  `CondensingReplayResults()`（成功回覆要精簡）——服務與轉送的那一條路仍然只有一條，不為重演開分支。

---

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `domains/replay_result_domain.go` | **Add** | 一份重演結果的精簡：資金曲線取樣、交易明細只留最近的、分段遞迴；看不懂即原封 |
| `domains/api_tool_domain.go` | **Modify** | `Waiting(responseWaitLimit)`、`CondensingReplayResults()`、`Relayed(response)` |
| `vo/trading_service_request_vo.go` | **Modify** | 多 `ResponseWaitLimit`（0 即外掛預設） |
| `service/api_tool_service.go` | **Modify** | 回覆先經 `apiTool.Relayed(...)` 再轉成工具結果（含換新身分後重試那一條） |
| `infrastructure/tradingservice/trading_service_proxy.go` | **Modify** | 等待時間改成逐次請求決定：有 `ResponseWaitLimit` 用它，否則用預設 |
| `cmd/server/config.go` · `dependencies.go` · `tool_catalog*.go` | **Modify** | `TRADING_SERVICE_REPLAY_TIMEOUT_SECONDS`（預設 120）；四件重演多兩格、新說明、兩個修飾 |
| 其他能力、身分處理、登入 | **Not touched** | |

---

## 3. New Classes / Modules

| Name | Kind | Responsibility | Satisfies |
| :--- | :--- | :--- | :--- |
| `ReplayResultDomain` | Domain Model | 讀一份重演結果（JSON 物件）；`ToCondensedContent()`：`equityCurve` 超過 200 點時平均取 200 點（頭尾必留）並記 `equityCurvePointTotalCount`；`closedTrades` 超過 100 筆時只留最後 100 筆並記 `closedTradeTotalCount`；`inSample`／`validation` 照同一套遞迴；其他欄位（含成績單）原封；不是物件或解不開即原封 | US-03 |

```go
// 能力自己說怎麼等、回覆怎麼交；服務只多一行 apiTool.Relayed(response)。
domains.NewApiToolDomain(...).Waiting(replayWaitLimit).CondensingReplayResults()
apiTool.Relayed(response vo.TradingServiceResponseVo) vo.TradingServiceResponseVo
```

JSON 以 `map[string]json.RawMessage` / `[]json.RawMessage` 讀寫——不解出任何數字，所以成績單的精確小數原封不動，也不用任意型別。

---

## 4. Modified Components

| Component | Change |
| :--- | :--- |
| `ApiToolDomain.BuildRequest` | 帶上 `ResponseWaitLimit` |
| `ApiToolDomain.Relayed` | 只有標記了精簡、且回覆成功時才精簡；其餘原封 |
| `TradingServiceProxy` | `http.Client` 不再帶固定 Timeout；`Send` 以 `context.WithTimeout(ctx, 這次的等待時間)` 包住送出與讀取 |
| `backtestParameters()` | 多 `fillTiming`、`validationStartTime`（四件共用） |
| 四件重演的說明 | 新增一段共用的短線說明（成交時點、樣本外、五格、允許時間、精簡的兩個總數欄位）；拿掉「沒有資金曲線」那句錯話 |

---

## 5. Component Relationships

```mermaid
flowchart TD
    Catalog[tool_catalog: 四件重演 .Waiting().CondensingReplayResults()] --> Tool[ApiToolDomain]
    Svc[ApiToolService] --> Tool
    Svc --> Proxy[TradingServiceProxy]
    Tool -->|Relayed| Replay[ReplayResultDomain]
```

---

## 6. Extensibility & Handoff Notes

- **Next likely:** 其他會變大的回覆（例如長段 K 線）也要精簡。**Where:** 另一個像 `ReplayResultDomain` 的模型加一個修飾；服務不改。
- **Do not hardcode:** 等待時間讀設定；兩百點、一百筆是精簡模型的常數，改只改一處。
- **Known debt:** 被精簡掉的點與交易助理拿不回來——要時請它縮短期間。

---

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| US-01 照轉兩格／沒給／拒絕照轉 | `backtestParameters()` + 既有 `BuildRequest` + `Relayed`（拒絕原封） |
| US-01 說明引導 | 四件重演的共用說明 |
| US-02 全部 | `Waiting` + `ResponseWaitLimit` + `TradingServiceProxy` + config |
| US-03 全部 | `ReplayResultDomain` + `CondensingReplayResults` + `Relayed` |
