# Contract Conformance — 在瀏覽器登入外掛

**Contract:** `PRD.md` · **Implementation:** `internal/{domain,application,controller,infrastructure}` connector authorization + `api_tool_service.go`, `cmd/server/{config,dependencies,tool_catalog}.go` · static audit, no invented scenarios executed.

## Clauses

| ID | Clause | Oracle (from spec) | Implementation | Test | Test audit | Code audit | Status |
|---|---|---|---|---|---|---|---|
| AC-1 | 帶著有效外掛授權照常接待 | 有效、發給這個外掛的授權 → 照常列出會做的事 | `internal/controller/connector_authorization_controller.go:56`, `internal/domain/models/domains/connector_authorization_domain.go:35` | `TestTheAssembledConnectorOffersEveryAbilityOverTheWire` (`cmd/server/dependencies_test.go:62`), `TestAValidConnectorAuthorizationReachesTheConnectorAsItsUser` | asserts-oracle | produces-oracle | ✅ |
| AC-2 | 沒有帶授權就不接待 | 不接待，回覆指出授權說明位址 | SDK `auth.RequireBearerToken` via `connector_authorization_controller.go:39` | `TestACallWithoutAValidConnectorAuthorizationIsTurnedAwayAndPointedAtTheMetadata/沒有帶授權`, `TestTheAssembledConnectorTurnsAwayACallWithoutAConnectorAuthorization` (401 + exact `WWW-Authenticate`) | asserts-oracle | produces-oracle | ✅ |
| AC-3 | 失效的授權不接待 | 不接待，回覆指出授權說明位址 | `connector_authorization_domain.go:35` → `ErrConnectorAuthorizationRejected` → `auth.ErrInvalidToken` | `…/交易服務判定失效` (controller), application table `交易服務判定失效` | asserts-oracle | produces-oracle | ✅ |
| AC-4 | 發給別的服務的授權不算數 | 不接待 | `connector_authorization_domain.go:35` | `…/發給別的服務` (controller + application + domain) | asserts-oracle | produces-oracle | ✅ |
| AC-5 | 對象只差結尾斜線算同一個 | 照常接待 | `connector_authorization_domain.go:22,35` | `TestAValidConnectorAuthorizationReachesTheConnectorAsItsUser/只差結尾斜線`, domain tests | asserts-oracle | produces-oracle | ✅ |
| AC-6 | 問不到交易服務不說成授權失效 | 不接待；說「連不到交易服務」；不指引重新授權 | `trading_service_proxy.go` `InspectConnectorAuthorization` → `ErrTradingServiceUnreachable`; `connector_authorization_controller.go:69` passes it through as non-401 | `TestATradingServiceThatCannotBeReachedIsNotAnsweredAsARejectedAuthorization` (not 401, no `WWW-Authenticate`, body 連不到交易服務) | asserts-oracle | produces-oracle | ✅ |
| AC-7 | 一分鐘內再出現不再確認 | 三十秒後不再問交易服務 | `connector_authorization_service.go:38`, repository `Find` | `TestTheSameAuthorizationIsAskedAboutAgainOnlyOnceItsJudgementRunsOut/三十秒後再來` (`Times(1)`) | asserts-oracle | produces-oracle | ✅ |
| AC-8 | 滿一分鐘重新確認 | 六十秒後重新問 | `connector_authorization_domain.go:44`, repository `Find` | `…/滿一分鐘再來` (`Times(2)`) | asserts-oracle | produces-oracle | ✅ |
| AC-9 | 不記得比授權到期更久 | 二十秒後到期、二十五秒後重新問 | `connector_authorization_domain.go:44` | `…/二十秒後到期、二十五秒後再來` | asserts-oracle | produces-oracle | ✅ |
| AC-10 | 問不到的結果不記 | 上次問不到 → 這次重新問 | `connector_authorization_service.go` returns before `Save` | `TestNotBeingAbleToAskIsNeverRemembered` | asserts-oracle | produces-oracle | ✅ |
| AC-11 | 查授權說明 | 說明寫著外掛正式位址、授權伺服器公開位址、授權放在標頭 | `connector_authorization_controller.go:46`, `cmd/server/config.go` `ProtectedResourceUrl` | `TestTheMetadataSaysWhichResourceThisIsAndWhereToGetAnAuthorization`, `TestTheAssembledConnectorServesTheSameMetadataAtBothAddresses` (exact JSON) | asserts-oracle | produces-oracle | ✅ |
| AC-12 | 帶外掛路徑的位址回同一份說明 | 同一份說明 | `cmd/server/dependencies.go:39-40` | `TestTheAssembledConnectorServesTheSameMetadataAtBothAddresses` (both paths) | asserts-oracle | produces-oracle | ✅ |
| AC-13 | 改行情的事帶著授權去問 | 詢問帶著這一次的外掛授權；回覆原樣帶回 | `internal/domain/service/api_tool_service.go` `Send(…, toolCallDto.AccessToken)`, `internal/controller/mcp_caller.go` | `TestEveryAbilityForwardsTheCallersConnectorAuthorization/個人資源`, `TestEveryAbilityTravelsUnderTheConnectorAuthorizationTheCallerBrought` | asserts-oracle | produces-oracle | ✅ |
| AC-14 | 看行情的事也帶著授權去問 | 詢問帶著這一次的外掛授權 | same | `TestEveryAbilityForwardsTheCallersConnectorAuthorization/公開資料` (`trading_get_k_candle`) | asserts-oracle | produces-oracle | ✅ |
| AC-15 | 交易服務不認得身分時請他重新連線 | 回覆請他到 Claude Code 外掛選單重新連線；不重試 | `api_tool_service.go:72-73`, `ErrReconnectRequired` | `TestAnAuthorizationTheTradingServiceDoesNotRecognizeAsksToReconnectFromClaudeCode` (`/mcp`, 重新連線, asked once), application test with `Times(1)` | asserts-oracle | produces-oracle | ✅ |
| AC-16 | 帳號還沒開通時原話帶回 | 尚未開通的原話原樣帶回 | `trading_service_response_vo.go` refused relay | `TestAnAccountNobodyHasLetInYetIsRefusedInTheTradingServicesOwnWords` | asserts-oracle | produces-oracle | ✅ |
| AC-17 | 能力清單裡沒有對話裡的登入 | 無登入、登出、換新登入、建立使用者；有「我是誰」「改密碼」 | `cmd/server/tool_catalog.go` `accountApiTools`; `authentication_controller.go` removed | `TestTheAssembledConnectorOffersEveryAbilityOverTheWire` (equals literal list, which has get_current_user/change_password and no register), `TestTheAssistantIsShownEveryAbilityWithItsFormAndNoSignInNote` | asserts-oracle | produces-oracle | ✅ |
| AC-18 | 使用說明不再要密碼 | 說明不請它向使用者要電子郵件與密碼；能力說明不附「請先登入」 | `cmd/server/dependencies.go` Instructions; `api_tool_controller.go` `toMcpTool` | `TestTheAssembledConnectorOffersEveryAbilityOverTheWire` (no `trading_sign_in`, contains 「絕不要向使用者要電子郵件或密碼」 — strengthened in this audit), `TestTheAssistantIsShownEveryAbilityWithItsFormAndNoSignInNote` | asserts-oracle | produces-oracle | ✅ |
| NFR-1 | 不經手密碼、不保管身分；授權不出現在回覆或紀錄；只記指紋 | 紀錄裡沒有授權；保管的是指紋 | `persistence/connector_authorization_verdict_repository.go` (`fingerprintOf`), `attempt_record.go` | `TestEveryAttemptLeavesATraceThatNamesNoSecret` (trace has no token); fingerprint-only storage has no black-box test (not observable through the public interface) | shallow | produces-oracle | 🟡 |
| NFR-2 | 可同時跑不只一份，重啟不必重新授權 | 沒有跨請求的身分狀態 | identity store removed; only the verdict cache (per-replica, safe to lose) remains | none (structural) | no-test | produces-oracle | 🟡 |
| NFR-3 | 同一份授權一分鐘內至多確認一次 | 同 AC-7/8 | same as AC-7/8 | same as AC-7/8 | asserts-oracle | produces-oracle | ✅ |

## Orphans

| Behavior | Note |
|---|---|
| `TokenInfo.UserID` is set to the authorization's subject (`connector_authorization_controller.go:73`) | Lets the SDK refuse a connection reused by a different user. Not in the PRD; a hardening side effect of using the SDK middleware. Keep. |
| An unreachable trading service surfaces as HTTP 500 | The SDK maps any non-`ErrInvalidToken` verifier error to 500; the PRD only requires "not a rejected authorization", which holds. |
| `/health` stays open without authorization | Infrastructure probe, not an ability; PRD US-01 scopes the rule to abilities. |

## Summary

✅ 19 conforms · 🔴 0 · 🟠 0 · 🟡 2 partial · ❌ 0 · ❔ 0 · ⚠️ 3 orphans (documented, kept) — Conformance 90% (19/21).
