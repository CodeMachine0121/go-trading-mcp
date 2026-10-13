# 合約訂單流（足跡）外掛說明 — Architecture Design

**Status:** Confirmed（autonomous run）
**Source PRD:** `.sdd/2026-10-14-contract-order-flow-footprint/PRD.md`

## 1. Design Goal

只改兩段說明文字，不動轉送邏輯：參數 `parameters` 的說明（`strategyScriptBodyParameters` 裡那一格）與 `contractKCandleScriptNote`。

## 2. Change Scope

| Area | Action | Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog.go` 參數說明 | **Modify** | 四種種類、各自讀法、訂單流價位寬度只給合約 |
| `cmd/server/tool_catalog_strategy.go` `contractKCandleScriptNote` | **Modify** | 列出 Footprint 與 PriceLevel、宣告了才有、上限、不落地 |
| 參數轉送 | **Not touched** | 已是原樣轉送（`json.RawMessage`），規則由交易服務判 |

## 3. New Classes / Modules

無。

## 6. Extensibility

- 交易服務再加參數種類時，改的是同一格說明；`contractKCandleScriptNote` 是建立、修改、讀取策略腳本共用的那一句。

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| 四種參數種類、只給合約 | 參數說明 |
| 原樣轉送 | 既有轉送（測試釘住） |
| 足跡的每一項、宣告了才有與上限 | `contractKCandleScriptNote` |
