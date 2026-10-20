# Product Requirements Document (PRD) — 一台機器人盯多個交易標的（外掛）

**Status:** Finalized
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering（交易服務、外掛同一個版本一起更新）

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** 交易服務讓一台策略機器人同時盯 1–10 個**監聽標的**，並且不再收單一標的的寫法。外掛的建立／修改機器人能力仍只收一個標的、讀機器人與執行紀錄的說明仍把機器人當成只盯一個標的——助理會送出被拒的單一標的、看不懂每個標的的狀態、把標的停擺誤讀成整台停擺。
- **Expected Outcome:**
  - 助理能替使用者用一串標的建立機器人、用完整清單修改監聽標的，而且知道漏列就是移除。
  - 助理讀得懂每個監聽標的的狀態，分得清整台停擺與標的停擺。
  - 助理讀執行紀錄時說得出每一筆是哪個標的，也能只讀某一個標的。
  - 助理在建立或加標的時說清楚：每個標的各自照整份部位規劃算（總曝險最多 N 倍）、合約機器人開著自動下單時每個標的都會真的下單。
  - 外掛各處不再把機器人說成只盯一個標的。
- **Out of Scope:**
  - 同時持倉數量上限、個別標的覆寫部位規劃或槓桿。
  - 外掛自己檢查標的數量、重複、追蹤名單或格式——交易服務才是規則的來源，外掛原樣轉送、拒絕照原話帶回。
  - 打開或關掉自動下單、下單、撤單（仍只能由使用者在網頁上做）。
  - 舊的單一標的寫法相容。

---

## 2. User Personas

- **Primary Role(s):** 助理（透過外掛代使用者操作）；機器人擁有者（對助理下指令的人）。
- **Usage Context:** 使用者在對話裡請助理建立、調整、查看機器人與它的執行紀錄。

---

## 3. User Stories & Acceptance Criteria

### US-01 — [priority: P0] 用一串標的建立機器人
**As a** 助理, **I want** 建立機器人時給一串交易標的, **so that** 使用者一台機器人就能盯好幾個標的。

```gherkin
Scenario: 建立盯兩個標的的機器人
  Given 使用者要用一份交易策略盯 BTCUSDT 與 ETHUSDT
  When 助理建立機器人
  Then 送給交易服務的監聽標的是 BTCUSDT、ETHUSDT 兩個
  And 交易服務回覆的機器人列出兩個監聽標的，都沒有上一次送出的信號

Scenario: 只盯一個標的
  Given 使用者只要盯 BTCUSDT
  When 助理建立機器人
  Then 送給交易服務的監聽標的只有 BTCUSDT 一個

Scenario: 正好 10 個標的
  Given 使用者給了 10 個不同的標的
  When 助理建立機器人
  Then 10 個標的原樣送給交易服務，建立成功

Scenario: 超過 10 個標的
  Given 使用者給了 11 個標的
  When 助理建立機器人
  Then 交易服務拒絕，說明「一台機器人最多盯 10 個交易標的」，助理照原話轉述並請使用者刪減

Scenario: 重複的標的
  Given 使用者給了 BTCUSDT 與 btcusdt
  When 助理建立機器人
  Then 交易服務拒絕，說明「交易標的 BTCUSDT 重複了」

Scenario: 合約機器人有標的不在追蹤名單
  Given DOGEUSDT 不在合約追蹤名單上
  When 助理用 BTCUSDT 與 DOGEUSDT 建立合約機器人
  Then 交易服務拒絕並點名 DOGEUSDT，助理建議先把它加進合約追蹤名單

Scenario: 總曝險提醒
  Given 部位規劃資金 1000、固定金額 1000，使用者要盯 3 個標的
  When 助理準備建立機器人
  Then 建立機器人能力的說明告訴助理：每個標的各自照整份部位規劃算，總曝險最多是部位規劃的 3 倍
```

### US-02 — [priority: P0] 以完整清單修改監聽標的
**As a** 助理, **I want** 修改機器人時給修改後的完整標的清單, **so that** 加減標的不必刪掉重建，也不會誤刪標的。

```gherkin
Scenario: 加一個標的
  Given 已停止的機器人盯 BTCUSDT（上一次送出買入）與 ETHUSDT
  When 使用者說「再加 SOL」，助理修改機器人
  Then 送出的監聽標的是 BTCUSDT、ETHUSDT、SOLUSDT
  And BTCUSDT 仍記著買入，SOLUSDT 是空白

Scenario: 只改別的設定時照抄標的
  Given 已停止的機器人盯 BTCUSDT 與 ETHUSDT
  When 使用者說「把間隔改成 15 分鐘」
  Then 修改機器人能力的說明告訴助理要把 BTCUSDT、ETHUSDT 照抄送出，漏列就是移除

Scenario: 移除一個空手的標的
  Given ETHUSDT 空手、沒有沒做完的自動下單
  When 使用者說「不要盯 ETH 了」，助理送出只有 BTCUSDT 的清單
  Then 修改成功，ETHUSDT 過去的執行紀錄仍看得到

Scenario: 移除還有機器人持倉的標的
  Given ETHUSDT 的機器人持倉做多 0.5
  When 助理送出不含 ETHUSDT 的清單
  Then 整次修改被拒，理由是那個標的還有機器人持倉
  And 助理說明要先讓 ETHUSDT 空手才能移除，不重送同一份

Scenario: 移除還有沒做完自動下單的標的
  Given ETHUSDT 有一筆沒做完的自動下單
  When 助理送出不含 ETHUSDT 的清單
  Then 整次修改被拒，理由是那個標的還有沒做完的自動下單
  And 助理說明要等它做完才能移除

Scenario: 舊的單一標的寫法不再送出
  Given 助理只填了單一標的那一格
  When 助理建立或修改機器人
  Then 那一格不會送給交易服務；能力只認一串監聽標的
```

