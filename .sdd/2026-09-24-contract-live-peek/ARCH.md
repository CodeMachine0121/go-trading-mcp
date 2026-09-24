# 看一眼合約即時更新 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-09-24-contract-live-peek/PRD.md`
**Tech context:** Go · MCP go-sdk · catalogue of `ApiToolDomain` entries in `cmd/server/tool_catalog*.go`

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** add one catalogue entry `trading_peek_live_contract_k_candle` → `GET /contract-k-candles/live?symbol=`, marked `.Watching(liveUpdateWaitLimit)` so the existing proxy path reads the first SSE `data:` event, exactly as the spot peek does.
- **Guiding principle:** no new mechanism. Watching an SSE stream is already a property any ability can carry (`ApiToolDomain.Watching`), consumed by `TradingServiceProxy`; the contract twin is data, not code. The backend opened a separate route, so by this repo's precedent it is a separate `contract` ability.

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_contract.go` | **Modify** | `contractApiTools(liveUpdateWaitLimit)` takes the wait limit and adds the new ability after the contract history-sync abilities |
| `cmd/server/tool_catalog.go` | **Modify** | pass `liveUpdateWaitLimit` to `contractApiTools` |
| catalogue tests | **Modify** | the contract ability list, the sign-in map, the "only watching abilities stay on the line" rule (now two), the address table |
| `cmd/server/tool_catalog_contract_live_peek_test.go` | **Add** | description and request assertions for every PRD scenario |
| README | **Modify** | one line in the contract abilities section |
| `TradingServiceProxy`, `ApiToolDomain`, controller | **Not touched** | the SSE read, the empty-wait sentence and the refusal pass-through already serve any watching ability |
| `trading_peek_live_k_candle` | **Not touched** | spot wording stays verbatim |

## 3. New Classes / Modules

None — one catalogue entry.

## 4. Modified Components

| Component | Current role | Change needed |
| :--- | :--- | :--- |
| `contractApiTools` | builds contract abilities | new parameter `liveUpdateWaitLimit time.Duration`; new entry, read verb, no sign-in, query `symbol` required |

## 5. Component Relationships

```mermaid
flowchart LR
    Assistant --> MCP[McpController] --> Catalog[trading_peek_live_contract_k_candle .Watching]
    Catalog --> Proxy[TradingServiceProxy watch path] -->|SSE first data line| Backend[/contract-k-candles/live]
```

## 6. Extensibility & Handoff Notes

- **Most likely next requirement:** a live mark-price peek once the venue offers it — another watching entry, or a description change if the backend adds fields to the same stream.
- **Where it lands:** the catalogue only. **Do not hardcode** the wait limit — it is `LIVE_UPDATE_WAIT_LIMIT_SECONDS`, shared with spot.

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| 看一眼合約標的 | new entry → `/contract-k-candles/live` with `symbol` query, watching |
| 現貨的看一眼照舊 | `trading_peek_live_k_candle` unchanged (pinned by test) |
| 等滿沒有更新 | existing proxy empty-wait sentence (watching path) |
| 不在合約追蹤名單上 / 系統不認得 | trading service refusal passed through; description names `trading_add_to_contract_watchlist` |
| 成形中 / 走完了 / 即時更新斷了 / 內容是最新價 | description text |

## 8. Risks & Open Decisions

- Until go-trading#70 is deployed the route answers "not found"; the refusal is passed through as is.
