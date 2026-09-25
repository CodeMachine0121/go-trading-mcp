# Contract Conformance — 改行情的事要帶身分

**Contract:** `PRD.md` · **Implementation:** `cmd/server/tool_catalog_market.go`, `cmd/server/tool_catalog_contract.go` (identity flags only) · static audit, no invented scenarios executed.

帶不帶身分、沒身分不去問、拒絕原話帶回，是每一件能力共用的同一條路（`internal/domain/service/api_tool_service.go:118` 依 `RequiresSignIn()` 分流，`internal/domain/models/domains/api_tool_domain.go:184` 把它寫成 `CarriesIdentity`）。
所以每一條「這件事帶身分」的條款由兩段組成：這件能力的旗標（本刀）＋那條共用的路（既有測試）。

## Clauses

| ID | Clause | Oracle (from spec) | Implementation | Test | Test audit | Code audit | Status |
|---|---|---|---|---|---|---|---|
| AC-1 | 已登入的人把標的加進觀察清單 | 這一次詢問帶著他的身分；回覆原樣帶回 | `cmd/server/tool_catalog_market.go:142` | `TestChangingMarketDataTravelsUnderTheSignedInIdentityWhileReadingItDoesNot/trading_add_to_watchlist` + `TestAnAbilityThatNeedsIdentityTravelsUnderThisConnectionsProof` (`internal/application/tests/api_tool_application_test.go:32`) | asserts-oracle | produces-oracle | ✅ |
| AC-2 | 已登入的人同步一段合約歷史 | 這一次詢問帶著他的身分 | `cmd/server/tool_catalog_contract.go:154` | `…/trading_sync_contract_k_candle_history` | asserts-oracle | produces-oracle | ✅ |
| AC-3 | 看歷史同步的進度也帶著身分 | 這一次詢問帶著他的身分 | `cmd/server/tool_catalog_contract.go:167`, `cmd/server/tool_catalog_market.go:104` | `…/trading_get_contract_k_candle_history_sync`, `…/trading_get_k_candle_history_sync` | asserts-oracle | produces-oracle | ✅ |
| AC-4 | 沒有人登入就不去問 | 外掛不去問交易服務，回「請先登入」 | `cmd/server/tool_catalog_market.go:72` + `internal/domain/service/api_tool_service.go:118-126` | `…/trading_delete_k_candle` + `TestAnAbilityThatNeedsIdentityIsNotEvenAttemptedWithoutOne` (`internal/application/tests/api_tool_application_test.go:45`, no send expected, content 「請先登入」) | asserts-oracle | produces-oracle | ✅ |
| AC-5 | 帳號還沒被放行時原話帶回 | 帶著身分去問；「尚未開通」與開通指示原樣帶回 | `cmd/server/tool_catalog_market.go:82` | `…/trading_backfill_k_candles` + the not-activated refusal test (`internal/controller/tests/mcp_controller_test.go:707-733`) | asserts-oracle | produces-oracle | ✅ |
| AC-6 | 需要身分的正是這十六件 | 八件 × 兩條線需要身分，其他能力不變 | the sixteen flags in both catalog files | `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse` (literal map, `cmd/server/tool_catalog_test.go:27`) | asserts-oracle | produces-oracle | ✅ |
| AC-7 | 沒登入也查得到一段 K 線 | 詢問不帶身分 | `cmd/server/tool_catalog_market.go:29` | `…/trading_list_k_candles` + `TestAnAbilityThatNeedsNoIdentityTravelsWithoutOne` | asserts-oracle | produces-oracle | ✅ |
| AC-8 | 沒登入也看得到合約的即時更新 | 詢問不帶身分 | `cmd/server/tool_catalog_contract.go:188` | `…/trading_peek_live_contract_k_candle`, `TestPeekingAtAContractAsksItsOwnLiveLineForTheFirstUpdate` | asserts-oracle | produces-oracle | ✅ |
| AC-9 | 沒登入也列得出合約交易標的 | 詢問不帶身分 | `cmd/server/tool_catalog_contract.go:200` | `…/trading_list_contract_trading_symbols` | asserts-oracle | produces-oracle | ✅ |

Falsified: flipping `/watchlist/{symbol}` back to anonymous turned the new table test and the catalogue test red; flipping the contract live peek to identified turned three tests red.

## Orphans

None. The diff touches only the sixteen flags, one stale one-line comment, and their tests.

## Summary

✅ 9 conforms · 🔴 0 · 🟠 0 · 🟡 0 · ❌ 0 · ❔ 0 · ⚠️ 0 — Conformance 100%.