### US-03 — [priority: P0] 讀懂每個監聽標的的狀態
**As a** 助理, **I want** 讀機器人時看懂每個監聽標的的狀態, **so that** 我能正確回答哪個標的有倉、哪個停擺。

```gherkin
Scenario: 一個標的停擺、另一個有倉
  Given BTCUSDT 上一次送出買入、機器人持倉做多 0.01；ETHUSDT 標的停擺（策略腳本算不出來）
  When 使用者問「這台怎麼樣」
  Then 查看機器人能力的說明讓助理說出：BTCUSDT 有多單 0.01、ETHUSDT 停擺、其他標的照常、機器人仍在跑

Scenario: 全部標的都停擺
  Given 每個監聽標的都停擺
  When 使用者問「這台怎麼樣」
  Then 助理讀到機器人上的停擺原因，說整台已停擺

Scenario: 讓停擺的標的恢復
  Given ETHUSDT 標的停擺
  When 使用者說「修好了，讓它恢復」
  Then 說明告訴助理停止再啟動這台機器人即恢復，機器人持倉不動
```

### US-04 — [priority: P0] 執行紀錄看得出標的、可以只看一個標的
**As a** 助理, **I want** 讀執行紀錄時知道每一筆是哪個標的、也能只讀一個標的, **so that** 使用者問某個標的時我不必翻全部紀錄。

```gherkin
Scenario: 只看 ETHUSDT
  Given 機器人盯 BTCUSDT 與 ETHUSDT
  When 使用者說「ETH 最近幾輪說了什麼」，助理只讀 ETHUSDT
  Then 交易服務只回 ETHUSDT 的執行紀錄，由新到舊

Scenario: 讀全部標的
  Given 機器人盯 BTCUSDT 與 ETHUSDT
  When 使用者說「最近幾輪」，助理不指定標的
  Then 不帶任何標的篩選，回覆每一筆都寫著它的標的

Scenario: 篩選的標的格式不對
  Given 助理給了格式不對的標的
  When 助理讀執行紀錄
  Then 交易服務拒絕，助理照原話改一改再試
```

### US-05 — [priority: P1] 其他說明不再說機器人只盯一個標的
**As a** 助理, **I want** 外掛其他說明也用監聽標的的說法, **so that** 我不會用錯誤的前提回答使用者。

```gherkin
Scenario: 把合約標的移出追蹤名單前查哪些機器人盯著它
  Given 使用者要把 ETHUSDT 移出合約追蹤名單
  When 助理讀移出追蹤名單能力的說明
  Then 說明要它看合約機器人的監聽標的裡有沒有 ETHUSDT，且只有那一個標的會失明、其他標的照常

Scenario: 自動下單的說明涵蓋每個標的
  Given 合約機器人開著自動下單、盯 3 個標的
  When 助理讀任何機器人能力的說明
  Then 說明說清楚每一個監聽標的都會真的在使用者的幣安帳戶上開倉平倉
```

---

## 4. Business Flow & Logic

- **Core Business Rules:**
  - 建立與修改機器人只收**一串監聽標的**；外掛原樣轉送，不檢查數量、重複、追蹤名單。
  - 修改時給的是**完整清單**：留下的保留狀態、新加的空白、沒列的移除；移除有機器人持倉或沒做完自動下單的標的整次被拒，助理照原話說明，不重送同一份。
  - 讀機器人：每個監聽標的有標的、上一次送出的信號、衝突記號、標的停擺原因，合約機器人另有機器人持倉；機器人上的停擺原因只代表整台停擺。
  - 停止再啟動清掉每個標的的停擺、上一次送出的信號與衝突記號；機器人持倉不動。對已停止的機器人立刻跑一輪時每個標的都跑（含停擺中的），跑出結論的恢復。
  - 執行紀錄每一筆有標的；讀取時可選一個標的篩選，不給就是全部；輪次編號在同一台機器人的各標的之間不重複；每個標的各記得最近 50 輪。
  - 部位規劃每個標的各自照整份算；合約機器人開著自動下單時每個標的都會真的下單。
  - 合約標的行情停了只跳過那一個標的那一輪（記為持有）。
- **Edge Cases:**
  - 只想改名或改間隔時漏列標的 → 等於移除，說明必須明白警告。
  - 移除被拒時不重送同一份，請使用者處理該標的的持倉或等自動下單做完。

---

## 5. UI/UX Design & Interaction

- N/A（外掛沒有畫面；對助理而言，能力的說明就是介面）。

---

## 6. Non-Functional Requirements

- **Compatibility:** 與交易服務同一個版本上線，不保留單一標的寫法；外掛要在交易服務之後合併。
- **Security:** 不變——每一件事仍以使用者授權給外掛的帳號進行，自動下單照舊只能在網頁上切換。

---

## 7. Dependencies & Risks

- **External Dependencies:** 交易服務的「一台機器人盯多個交易標的」（go-trading PR #112）必須先上線。
- **Known Risks:**
  - 助理修改其他設定時漏抄標的而誤刪標的——以說明明白警告；交易服務也會擋下有持倉或自動下單的標的。
  - 總曝險放大——以說明讓助理在建立或加標的時提醒使用者。

---

## 8. Appendix

- 交易服務切片：`go-trading/.sdd/2026-10-20-multi-symbol-strategy-bot/PRD.md`。
- 相關外掛切片：`2026-09-24-contract-strategy-bots`、`2026-10-12-auto-order-limit-entry`。
