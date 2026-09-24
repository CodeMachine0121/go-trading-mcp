# Contract Conformance — 看一眼合約即時更新

**Contract:** `PRD.md` · **Implementation:** `cmd/server/tool_catalog_contract.go:166` (one catalogue entry) · static audit, no invented scenarios executed.

## Clauses

| ID | Clause | Oracle (from spec) | Implementation | Test | Test audit | Code audit | Status |
|---|---|---|---|---|---|---|---|
| AC-1 | 看一眼合約標的 | 助理問的是合約那一邊的即時入口、帶著那個合約標的，回第一則更新 | `cmd/server/tool_catalog_contract.go:166` read `/contract-k-candles/live`, query symbol, `.Watching` | `TestPeekingAtAContractAsksItsOwnLiveLineForTheFirstUpdate` (cmd/server/tool_catalog_contract_live_peek_test.go) + `TestWatchingStopsAtTheFirstUpdate` (internal/infrastructure/tradingservice/tests/trading_service_proxy_test.go:239) | asserts-oracle | produces-oracle | ✅ |
| AC-2 | 現貨的看一眼照舊 | 現貨那一件照舊走現貨入口、說明一字不差 | `tool_catalog_market.go` untouched | `TestPeekingAtSpotIsUnchangedByTheContractTwin`, spot path assertion | asserts-oracle | produces-oracle | ✅ |
| AC-3 | 等滿沒有更新 | 回「這段時間沒有收到即時更新，不是錯誤」 | proxy watch path (shared) | `TestWatchingAQuietMarketIsNotAFailure` (internal/infrastructure/tradingservice/tests/trading_service_proxy_test.go:259) + wait-limit assertion | asserts-oracle | produces-oracle | ✅ |
| AC-4 | 不在合約追蹤名單上 | 照交易服務的說法回；說明指向加進合約追蹤名單那一件 | refusal pass-through + description | `TestWatchingSomethingTheTradingServiceRefusesIsStillARefusal` (internal/infrastructure/tradingservice/tests/trading_service_proxy_test.go:277); description asserts `trading_add_to_contract_watchlist` | asserts-oracle | produces-oracle | ✅ |
| AC-5 | 系統不認得的代號 | 照交易服務的說法回：找不到 | refusal pass-through + description | same proxy test; description asserts 「系統不認得的代號回找不到」 | asserts-oracle | produces-oracle | ✅ |
| AC-6 | 成形中 | 說明：還在走、數字會變、系統不存 | description | `TestPeekingAtAContractSaysWhatItShowsAndWhatGetsItRefused` | asserts-oracle | produces-oracle | ✅ |
| AC-7 | 走完了 | 說明：不是由即時更新存下，每分鐘那一輪一分鐘內存入 | description | same | asserts-oracle | produces-oracle | ✅ |
| AC-8 | 即時更新斷了 | 說明：系統自己重連，等一下就好 | description | same | asserts-oracle | produces-oracle | ✅ |
| AC-9 | 內容是最新價 | 說明：最新價開高低收與成交量、不含標記價格；不會出現分不到名額或休市 | description | same | asserts-oracle | produces-oracle | ✅ |
| BR-1 | 看法與現貨相同、同一個上限 | 等待上限與現貨相同 | `.Watching(liveUpdateWaitLimit)` shared | wait-limit equality + `TestOnlyWatchingAnAbilityStaysOnTheLine` | asserts-oracle | produces-oracle | ✅ |
| BR-2 | 不需要登入 | 不帶身分 | entry `false` | `CarriesIdentity` false + sign-in map | asserts-oracle | produces-oracle | ✅ |
| BR-3 | 拒絕原樣轉給助理 | 交易服務的話原樣回 | proxy | internal/infrastructure/tradingservice/tests/trading_service_proxy_test.go:277 | asserts-oracle | produces-oracle | ✅ |
| BR-4 | 名字與說明帶合約 | 名字帶 contract、只問合約那一條 | entry name/path | `TestContractAbilitiesOnlyEverReachTheContractLine` | asserts-oracle | produces-oracle | ✅ |
| NFR-1 | 絕不維持一直開著的通道 | 看一眼後收掉連線 | proxy `context.WithTimeout` + first-event return | internal/infrastructure/tradingservice/tests/trading_service_proxy_test.go:239, internal/infrastructure/tradingservice/tests/trading_service_proxy_test.go:259 | asserts-oracle | produces-oracle | ✅ |

## Orphans

None. Out-of-scope items (a standing channel, live mark price, spot changes) are not implemented.

## Summary

14 clauses — ✅ 14 conforms · 🔴 0 · 🟠 0 · 🟡 0 · ❌ 0 · ❔ 0 · ⚠️ 0. Conformance 100%.
AC-1/3/4/5 rely on the proxy's generic watching tests for the response half; the ability-specific half (address, symbol, wait limit, no identity) is pinned by this slice's tests.
