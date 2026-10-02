# Product Requirements Document (PRD) — 歷史同步的推定休市天數

**Status:** Finalized
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** 交易服務的現貨歷史同步會跳過來源說沒資料的平日，並在輪次上記下推定休市天數；外掛的說明沒提，助理會把「完成但存得少」誤判為故障，或把它與略過根數混為一談。
- **Expected Outcome:** 兩個現貨歷史同步能力的說明講清楚推定休市天數的意思、它不是來源不答話、它與略過根數不同，以及國定假日不會讓同步停下。
- **Out of Scope:** 交易服務本身；合約歷史同步的說明。

---

## 2. User Personas

- **Primary Role(s):** 代使用者操作交易服務的助理。
- **Usage Context:** 發起歷史同步前、以及回頭查一趟同步走到哪時讀能力說明。

---

## 3. User Stories & Acceptance Criteria

### US-01 — [priority: P0] 讀進度時懂推定休市天數
**As a** 助理, **I want** 能力說明告訴我推定休市天數是什麼, **so that** 我不會把被跳過的假日回報成故障。

```gherkin
Scenario: 說明列出推定休市天數並說出它的意思
  When 助理讀「看一趟歷史同步走到哪」的說明
  Then 說明列出推定休市天數
  And 說明說它是被跳過的平日數：來源說那一天沒有這個標的的資料
  And 說明說它不算來源不答話

Scenario: 說明分開推定休市天數與略過根數
  When 助理讀「看一趟歷史同步走到哪」的說明
  Then 說明說推定休市天數與略過根數是兩件事
  And 說明說略過根數是來源答了、但某幾根不合格
```

### US-02 — [priority: P0] 發起同步前就知道假日不會讓它停下
**As a** 助理, **I want** 發起同步的說明告訴我國定假日會被跳過, **so that** 我不會叫使用者把長區間拆成好幾段。

```gherkin
Scenario: 發起同步的說明講假日
  When 助理讀「同步一段指定長度的歷史」的說明
  Then 說明說落在平日的國定假日會被跳過、記在推定休市天數上
  And 說明說這不會讓同步停下

Scenario: 發起同步的說明講連續太久的例外
  When 助理讀「同步一段指定長度的歷史」的說明
  Then 說明說連續 15 個交易日都沒資料時同步會停下
  And 說明說那多半是來源不認得這個代號、原因寫在來源拒絕原因

Scenario: 合約那邊不提推定休市
  When 助理讀合約歷史同步的兩個說明
  Then 說明裡沒有推定休市天數
```

---

## 4. Business Flow & Logic

- 純說明文字；外掛把交易服務的回覆原樣轉交，推定休市天數本來就會出現在回覆裡。

## 5. UI/UX Design & Interaction

- N/A。

## 6. Non-Functional Requirements

- N/A。

## 7. Dependencies & Risks

- 依賴交易服務回覆中的推定休市天數欄位（`go-trading` 同日的切片）。外掛先上線也無害：說明提到的數字在舊服務上不存在，助理只是讀不到。

## 8. Appendix

- `go-trading/.sdd/2026-10-02-history-sync-skips-closed-days`。
