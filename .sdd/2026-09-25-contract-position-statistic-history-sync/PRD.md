# Product Requirements Document (PRD)

**Status:** Draft
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** 交易服務的合約歷史同步已經會一併補齊持倉統計（從持倉統計歷史資料庫，可回溯好幾年），
  但助理讀到的說明仍說「持倉統計只留三十天、更早的查不到」。助理會因此回答使用者「補不了」，而那不是事實。
- **Expected Outcome:** 助理讀到的五份說明跟上交易服務現在的行為；不新增、不移除能力，參數不變。
- **Out of Scope:** 交易服務本身的行為、專門補持倉統計的新能力、前端。

## 2. User Personas

- **Primary Role:** 透過交易助手操作的系統擁有者，以及替他讀說明、決定怎麼呼叫的助理。
- **Usage Context:** 設計或回測合約策略時，想要比三十天更長的持倉統計。

## 3. User Stories & Acceptance Criteria

### US-01 — 助理讀得出同步也補持倉統計 [priority: P0]
**As a** 助理，**I want** 同步一段合約歷史與看進度的說明講清楚持倉統計那一份，**so that** 我能正確地建議使用者並讀懂輪次。

```gherkin
Scenario: 同步的說明說出同一趟也補持倉統計
  Given 助理要替使用者補一段合約歷史
  When 助理讀「同步一段合約歷史」的說明
  Then 讀得到「同一趟也補持倉統計、用同一個回溯天數、先補完合約 K 線再補」
  And 讀得到「只存沒有的」

Scenario: 同步的說明說出沒有那一天的檔案不算失敗
  Given 今天的持倉統計沒補到
  When 助理讀「同步一段合約歷史」的說明
  Then 讀得到「沒有那一天的檔案不算失敗」

Scenario: 看進度的說明說出持倉統計那一組與合約 K 線分開
  Given 一趟輪次帶著 positionStatistic
  When 助理讀「看一趟合約同步走到哪」的說明
  Then 讀得到 positionStatistic 與它的五項（totalDays、completedDays、storedCount、skippedCount、fetchFailureReason）
  And 讀得到「與合約 K 線那組分開、不加總」

Scenario: 看進度的說明說出歷史資料庫不答話不是失敗
  Given 輪次的持倉統計來源原因有值、狀態是成功
  When 助理讀「看一趟合約同步走到哪」的說明
  Then 讀得到「只停下持倉統計那一份、這趟仍算 succeeded」
```

### US-02 — 助理不再說「更早的查不到」 [priority: P0]
**As a** 助理，**I want** 其他提到持倉統計歷史長度的說明不再說它只有三十天，**so that** 我不會拒絕一件做得到的事。

```gherkin
Scenario: 查持倉統計的說明指向合約歷史同步
  Given 使用者問三十天以前的持倉統計
  When 助理讀「查持倉統計」的說明
  Then 讀得到「更早的可以用 trading_sync_contract_k_candle_history 補」
  And 讀不到「從來沒有人錄」

Scenario: 寫合約策略腳本的說明不再說只留三十天
  Given 使用者寫一支用持倉量的合約腳本
  When 助理讀寫合約策略腳本的說明
  Then 讀不到「持倉統計只留三十天」
  And 讀得到「只有錄到或同步過的那段才有值」

Scenario: 手動補齊的說明指向合約歷史同步
  Given 使用者要手動補齊合約 K 線
  When 助理讀「手動補齊合約 K 線」的說明
  Then 讀得到它只補 K 線
  And 讀得到更早的持倉統計要用 trading_sync_contract_k_candle_history

Scenario: 能力清單不變
  Given 這一刀之前的能力清單
  When 列出助理的能力
  Then 能力名稱、路徑與參數與之前一模一樣
```

## 4. Business Flow & Logic

- 只改說明文字；路由、參數、需不需要身分都不變。

## 5. UI/UX Design & Interaction

- N/A（沒有畫面）。

## 6. Non-Functional Requirements

- N/A。

## 7. Dependencies & Risks

- 依賴交易服務的合約歷史同步已部署；說明先於部署上線時，助理會描述一件還沒生效的行為——與交易服務同步上線。

## 8. Appendix

- 交易服務那一刀：合約歷史同步一併補齊持倉統計。
