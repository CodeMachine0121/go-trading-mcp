# Contract Verification — 加入市集的策略腳本是拿到一份副本（外掛）

Oracle: `PRD.md` Acceptance Criteria — 3 clauses. Static audit; each mapped test was run once and passed.

| ID | Clause | Oracle | Code | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01 | 沒有放棄採用這個工具 | 工具清單不含放棄採用 | `cmd/server/tool_catalog_strategy.go`（該項已移除） | `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse`（清單與期望集合逐項相等，期望集合已不含它） | ✅ conforms |
| AC-02 | 加入的說明講清楚是複製 | 提到複製一份、看不到算式、改不動、與原本那一支無關、同名被拒、不要了就刪 | `trading_adopt_strategy_script` 說明 | `TestAdoptingSaysItHandsOverACopyThatTheAuthorCanNoLongerChange` | ✅ conforms |
| AC-03 | 交易策略的說明講清楚只能用自己的 | 建立與改寫都提到只能指名自己的（含副本），別人的要先加入 | `ownStrategyScriptsOnlyNote` 接在兩者說明後 | `TestBothTradingStrategyWritesSayOnlyOnesOwnScriptsMayBeNamed` | ✅ conforms |

## Orphans
- 列出、讀取、改寫腳本、上架、下架、瀏覽市集的說明也一併改成副本的說法——與交易服務現況一致，屬於 BRIEF 的 Requirements，不是範圍外。

## Summary
✅ 3 conforms · Conformance 100%
