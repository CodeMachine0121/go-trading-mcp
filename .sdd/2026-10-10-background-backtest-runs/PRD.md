# Product Requirements Document (PRD) — 回測改為送出後查看（外掛）

**Status:** Finalized
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering

---

## 1. Background & Goal (Why & Goal)

- **Problem Statement:** 交易服務把回測改成先記下一筆回測輪次、立刻回覆，成績單留在輪次上。外掛今天的四個回測能力是「送出、等最多兩分鐘、拿到濃縮成績單」，交易服務不再當場回成績單後它們會拿到一筆輪次卻把它當成績單讀。另外外掛只能開一份，因為連線狀態活在記憶體裡。
- **Expected Outcome:**
  - 四個回測能力輸入不變、回覆改成回測輪次，說明告訴助理去看輪次。
  - 新增看一筆（可順便等一下）、列出我的、取消三個能力。
  - 助理從說明就知道排隊中不必重送、失敗原因照說。
  - 外掛可以開多份，任何一次呼叫落在哪一份都一樣。
- **Out of Scope:** 交易服務本身；操作台；行情助手的重演。

---

## 2. User Personas

- **Primary Role(s):** 在 Claude Code 裡透過助理操作交易服務的使用者；助理本身（讀說明）。
- **Usage Context:** 使用者請助理跑長區間回測，等結果時繼續聊別的，稍後請助理查。

---

## 3. User Stories & Acceptance Criteria

### US-01 — [priority: P0] 回測能力送出就回輪次
**As a** 使用者, **I want** 助理送出回測後立刻告訴我輪次編號, **so that** 長回測不會因為等太久而斷掉。

```gherkin
Scenario: 合約腳本回測送出就回
  Given 合約腳本 #88 是我的
  When 助理送出 BTCUSDT、1 小時刻度、四年的合約腳本回測
  Then 立刻拿到輪次 #40（排隊中或進行中）
  And 助理回覆我「已送出，輪次 #40，可以隨時回來看」

Scenario: 四個回測能力的輸入一字不改
  When 助理以今天相同的輸入送出任一種回測
  Then 交易服務接受，回覆那筆輪次

Scenario: 不合法的要求拒絕原話帶回
  When 助理送出結束早於開始的回測
  Then 交易服務的拒絕原話帶回
  And 沒有輪次

Scenario: 回測能力不再比別的能力等更久
  When 助理送出任一種回測
  Then 外掛用一般等待時間等交易服務回覆
```

### US-02 — [priority: P0] 看一筆輪次，可順便等
**As a** 使用者, **I want** 助理拿輪次編號就看得到進度與成績單, **so that** 我不必自己開網頁。

```gherkin
Scenario: 完成的輪次帶精簡成績單
  Given #40 已完成，有 3,000 筆交易、權益曲線 35,000 點
  When 助理看 #40
  Then 回覆完成
  And 成績單帶最近 100 筆交易並說總筆數 3,000、曲線 200 點並說總點數 35,000
  And 成績單的數字一個不動

Scenario: 順便等一下、提前結束就提前回
  Given #40 進行中 40%
  When 助理看 #40 並說最多等 30 秒，第 12 秒時完成
  Then 第 12 秒回覆完成與成績單

Scenario: 等滿仍未結束回當時狀態
  Given #40 進行中
  When 助理看 #40 並說最多等 30 秒，30 秒後仍在跑
  Then 回覆進行中與當時進度，沒有成績單

Scenario: 最多等六十秒
  When 助理看 #40 並說最多等 600 秒
  Then 外掛只等 60 秒

Scenario: 失敗原因照說
  Given #40 失敗：超過時間上限
  When 助理看 #40
  Then 回覆失敗與「超過時間上限」
  And 助理照原因說，不猜成腳本壞了

Scenario: 別人的輪次「找不到」
  Given #41 是別人的
  When 助理看 #41
  Then 回覆「找不到」
```

### US-03 — [priority: P0] 列出我的、取消
**As a** 使用者, **I want** 助理找得到我以前跑過的回測、取消還沒跑完的, **so that** 不必憑記憶重跑。

```gherkin
Scenario: 列出新到舊並說總數
  Given 我有 25 筆輪次
  When 助理列出，最多 20 筆
  Then 回覆新到舊 20 筆並說總數 25

Scenario: 只看失敗的
  When 助理列出狀態為失敗的
  Then 只回覆失敗的那幾筆

Scenario: 取消排隊中的
  Given #42 排隊中
  When 助理取消 #42
  Then #42 已取消

Scenario: 已結束的拒絕原話帶回
  Given #43 已完成
  When 助理取消 #43
  Then 交易服務的拒絕原話帶回

Scenario: 使用者問昨天跑過的
  When 使用者說「幫我看昨天那幾次回測」
  Then 助理用列出能力找，不重跑
```

### US-04 — [priority: P0] 說明讓助理做對事
**As a** 使用者, **I want** 助理看到排隊中不叫我重送, **so that** 不會浪費每人上限。

```gherkin
Scenario: 排隊中的說法
  Given #44 排隊中
  When 助理看 #44
  Then 助理回覆「前面還有在跑的，會自動開始」，不叫我重送

Scenario: 每人上限滿了
  When 助理送出回測被拒「你同時最多 3 筆」
  Then 助理原話轉達並建議等一筆結束或取消一筆
```

### US-05 — [priority: P0] 外掛可以多份一起接
**As a** 系統擁有者, **I want** 外掛開三份時使用者感覺不出來, **so that** 部署可以水平擴充。

```gherkin
Scenario: 連續兩次呼叫落在不同分身
  Given 外掛開三份
  When 同一位使用者連續兩次呼叫落在不同分身
  Then 兩次都正常回覆，不要求重新連線

Scenario: 重新連線的指引不變
  Given 交易服務不認得這份外掛授權
  When 任一次呼叫
  Then 回覆請到 Claude Code 的外掛選單重新連線
```

---

## 4. Business Flow & Logic

- 送出 → 交易服務回輪次 → 助理告訴使用者編號 → 使用者想看時助理看一筆（可順便等）→ 完成就精簡成績單、失敗就原因。
- 精簡規則與今天相同：權益曲線兩百點、交易最近一百筆、分段各自照做、成績單數字不動。
- 外掛不保留任何跨呼叫的狀態。

---

## 5. UI/UX Design & Interaction

- 說明文字是外掛的介面：四個回測能力說「它不等算完，請拿編號看」；看一筆說「可順便等最多 60 秒」；列出說「回覆時說出總數」；取消說「只能取消排隊中與進行中」。

---

## 6. Non-Functional Requirements

- **Performance:** 回測能力只等一般等待時間；看一筆的等待上限 60 秒加一般等待時間。
- **Security:** 每一次呼叫自帶外掛授權；外掛不存任何連線狀態。
- **Compatibility:** 工具清單與 README 同步；既有能力名字不變。

---

## 7. Dependencies & Risks

- 交易服務切片 `go-trading/.sdd/2026-10-10-background-backtest-runs` 先完成。
- 風險：助理若忽略說明仍可能重送；每人上限會擋住，不會造成傷害。

---

## 8. Appendix

- `BRIEF.md`（本切片）；交易服務的 PRD。
