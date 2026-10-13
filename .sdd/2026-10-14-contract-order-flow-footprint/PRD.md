# Product Requirements Document (PRD)

**Status:** Draft · **Version:** v1.0 · **Owner:** James Hsueh

## 1. Background & Goal

- **Problem Statement:** 外掛的參數說明還停在「只有兩種」，也沒有提到合約行情格的訂單流足跡，助手照著寫會寫出缺了是非、不會用足跡的策略腳本。
- **Expected Outcome:** 助手讀到的說明與交易服務的規則一致。
- **Out of Scope:** 訂單流查詢工具。

## 2. User Personas

- 透過外掛代使用者寫策略腳本的助手。

## 3. User Stories & Acceptance Criteria

### US-01 — 參數種類說得完整 [P0]

```gherkin
Scenario: 四種參數種類
  When 助手讀建立策略腳本的參數說明
  Then 說明列出 lookbackCount、number、boolean、orderFlowPriceStep 四種,各自說出它是什麼與怎麼讀

Scenario: 訂單流價位寬度只給合約
  When 助手讀參數說明
  Then 說明說出 orderFlowPriceStep 只給 marketDataKind 為 contractKCandle 的策略腳本、至多一個、大於零、最細 0.00000001

Scenario: 原樣轉送
  Given 助手送出一個 kind 為 orderFlowPriceStep 的參數
  When 建立策略腳本
  Then 交易服務收到的參數一字不差
```

### US-02 — 合約行情說明有足跡 [P0]

```gherkin
Scenario: 足跡的每一項
  When 助手讀合約行情算式的說明
  Then 說明列出 Footprint（indicator.Footprint）的 PriceStep、Levels、Delta、PointOfControl
  And 列出 indicator.PriceLevel 的 Price、BuyVolume、SellVolume、TradeCount

Scenario: 宣告了才有、上限與不落地
  When 助手讀合約行情算式的說明
  Then 說明說出沒宣告訂單流價位寬度時足跡是空的、一次至多 31 天、一格至多 1,000 個價位、訂單流每次現拿
```

## 4. Business Flow & Logic

- 外掛只說明、只轉送，規則由交易服務判。

## 5–7

N/A；依賴交易服務的 `orderFlowPriceStep` 參數種類。
