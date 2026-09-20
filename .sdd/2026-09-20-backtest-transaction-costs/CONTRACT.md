# 回測的交易成本（連接器）— Contract Verification Matrix

**Contract source:** `.sdd/2026-09-20-backtest-transaction-costs/PRD.md`（Acceptance Criteria 為 oracle）
**Design map:** `.sdd/2026-09-20-backtest-transaction-costs/ARCH.md`
**Scope:** `go-trading-mcp`（連接器）
**Verified:** 2026-09-20
**Ceiling:** 靜態一致性稽核。逐條把**測試斷言**與**清單內容**各自對照規格推出的 oracle。

---

## Clauses

### US-01 — 助手說得出那兩個成本率

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01.1 | 兩支重演都有那兩格 | 兩支皆宣告 | `tool_catalog.go` 的 `backtestParameters()` | `TestBothReplaysTakeTheSameTwoCostRates`（兩支 × 兩格，共四個子案例） | ✅ conforms |
| AC-01.2 | 兩格皆選填 | 不填是常態 | `bodyParameter(..., false)` | 同上（`assert.False(box.IsRequired)`） | ✅ conforms |
| AC-01.3 | 兩格皆為字串 | 與清單裡每個精確小數一致 | `vo.ToolParameterKindString` | 同上（`assert.Equal(..., box.Kind)`） | ✅ conforms |
| AC-01.4 | 住共用那一組，不是各自加一次 | 不可能只改到一邊 | 兩格寫在 `backtestParameters()` 內，兩支能力各自 `append` 它 | **結構上成立**：`TestBothReplaysTakeTheSameTwoCostRates` 兩支都通過，而清單只有一份——單獨改壞一邊在原始碼上不存在這個狀態 | ✅ conforms |

### US-02 — 說明講得出助手猜不到的那幾件事

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AC-02.1 | 留白＝完全不計，不是套用常見費率 | 說明講明 | `entryCostPercentage` 說明 | `TestTheCostRatesSayWhatCannotBeDiscoveredBySending`（`"不給就是完全不計手續費"`、`"偏樂觀"`） | ✅ conforms |
| AC-02.2 | 偏差與交易次數成正比，並點名比較兩支策略的場景 | 說明講明 | 同上（含 0.47%／1.9%／六成三組數字） | 同上（`"與交易次數成正比"`、`"不填費率等於沒有在比較"`） | ✅ conforms |
| AC-02.3 | 出場留白沿用進場，且與出場距離規則**不同** | 說明講明兩件事 | `exitCostPercentage` 說明 | 同上（`"沿用 entryCostPercentage"`、`"不一樣"`） | ✅ conforms |
| AC-02.4 | 給得出台股與幣安的實際數字 | 助手看不到券商 | 同上 | 同上（`"0.0855"`、`"0.3855"`、`"幣安"`） | ✅ conforms |
| AC-02.5 | 負的與超過 100 被拒絕，正好 100 可以 | — | `entryCostPercentage` 說明 | 同上（`"正好 100 可以"`） | ✅ conforms |

### US-03 — 說明講得出成績單多出來的兩件事

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AC-03.1 | 講得出 `totalTransactionCost` | — | `costedReportCardNote`（兩支能力共用同一份） | `TestBothReplaysSayHowToReadACostedReportCard`（兩支各一個子案例） | ✅ conforms |
| AC-03.2 | 講得出「抓價差不行 vs 被手續費吃掉」 | — | 同上 | 同上（`"被手續費吃掉"`） | ✅ conforms |
| AC-03.3 | 講明 `profit` 是淨額、勝率照淨額算 | — | 同上 | 同上（`"淨額"`、`"勝率"`） | ✅ conforms |

---

## 稽核過程中補上的缺口

**兩支能力的成績單說明原本是兩份一模一樣的字面字串。** 兩份都通過測試，
但它們是**兩個可以各自被改的地方**——而這段話的內容正是「更正一個助手已經有的讀法」。
只改好一邊的那一天，一支會說 `profit` 是淨額、另一支說它是毛額，
而兩邊回來的數字**長得一模一樣**，沒有任何東西會露出馬腳。
抽成 `costedReportCardNote` 之後，那個狀態在原始碼上不存在。

---

## 沒有 orphan

新增的每一樣東西都指得回一條 AC：

| 新增 | 指回 |
| :--- | :--- |
| `entryCostPercentage`／`exitCostPercentage` 兩格 | AC-01.\*、AC-02.\* |
| `costedReportCardNote` | AC-03.\* |

沒有新型別、沒有新分支、沒有新驗證——**連接器一個數字都不碰**（PRD R-2），
所以這一刀在程式碼上就只有清單資料與文字。

---

## 已知且刻意的落差

| 落差 | 依據 |
| :--- | :--- |
| 連接器不驗證那兩格 | PRD Out of Scope ＋ R-2。負的與超過 100 由交易服務拒絕，說法原封帶回 |
| 不替助手挑費率 | PRD Out of Scope。說明給得出常見數字，選哪一個是使用者的事 |
| 每筆最低手續費不支援 | 交易服務那一側刻意不做；這裡沒有格子可加，也不該自己補 |
