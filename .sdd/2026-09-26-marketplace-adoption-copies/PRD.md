# Product Requirements Document (PRD)

**Status:** Finalized · **Owner:** James Hsueh

## 1. Background & Goal
交易服務改了加入的意義並拿掉取消加入；外掛的工具清單與說法要一致。

## 3. User Stories & Acceptance Criteria
### US-01 — 工具清單與說法跟上 [P0]
```gherkin
Scenario: 沒有放棄採用這個工具
  When 代理列出外掛提供的工具
  Then 清單裡沒有放棄採用策略腳本

Scenario: 加入的說明講清楚是複製
  When 代理讀加入策略腳本的說明
  Then 說明提到這是複製一份給你、看不到算式、改不動、之後與原本那一支無關、同名會被拒絕、不要了就刪掉

Scenario: 交易策略的說明講清楚只能用自己的
  When 代理讀建立或改寫交易策略的說明
  Then 說明提到信號來源只能指名你自己的策略腳本（含從市集加入的副本），別人的要先加入
```

## 8. Appendix
- 交易服務切片：go-trading `.sdd/2026-09-26-marketplace-adoption-copies/`
