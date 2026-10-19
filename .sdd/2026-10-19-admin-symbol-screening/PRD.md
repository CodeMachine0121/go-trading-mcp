# Product Requirements Document (PRD) — 管理員標的篩選（外掛）

**Status:** Finalized
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** 交易服務新增了只給管理員用的標的篩選，但助理在 Claude Code 裡碰不到它：管理員得自己開網頁或打交易服務，才能管理篩選設定、送出一串候選標的、回頭看哪些通過、哪些被淘汰。
- **Expected Outcome:**
  - 助理能代管理員做篩選的每一件事：列出／查看／建立／修改／刪除篩選設定，送出篩選，列出／查看（可順便等一下）／取消篩選——共九件標的篩選能力。
  - 每一件的說明都寫明只有管理員能用；非管理員被拒時助理照原話轉達「這是管理員功能」、不重試。
  - 助理讀得懂結果並用人話講：每個候選標的通過、淘汰在哪一關、無法完成的原因、組合貢獻、推薦順序。
  - 「我是誰」會說出這個帳號是不是管理員。
- **Out of Scope:**
  - 開通或取消管理員。
  - 外掛自己檢查任何篩選規則、整理標的清單或補預設值——全部交給交易服務。
  - 把通過的標的自動加入追蹤或建立機器人。

---

## 2. User Personas

- **Primary Role(s):** 管理員（在 Claude Code 裡請助理代辦篩選）；一般使用者（只會碰到拒絕）。
- **Usage Context:** 研究要不要加標的時，對助理說「幫我篩這幾個幣」，之後再問「結果出來了嗎」。

---

## 3. User Stories & Acceptance Criteria

### US-01 — [priority: P0] 只有管理員能用
**As a** 管理員, **I want** 助理能代我做篩選、而一般使用者用不到, **so that** 這個研究工具不會被其他人誤用。

```gherkin
Scenario: 管理員透過助理送出篩選
  Given 外掛授權的帳號是管理員
  And 篩選設定 4 是這位管理員的
  When 助理用篩選設定 4 送出「XRPUSDT, LINKUSDT」
  Then 帶回一筆排隊中的篩選與它的編號
  And 標的篩選能力的說明要求助理把篩選編號說給使用者

Scenario: 非管理員被擋下，助理照原話轉達
  Given 外掛授權的帳號不是管理員
  When 助理列出篩選設定
  Then 帶回交易服務的拒絕「這是管理員功能」
  And 標的篩選能力的說明告訴助理照原話轉達、不要重試

Scenario: 我是誰說出是不是管理員
  When 助理問我是誰
  Then 我是誰的說明寫明回覆會說出這個帳號是不是管理員
```

### US-02 — [priority: P0] 送出篩選時清單照原樣交給交易服務
**As a** 管理員, **I want** 貼什麼就送什麼, **so that** 整理與拒絕都由交易服務一處決定。

```gherkin
Scenario: 大小寫與全形逗號原樣送出
  Given 使用者貼「xrpusdt，LINKUSDT」
  When 助理送出篩選
  Then 交易服務收到的標的清單一字不改是「xrpusdt，LINKUSDT」

Scenario: 正好二十個標的照常送出
  Given 使用者貼二十個不重複的標的
  When 助理送出篩選
  Then 二十個標的一字不改送給交易服務

Scenario: 不允許的字元由交易服務拒絕並原話帶回
  Given 使用者貼「BTC-USDT」
  When 助理送出篩選
  Then 帶回交易服務的拒絕原話
```

### US-03 — [priority: P0] 看一筆篩選可以順便等一下
**As a** 管理員, **I want** 助理問一次就能等到結果, **so that** 不必每隔幾秒問一次。

```gherkin
Scenario: 說要等就把最多等幾秒帶給交易服務
  When 助理看篩選 9 並說最多等 30 秒
  Then 交易服務被要求看篩選 9、最多等 30 秒

Scenario: 沒說要等就立刻問
  When 助理看篩選 9、沒說要等
  Then 交易服務被要求看篩選 9，而且沒有被要求等
```

### US-04 — [priority: P0] 建立設定只送使用者給的項目
**As a** 管理員, **I want** 沒說的項目交給交易服務用預設值, **so that** 預設值只有一個地方決定。

```gherkin
Scenario: 只給必要項目
  Given 使用者只說名稱「趨勢 V3」、策略腳本 108、看盤間隔每 4 小時
  When 助理建立篩選設定
  Then 交易服務只收到名稱、策略腳本與看盤間隔三項

Scenario: 明確給出等於預設的歷史天數
  Given 使用者另外說歷史天數 730
  When 助理建立篩選設定
  Then 交易服務也收到歷史天數 730

Scenario: 看盤間隔太細由交易服務拒絕
  Given 使用者說看盤間隔每 15 分鐘
  When 助理建立篩選設定
  Then 帶回交易服務的拒絕原話
```

### US-05 — [priority: P1] 結果講成人話
**As a** 管理員, **I want** 助理用白話說明每個標的的結論, **so that** 我不用自己讀欄位。

```gherkin
Scenario: 看一筆篩選的說明教助理怎麼讀結論
  When 助理讀看一筆篩選的說明
  Then 說明解釋通過、淘汰、無法完成、已取消四種結論
  And 說明解釋淘汰會卡在基本條件或策略移植哪一關
  And 說明解釋組合貢獻是現有組合最長空窗縮短的天數、推薦順序第一名最值得考慮
  And 說明提醒通過不是買賣建議
```

---

## 4. Business Flow & Logic

- **Core Business Rules:**
  - 九件標的篩選能力只轉達：管理篩選設定（列出、查看、建立、修改、刪除）、送出篩選、列出篩選、查看篩選（最多等六十秒）、取消篩選。
  - 外掛不檢查管理員身分、不整理清單、不補預設值；交易服務的拒絕一律原話帶回（未登入照舊請使用者重新連線）。
  - 每件能力說明寫明只有管理員能用、被拒時照原話轉達「這是管理員功能」、不重試、不教人怎麼變成管理員。
  - 送出後說明要求助理說出篩選編號；排隊中代表前面還有在跑的、會自動開始，不要重送；同一時間只能有一筆自己的篩選。
- **Edge Cases:** 交易服務連不到時照外掛既有說法回「連不到交易服務」，晚一點再試同一件事。

---

## 5. UI/UX Design & Interaction

N/A（外掛沒有畫面；能力說明就是介面）。

---

## 6. Non-Functional Requirements

- **Performance:** 查看一筆篩選最多等六十秒，與看一筆回測輪次相同。
- **Security:** 外掛不持有任何身分判斷；全部以授權的那位使用者身分轉達，由交易服務決定能不能。
- **Compatibility:** 既有能力的名稱與輸入不變；能力清單多九件。

---

## 7. Dependencies & Risks

- **External Dependencies:** 交易服務的標的篩選（`go-trading` 同日功能）。
- **Known Risks:** 交易服務的篩選若改了欄位，說明需要同步；以能力清單完整性的檢查守住「交易服務有、外掛也有」。

---

## 8. Appendix

- 交易服務規格：`go-trading/.sdd/2026-10-19-admin-symbol-screening/`
- 相似先例：回測改為送出後查看（`2026-10-10-background-backtest-runs`）
