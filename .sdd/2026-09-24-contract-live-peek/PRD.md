# Product Requirements Document (PRD) — 看一眼合約即時更新

**Status:** Draft · **Version:** v1.0 · **Owner:** James Hsueh

---

## 1. Background & Goal

- **Problem Statement:** 交易服務已替永續合約開了自己的即時跟盤，但助理只看得到現貨那一根的即時樣子；
  問合約現在的價格，它只能用每分鐘那一輪存下的合約 K 線回答，最多晚一分鐘。
- **Expected Outcome:** 助理多一件**看一眼合約即時更新**的能力，與現貨那一件分開；它讀得懂每一種狀態、
  知道內容是最新價而非標記價格、知道只看得到合約追蹤名單上的合約標的。現貨的看一眼一字不變。
- **Out of Scope:** 一直開著的即時通道；標記價格、指數價格的即時更新；現貨看一眼的任何改變。

## 2. User Personas

- **Primary Role:** 透過外部助理操作交易服務的使用者（助理代他查行情）。
- **Usage Context:** 對話中問「某個永續合約現在怎樣」，助理看一眼再回答。

## 3. User Stories & Acceptance Criteria

### US-01 — 看一眼合約標的現在的樣子 [P0]

**As a** 透過助理查行情的使用者, **I want** 助理看得到合約標的進行中那一根的即時樣子,
**so that** 我問合約現在的價格時，得到的是現在，而不是一分鐘前、更不是現貨的價格。

```gherkin
Scenario: 看一眼合約標的
  Given BTCUSDT 在合約追蹤名單上，這一分鐘有成交
  When 助理看一眼合約 BTCUSDT
  Then 回的是合約那一邊進行中那一根的樣子

Scenario: 現貨的看一眼照舊
  Given 同一個代號 BTCUSDT
  When 助理看一眼現貨 BTCUSDT
  Then 用的是原本現貨那一件能力，說明與今天一字不差

Scenario: 等滿沒有更新
  Given 合約 BTCUSDT 在上限時間內沒有任何更新
  When 助理看一眼合約 BTCUSDT
  Then 回「在這段時間之內沒有收到任何即時更新，這不是錯誤」
```

### US-02 — 只看得到合約追蹤名單上的 [P0]

**As a** 使用者, **I want** 助理在合約標的不在追蹤名單上時知道該怎麼辦, **so that** 它不會一再重試一件不會自己好的事。

```gherkin
Scenario: 不在合約追蹤名單上
  Given DOGEUSDT 系統認得，但不在合約追蹤名單上
  When 助理看一眼合約 DOGEUSDT
  Then 照交易服務的說法回：要先把它加進合約追蹤名單
  And 這件能力的說明指向「加進合約追蹤名單」那一件能力

Scenario: 系統不認得的代號
  Given NOPEUSDT 系統不認得
  When 助理看一眼合約 NOPEUSDT
  Then 照交易服務的說法回：找不到這個合約標的
```

### US-03 — 讀得懂每一種狀態 [P0]

**As a** 使用者, **I want** 助理讀懂每一則更新的狀態, **so that** 它不會把成形中的數字當成定案，也不會把斷線當成壞掉。

```gherkin
Scenario: 成形中
  Given 第一則更新是成形中
  When 助理讀這件能力的說明
  Then 說明寫出：這一根還在走、數字還會變、系統不存它

Scenario: 走完了
  Given 第一則更新是走完了
  When 助理讀這件能力的說明
  Then 說明寫出：這一根不是由即時更新存下，每分鐘那一輪會在一分鐘內存入完整的那一根

Scenario: 即時更新斷了
  Given 第一則更新是即時更新斷了
  When 助理讀這件能力的說明
  Then 說明寫出：系統自己在重連，等一下就好

Scenario: 內容是最新價
  When 助理讀這件能力的說明
  Then 說明寫出：內容是最新價的開高低收與成交量，不含標記價格
  And 說明寫出：合約不會出現「分不到名額」或「休市中」
```

## 4. Business Flow & Logic

- 看法與現貨相同：收到第一則即時更新即回；等滿上限（與現貨同一個上限）仍無更新是正常結果。
- 不需要登入。
- 交易服務的拒絕（空白、找不到、不在名單上、暫停服務）原樣轉給助理。
- 名字與說明都帶「合約」，與現貨那一件分得開。

## 5. UI/UX Design & Interaction

N/A（助理能力的說明文字即介面）。

## 6. Non-Functional Requirements

- 絕不維持一條一直開著的通道：看一眼後即收掉連線（與現貨同一條規則）。

## 7. Dependencies & Risks

- 依賴交易服務的合約即時跟盤入口（CodeMachine0121/go-trading#70）；它上線前呼叫會被交易服務以「找不到」回覆。

## 8. Appendix

- `.sdd/2026-09-24-contract-live-peek/BRIEF.md`
