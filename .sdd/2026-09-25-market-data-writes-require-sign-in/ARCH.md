# 改行情的事要帶身分 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-09-25-market-data-writes-require-sign-in/PRD.md`
**Tech context:** Go · MCP server · 能力目錄（`cmd/server/tool_catalog_*.go`）

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 把十六件改行情的能力宣告從「不需要身分」翻成「需要身分」，其餘一格不動。
- **Guiding principle:** 需不需要身分是**能力宣告上的一格**（`NewApiToolDomain` 的 `requiresSignIn`），
  帶不帶身分、沒身分就不去問、拒絕原話帶回，全由既有的 `ApiToolDomain` / `ApiToolApplication` 一條路處理。
  這一刀不新增任何分支。

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_market.go` | **Modify** | 現貨八件：新增／改／刪 K 線、手動補齊、同步歷史、看同步進度、加進／移出觀察清單 → `true`；合約指標計算旁那段「合約能力不需要身分」的一行註解改寫 |
| `cmd/server/tool_catalog_contract.go` | **Modify** | 合約八件同上 → `true` |
| `cmd/server/tool_catalog_test.go` | **Modify** | `everyAbilityTheTradingServiceOffers` 那十六列改 `true`——這份字面清單就是「正是這十六件」的守門員 |
| `cmd/server/tool_catalog_market_data_sign_in_test.go` | **Add** | table-driven：十六件組出的請求帶身分、看行情的幾件不帶 |
| `README.md` | **Modify** | 若有說行情一組不需身分的地方 |
| `.sdd/UL-MAP.md` | **Modify** | `anonymous` 與「合約能力」兩列改成「看的不需要、改的需要」 |
| `internal/…`（domain / application / infrastructure） | **Not touched** | 帶身分、沒身分不去問、拒絕原話帶回、續用都是既有行為，已有測試 |
| 能力名稱、路徑、欄位 | **Not touched** | PRD 明定不變 |

## 3. New Classes / Modules

無。

## 4. Modified Components

| Component | Current role | Change needed |
| :--- | :--- | :--- |
| 十六件能力宣告 | `requiresSignIn=false` | `requiresSignIn=true` |

## 5. Component Relationships

```
能力宣告(requiresSignIn=true) → ApiToolApplication.CallApiTool
   ├─ 連線有身分 → 帶著登入憑證去問 → 交易服務回覆（含尚未開通）原話帶回
   └─ 沒身分   → 不去問，回「請先登入」
```

## 6. Extensibility & Handoff Notes

- **Most likely next requirement:** 交易服務再收緊或放寬某一件。
- **Where it lands:** 那一件宣告上的布林值一格 ＋ 字面清單一列；不必動任何流程。

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| US-01 加進觀察清單帶身分 | 宣告翻轉 + 新測試（請求帶身分）+ 既有 `TestAnAbilityThatNeedsIdentityTravelsUnderThisConnectionsProof` |
| US-01 同步合約歷史帶身分 | 同上 |
| US-01 看同步進度也帶身分 | 同上 |
| US-01 沒人登入就不去問 | 既有 `TestAnAbilityThatNeedsIdentityIsNotEvenAttemptedWithoutOne` + 宣告翻轉 |
| US-01 尚未開通原話帶回 | 既有「交易服務拒絕原話轉述」行為 + 宣告翻轉 |
| US-01 正是這十六件 | `everyAbilityTheTradingServiceOffers` 字面清單 |
| US-02 查 K 線、看合約即時更新、列合約標的不帶身分 | 新測試（請求不帶身分）+ 字面清單 |

## 8. Risks & Open Decisions

- 部署順序：本 PR 先於交易服務那一刀部署。先上線時交易服務尚未收緊，帶身分去問照樣成功，沒有空窗。
- 代價：沒登入的連線從此不能叫助理改行情——刻意的。
