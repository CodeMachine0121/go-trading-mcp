# Product Requirements Document (PRD)

**Status:** Draft
**Version:** v1.0
**Owner:** James Hsueh
**Stakeholders:** Engineering

---

## 1. Background & Goal (Why & Goal)

### Problem Statement

交易服務的永續合約是一條獨立於現貨的線：自己的 K 線、自己的追蹤名單、自己的同步輪次，
而且剛多了資金費率結算、持倉統計、交易規格、完整維持保證金分級四種資料。
外掛的能力清單裡一件合約的事都沒有，助理只能去問現貨、拿回不相干的數字，或說做不到。

### Expected Outcome

- 合約那一條線的每一件事都是一個能力，共 **14 件**，能力從 51 件變成 **65 件**。
- 合約的能力名字一律帶「合約」，只問得到合約那一條線；現貨的能力一件都不變。
- 每一件合約能力的說明，寫出助理在送出之前不可能自己知道的事。

### Out of Scope

- 合約的重演（交易服務還沒有）。
- 交易服務內建的行情對話助手。
- 任何現貨能力的行為改變。

---

## 2. User Personas

**Primary Role:** 助理（代使用者操作交易服務的 AI），以及它背後的使用者。
助理只讀得到能力清單與每一件能力的說明，靠它們決定要做哪一件、填什麼、怎麼解讀回來的東西。

---

## 3. User Stories & Acceptance Criteria

### US-01 — 合約的事認得出、問得對 [priority: P0]

**As a** 助理，**I want** 合約與現貨的能力一眼分得出來、各自只問到自己那一條線，**so that** 我不會拿現貨的數字回答合約的問題。

```gherkin
Scenario: 查合約 K 線問到合約那一條線
  Given 使用者要看 BTCUSDT 永續合約這一小時的 K 線
  When 助理用查合約 K 線的能力
  Then 問到合約那一條線,回來的是帶標記價格、指數價格、溢價指數的合約 K 線

Scenario: 查現貨 K 線仍問到現貨那一條線
  Given 使用者要看 BTCUSDT 現貨這一小時的 K 線
  When 助理用查 K 線的能力
  Then 問到現貨那一條線,與合約無關

Scenario: 名字分得出合約與現貨
  Given 能力清單上的每一件能力
  When 看它的名字
  Then 合約的每一件都帶「合約」,現貨的沒有一件帶

Scenario: 合約的能力不會問到現貨那一條線
  Given 任何一件合約的能力
  When 它被送出
  Then 它只問到合約那一條線
```

### US-02 — 寫一根合約 K 線要每一項 [priority: P0]

**As a** 助理，**I want** 寫合約 K 線時每一格都標成必填，**so that** 我不會漏填而被整份拒絕。

```gherkin
Scenario: 每一項都填了就整根轉達
  Given 新增合約 K 線時每一項都填了
  And 溢價指數收盤是 -0.0005
  When 送出
  Then 交易服務收到的那一根帶著溢價指數收盤 -0.0005 與其他每一項

Scenario: 沒填指數價格還沒送出就被擋下
  Given 新增合約 K 線時沒填指數價格
  When 送出
  Then 在送到交易服務之前就被擋下,並說出缺的是指數價格那一格

Scenario: 成交筆數填零照常轉達
  Given 新增合約 K 線時成交筆數填 0,其他每一項都填了
  When 送出
  Then 交易服務收到成交筆數 0

Scenario: 修改時沒填溢價指數被擋下
  Given 修改合約 K 線時沒填溢價指數
  When 送出
  Then 在送到交易服務之前就被擋下,並說出缺的是溢價指數那一格
```

### US-03 — 合約特有的資料查得到 [priority: P0]

**As a** 助理，**I want** 查得到資金費率結算、持倉統計、完整維持保證金分級，**so that** 我回答得出合約持倉的成本與市場擁擠程度。

