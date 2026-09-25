# Contract Conformance — 告訴助理：合約歷史同步也補持倉統計

**Contract:** `PRD.md` · **Implementation:** `cmd/server/tool_catalog_contract.go`, `cmd/server/tool_catalog_strategy.go` (description text only) · static audit, no invented scenarios executed.

## Clauses

| ID | Clause | Oracle (from spec) | Implementation | Test | Test audit | Code audit | Status |
|---|---|---|---|---|---|---|---|
| AC-1 | 同步的說明說出同一趟也補持倉統計 | 讀得到「同一趟也補持倉統計、用同一個回溯天數、先補完合約 K 線再補」與「只存沒有的」 | `cmd/server/tool_catalog_contract.go:149` | `TestTheContractAbilitiesSayWhatCannotBeDiscoveredBySending/trading_sync_contract_k_candle_history` (`cmd/server/tool_catalog_contract_test.go:287`) | asserts-oracle | produces-oracle | ✅ |
| AC-2 | 沒有那一天的檔案不算失敗 | 讀得到「沒有那一天的檔案不算失敗」 | `cmd/server/tool_catalog_contract.go:152` | same, phrase `沒有那一天的檔案不算失敗` | asserts-oracle | produces-oracle | ✅ |
| AC-3 | 看進度的說明說出持倉統計那一組分開 | 讀得到 positionStatistic 與五項，以及「與合約 K 線那組分開、不加總」 | `cmd/server/tool_catalog_contract.go:163` | `…/trading_get_contract_k_candle_history_sync` (`cmd/server/tool_catalog_contract_test.go:291`) | asserts-oracle | produces-oracle | ✅ |
| AC-4 | 歷史資料庫不答話不是失敗 | 讀得到「只停下持倉統計那一份、這趟仍算 succeeded」 | `cmd/server/tool_catalog_contract.go:165` (and :153 on the sync) | same, phrases `這趟仍算 succeeded`, `只停下持倉統計那一份` | asserts-oracle | produces-oracle | ✅ |
| AC-5 | 查持倉統計的說明指向合約歷史同步 | 讀得到「更早的可以用 trading_sync_contract_k_candle_history 補」，讀不到「從來沒有人錄」 | `cmd/server/tool_catalog_contract.go:244` | `…/trading_list_contract_position_statistics` mustSay + mustNotSay (`cmd/server/tool_catalog_contract_test.go:282`) | asserts-oracle | produces-oracle | ✅ |
| AC-6 | 寫合約策略腳本的說明不再說只留三十天 | 讀不到「持倉統計只留三十天」，讀得到「只有錄到或同步過的那段才有值」 | `cmd/server/tool_catalog_strategy.go:192` | `TestTheContractScriptNoteNamesEveryFigure` (`cmd/server/tool_catalog_contract_script_test.go:130`) + NotContains (:152) | asserts-oracle | produces-oracle | ✅ |
| AC-7 | 手動補齊的說明指向合約歷史同步 | 讀得到「只補 K 線」與「更早的持倉統計用 trading_sync_contract_k_candle_history」 | `cmd/server/tool_catalog_contract.go:136-137` | `…/trading_backfill_contract_k_candles` (`cmd/server/tool_catalog_contract_test.go:296`) | asserts-oracle | produces-oracle | ✅ |
| AC-8 | 能力清單不變 | 能力名稱、路徑與參數與之前一模一樣 | no catalogue entry added/removed; parameters untouched | `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse`, route table in `cmd/server/tool_catalog_contract_test.go:115`, `TestSyncingAContractHistoryStillAsksForOnlyTheSymbolAndTheLookback` | asserts-oracle | produces-oracle | ✅ |

Each phrase-level assertion was falsified by restoring the old wording (five mutants, all red).

## Orphans

None. The diff touches only the five descriptions the PRD names and their tests.

## Summary

✅ 8 conforms · 🔴 0 · 🟠 0 · 🟡 0 · ❌ 0 · ❔ 0 · ⚠️ 0 — Conformance 100%.
AC-8 was 🟡 partial on the first pass (parameters not pinned by a test) and was closed by adding `TestSyncingAContractHistoryStillAsksForOnlyTheSymbolAndTheLookback`.
