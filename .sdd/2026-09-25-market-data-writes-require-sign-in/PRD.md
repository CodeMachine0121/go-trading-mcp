# Product Requirements Document (PRD)

**Status:** Draft
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** 交易服務對外開著，而改動行情的事（新增／改／刪 K 線、手動補齊、同步歷史、
  加進／移出觀察清單）今天任何人不必登入就做得到；被改壞的行情會讓每一個人的重演與機器人跟著錯。
  交易服務要把這些事收到「已登入、已放行」之後。外掛對這些事的定義仍寫著不需要身分，會空手去問。
- **Expected Outcome:** 外掛對每一件改行情的事（以及看歷史同步進度）都帶著使用者的身分去問；
  看行情的事照舊不帶身分；這一刀**先於**交易服務上線。
- **Out of Scope:** 交易服務本身的把關、前端、管理者身分或角色之分。

## 2. User Personas

- **Primary Role:** 透過交易助手操作的系統擁有者，以及替他決定怎麼呼叫的助理。
- **Usage Context:** 管理觀察清單、手動補資料或同步歷史時。

## 3. User Stories & Acceptance Criteria

### US-01 — 改行情的事帶著身分去問 [priority: P0]
**As a** 使用者，**I want** 助理替我改行情時帶著我的身分，**so that** 交易服務收緊之後這些事照樣做得成。

```gherkin
Scenario: 已登入的人把標的加進觀察清單
  Given 使用者已登入、帳號已放行
  When 助理把 BTCUSDT 加進觀察清單
  Then 這一次詢問帶著他的身分
  And 交易服務的回覆原樣帶回

Scenario: 已登入的人同步一段合約歷史
  Given 使用者已登入
  When 助理同步 BTCUSDT 三十天的合約歷史
  Then 這一次詢問帶著他的身分

Scenario: 看歷史同步的進度也帶著身分
  Given 使用者已登入
  When 助理看那一趟合約歷史同步走到哪
  Then 這一次詢問帶著他的身分

Scenario: 沒有人登入就不去問
  Given 這個連線沒有人登入
  When 助理刪一根 K 線
  Then 外掛不去問交易服務
  And 回覆「請先登入」

Scenario: 帳號還沒被放行時原話帶回
  Given 使用者已登入、帳號還沒被放行
  When 助理手動補齊一個標的
  Then 這一次詢問帶著他的身分
  And 交易服務說「尚未開通」的原話與開通指示原樣帶回

Scenario: 需要身分的正是這十六件
  Given 現貨與合約兩條線
  When 列出每一件能力需不需要身分
  Then 新增、改、刪 K 線、手動補齊、同步歷史、看同步進度、加進與移出觀察清單這八件在兩條線上都需要身分
  And 其他能力需不需要身分與之前一模一樣
```

### US-02 — 看行情照舊不必登入 [priority: P0]
**As a** 還沒登入的使用者，**I want** 照樣看得到行情，**so that** 看行情不必先登入。

```gherkin
Scenario: 沒登入也查得到一段 K 線
  Given 這個連線沒有人登入
  When 助理查 BTCUSDT 一段 K 線
  Then 這一次詢問不帶身分
  And 照常查得到

Scenario: 沒登入也看得到合約的即時更新
  Given 這個連線沒有人登入
  When 助理看一眼合約的即時更新
  Then 這一次詢問不帶身分

Scenario: 沒登入也列得出合約交易標的
  Given 這個連線沒有人登入
  When 助理列出合約交易標的
  Then 這一次詢問不帶身分
```

## 4. Business Flow & Logic

- 需要身分的事：有身分 → 帶著去問；沒身分 → 不去問、回「請先登入」；登入憑證過期 → 外掛自己換新再問（既有行為）。
- 交易服務的拒絕（包括尚未開通）一律原話帶回。
- 說明文字裡任何「行情那一整組不需要身分」的說法跟著改；能力名稱、欄位、路徑不變。

## 5. UI/UX Design & Interaction

- N/A（沒有畫面）。

## 6. Non-Functional Requirements

- 安全：這一刀的存在理由就是防止未登入的人改動共用行情。

## 7. Dependencies & Risks

- **部署順序**：外掛必須先於交易服務那一刀部署。外掛先上線時，交易服務還沒收緊，帶著身分去問照樣成功；
  反過來則有一段時間助理什麼都改不了。
- 沒登入的使用者從此不能叫助理加觀察清單——這是刻意的。

## 8. Appendix

- 交易服務那一刀：行情的寫入路線掛上登入把關（含開通把關）。