```gherkin
Scenario: 查資金費率結算
  Given BTCUSDT 兩天內有六次結算
  When 助理指定 BTCUSDT 與這兩天查資金費率結算
  Then 問到合約那一條線的資金費率結算,帶著合約標的與起訖時間

Scenario: 查持倉統計
  Given BTCUSDT 一小時內有十二筆持倉統計
  When 助理指定 BTCUSDT 與這一小時查持倉統計
  Then 問到合約那一條線的持倉統計,帶著合約標的與起訖時間

Scenario: 沒設定帳戶金鑰時分級是空的,說明早已告知
  Given 交易服務沒有設定帳戶金鑰
  When 助理讀查維持保證金分級的說明
  Then 說明寫著沒有帳戶金鑰時回空的,而且那不是錯誤

Scenario: 沒指定合約標的被擋下
  Given 查資金費率結算、持倉統計或維持保證金分級時沒指定合約標的
  When 送出
  Then 在送到交易服務之前就被擋下,並說出缺的是合約標的那一格
```

### US-04 — 說明寫出送出前不可能知道的事 [priority: P0]

**As a** 助理，**I want** 每件合約能力的說明寫出規則背後的意思，**so that** 我不必靠被拒絕或讀錯數字才發現。

```gherkin
Scenario: 資金費率說明寫出誰付給誰
  When 助理讀查資金費率結算的說明
  Then 說明寫著費率為正時做多的人付給做空的人
  And 寫著結算時間不要自行取整

Scenario: 持倉統計說明寫出只有三十天
  When 助理讀查持倉統計的說明
  Then 說明寫著來源只留最近三十天

Scenario: 合約 K 線說明寫出空的不是零
  When 助理讀查合約 K 線的說明
  Then 說明寫著指數價格與溢價指數空的那幾根是舊資料,並指出同步那一段就會補上

Scenario: 合約標的清單說明寫出只是最小那一級
  When 助理讀列出合約標的的說明
  Then 說明寫著交易規格裡的維持保證金率只是最小那一級,並指出完整分級要另外查

Scenario: 加入合約追蹤名單說明寫出要等與代號不對應
  When 助理讀加入合約追蹤名單的說明
  Then 說明寫著要等二十秒左右
  And 寫著現貨 SHIBUSDT 在合約叫 1000SHIBUSDT

Scenario: 合約同步進度說明寫出編號自己一串
  When 助理讀看合約同步進度的說明
  Then 說明寫著合約的同步輪次編號是自己那一串
```

### US-05 — 加入與移出合約追蹤名單 [priority: P0]

```gherkin
Scenario: 加入合約追蹤名單
  Given 使用者要追 BTCUSDT 永續
  When 助理加入合約追蹤名單
  Then 問到合約那一條線的追蹤名單,帶著 BTCUSDT

Scenario: 移出只停止追蹤
  When 助理讀移出合約追蹤名單的說明
  Then 說明寫著只停止追蹤、已存下的一筆都不刪、現貨不受影響
```

### US-06 — 其他能力不受影響 [priority: P0]

```gherkin
Scenario: 能力清單剛好多了這十四件
  Given 這一刀之前的五十一件能力
  When 看能力清單
  Then 清單是那五十一件加上合約的十四件,不多也不少,而且每一件要不要身分與原本相同

Scenario: 合約能力都不需要身分
  When 看合約的十四件能力
  Then 沒有一件需要身分
```

---

## 4. Business Flow & Logic

- 外掛只轉達、不把關：必填欄位沒填在送出前擋下，其餘一切合法性由交易服務判斷，拒絕的原話原樣帶回。
- 每一件能力的說明寫給助理讀：做什麼，以及**什麼會讓它被拒絕、回來的東西怎麼讀**。

## 5. UI/UX Design & Interaction

N/A

## 6. Non-Functional Requirements

N/A

## 7. Dependencies & Risks

- 依賴交易服務 PR #64（資金費率結算、持倉統計、完整維持保證金分級三件能力要它合併部署後才有東西可問）。

## 8. Appendix

- 需求共識：`BRIEF.md`
