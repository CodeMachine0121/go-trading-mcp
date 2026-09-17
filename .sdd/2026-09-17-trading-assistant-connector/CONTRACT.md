# 交易助理外掛 — Contract Verification

**Contract source:** `.sdd/2026-09-17-trading-assistant-connector/PRD.md`（Acceptance Criteria 為 oracle）
**Design map:** `ARCH.md`（scenario → component）
**Glossary:** `.sdd/UL-MAP.md`
**Ceiling:** 這是一份**靜態符合度稽核**——逐條把「測試斷言了什麼」與「程式產出什麼」各自對照規格推導出的預期結果，**不是**跑一次測試看綠燈。它不自行捏造情境去執行。

---

## Clauses

### US-01 — 把帳號交給外掛

| ID | Clause | Oracle（只由規格推導） | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01 | 電子郵件與密碼對得上 | 回成功；訊息含電子郵件與到期時刻；**無任何憑證**；密碼不留存 | `authentication_service.go:48`、`authentication_controller.go:76` | `authentication_application_test.go:17`、`mcp_controller_test.go:196` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-02 | 密碼打錯 | 拒絕，訊息＝交易服務原話 | `authentication_service.go:63` | `authentication_application_test.go:33`、`mcp_controller_test.go:365` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03 | 交易服務暫時簽不出登入 | 拒絕，並轉述「暫時不能簽發登入」 | `trading_service_proxy.go:242`（非 2xx→refused，原話帶回） | `mcp_controller_test.go`（`ATradingServiceThatCannotSignAnybodyInSaysSoInItsOwnWords`，並斷言不說成「連不到」） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-04 | 換一個帳號登入 | 連線上的身分變成後者；前者不再使用 | `authentication_service.go:71`、`signed_in_session_repository.go:42` | `authentication_application_test.go:74`、`signed_in_session_repository_test.go:33` | asserts-oracle | produces-oracle | ✅ conforms |

### US-02 — 沒有身分就不代辦要身分的事

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-05 | 已登入代辦要身分的事 | 成功；請求帶該連線的憑證 | `api_tool_service.go:109` | `api_tool_application_test.go:32` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-06 | 未登入代辦要身分的事 | 拒絕並說「請先登入」；**完全沒有送出請求** | `api_tool_service.go:123`、`failure_reason_domain.go:37` | `api_tool_application_test.go:45`、`mcp_controller_test.go:155` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-07 | 未登入代辦不需要身分的事 | 照常回覆；請求不帶憑證 | `api_tool_domain.go:80`、`api_tool_service.go:118` | `api_tool_application_test.go:54`、`trading_service_proxy_test.go:96` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-08 | 未登入建立一位使用者 | 照常建立；回覆不含密碼 | `tool_catalog.go`（`trading_register_user`，不需要身分） | `mcp_controller_test.go`（`BuildingAUserNeedsNoSignInAndReachesTheTradingServiceAsIs`）、`tool_catalog_test.go:83` | asserts-oracle | produces-oracle | ✅ conforms |

### US-03 — 過期不打擾人

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-09 | 登入還沒過期 | 直接送出，不續用 | `signed_in_session_domain.go:50`、`authentication_service.go:114` | `api_tool_application_test.go:32`、`signed_in_session_domain_test.go:20` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 過期但續用還有效 | 先換一份新的，再把原事做完；使用者無感 | `authentication_service.go:118`、`:153` | `api_tool_application_test.go:65` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-11 | 續用也失效 | 拒絕並說「登入已失效，請重新登入」 | `authentication_service.go:171`、`failure_reason_domain.go:42` | `api_tool_application_test.go:80`、`:90` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-12 | 換新時連不到交易服務 | 拒絕並說連不到；**不說「請重新登入」** | `failure_reason_domain.go:47` | `api_tool_application_test.go:105`（明確斷言不含「請重新登入」） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-13 | 同一份續用不會被換兩次 | 續用只被呼叫 1 次；兩件事都完成 | `session_renewal_gate.go:35`、`authentication_service.go:166` | `api_tool_application_test.go:119`、`:343`（`Times(1)` 且有同步柵欄） | asserts-oracle | produces-oracle | ✅ conforms |

### US-04 — 登出

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-14 | 已登入時登出 | 回成功；之後要身分的事說「請先登入」 | `authentication_service.go:79` | `authentication_application_test.go:96`、`mcp_controller_test.go:212` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-15 | 本來就沒登入時登出 | 回成功，不算失敗，不呼叫交易服務作廢 | `authentication_service.go:84` | `authentication_application_test.go:115`、`mcp_controller_test.go:231` | asserts-oracle | produces-oracle | ✅ conforms |

### US-05 — 兩個人互不相見

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-16 | 各自的連線各自的身分 | 甲帶甲的憑證、乙帶乙的 | `mcp_caller.go:35`、`signed_in_session_repository.go:31` | `api_tool_application_test.go:161`、`signed_in_session_repository_test.go:21` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-17 | 沒登入的連線借不到別人的身分 | 拒絕並說請先登入；不使用甲的身分 | 同上 | `api_tool_application_test.go:147`、`mcp_controller_test.go:303` | asserts-oracle | produces-oracle | ✅ conforms |

### US-06 — 自備身分

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-18 | 呼叫端自備有效憑證 | 以那份憑證的主人身分完成 | `mcp_caller.go:44`、`api_tool_service.go:114` | `api_tool_application_test.go:174`、`mcp_controller_test.go:241` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-19 | 自備的優先於連線保管的 | 用自備的那一份 | `api_tool_service.go:114`（最先判斷） | `api_tool_application_test.go:190` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-20 | 自備的已失效 | 說登入已失效；**不嘗試續用、不重試** | `api_tool_service.go:152` | `api_tool_application_test.go:207`（`Times(1)` 鎖死不重試） | asserts-oracle | produces-oracle | ✅ conforms |

### US-07 — 把交易服務的話原話帶回來

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-21 | 一切正常查 K 線 | 回那段 K 線 | `api_tool_domain.go:111`、`trading_service_proxy.go:51` | `trading_service_proxy_test.go:77`、`mcp_controller_test.go:143` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-22 | 回溯天數超過上限 | 拒絕，並帶回「上限是幾天」 | `trading_service_response_vo.go`（`ToToolResultDto`） | `api_tool_application_test.go:256`（一字不差斷言） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-23 | 沒登錄過的交易標的 | 拒絕，並說那個標的不存在 | `trading_service_proxy.go:242`（404→refused、原話） | `trading_service_proxy_test.go:107`（404 列）、`mcp_controller_test.go:259` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-24 | 讀別人的、沒上架的策略腳本 | 拒絕，並轉述交易服務的說法 | 同上（403 →refused、原話） | `trading_service_proxy_test.go:107`（403 列） | asserts-oracle | produces-oracle | ✅ conforms |
| ~~AC-25~~ | ~~今日助手額度用盡~~ | — | **已移除**：行情對話助手不再被代理（見下方決策），這條情境隨之從契約移除。429 的原話轉述本身仍有測試守著（`trading_service_proxy_test.go:107`） | — | — | — | ➖ withdrawn |
| AC-26 | 交易服務整個連不上 | 拒絕並說「連不到交易服務」，明確不是使用者做錯了 | `trading_service_proxy.go:230`、`failure_reason_domain.go:47` | `api_tool_application_test.go:269`、`trading_service_proxy_test.go:155` | asserts-oracle | produces-oracle | ✅ conforms |

### US-08 — 看一眼即時更新

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-27 | 那個標的正在更新 | 回收到的那則更新與它的狀態 | `trading_service_proxy.go:156` | `trading_service_proxy_test.go:238` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-28 | 等滿了仍然沒有更新 | 回「這段時間沒有更新」；**不算失敗**；不繼續等 | `trading_service_proxy.go:186` | `trading_service_proxy_test.go:258` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-29 | 那個標的分不到即時名額 | 回「沒有即時名額」，並說明這不會自己好 | `trading_service_proxy.go:178`（原話轉述 `status:"unavailable"`）＋ `tool_catalog_market.go`（能力說明解釋「不會自己好」） | `mcp_controller_test.go`（`AnAbilityWithNoLiveSlotSaysSoAndSaysItWillNotFixItself`，同時驗轉述與說明） | asserts-oracle | produces-oracle | ✅ conforms |

### US-09 — 一件不漏

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-30 | 能力清單涵蓋每一件事 | 九類共 48 件轉達能力全在清單上，且**行情對話助手那三件不在** | `cmd/server/tool_catalog*.go` | `tool_catalog_test.go:83`（對照手寫全集，相等斷言）、`dependencies_test.go:21`（真的掛上去了） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-31 | 每件事說得出要不要身分與要填什麼 | 每件都帶說明、欄位與身分需求 | `api_tool_domain.go:85`、`api_tool_controller.go:71` | `tool_catalog_test.go:103`、`mcp_controller_test.go:117` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-32 | 交易服務新增了一件事而外掛還沒補上 | 那一件不在清單上 | 清單是宣告式的，未宣告即不存在 | `tool_catalog_test.go:83`（相等斷言，多一件少一件都紅） | asserts-oracle | produces-oracle | ✅ conforms |

### Core Business Rules

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| BR-1 | 身分以連線為單位，連線之間互不相見 | 只能以連線鍵取得身分 | `i_signed_in_session_repository.go`（只有以連線鍵存取的方法） | `signed_in_session_repository_test.go:21` | ✅ conforms |
| BR-2 | 自備身分優先，且外掛不替它續用 | 自備者先用；401 時不續用 | `api_tool_service.go:114`、`:152` | `api_tool_application_test.go:190`、`:207` | ✅ conforms |
| BR-3 | 續用只做一次，大家共用換來的那一份 | 並行時續用只被呼叫一次 | `session_renewal_gate.go`、`authentication_service.go:166` | `api_tool_application_test.go:119`、`:343` | ✅ conforms |
| BR-4 | 過期與失效是兩句不同的話 | 兩者絕不共用同一句 | `failure_reason_domain.go:37`/`:42` | `failure_reason_domain_test.go:14` | ✅ conforms |
| BR-5 | 拒絕與連不上是兩句不同的話 | 兩者絕不共用同一句 | `failure_reason_domain.go:47`、`trading_service_response_vo.go` | `failure_reason_domain_test.go:45` | ✅ conforms |
| BR-6 | 密碼用完即丟；憑證不出現在回答裡 | 回答不含密碼與兩份憑證 | `signed_in_session_domain.go:74`、`authentication_controller.go:100` | `mcp_controller_test.go:196`、`signed_in_session_domain_test.go:77` | ✅ conforms |
| BR-7 | 外掛不重複交易服務的規則 | 只判斷「會不會做這件事」與「有沒有身分」 | `api_tool_service.go:64`（僅未知能力、必填欄位、身分三種自判） | `api_tool_application_test.go:281`、`:290` | ✅ conforms |
| BR-8 | 清單說的是外掛真的做得到的 | 未宣告即不在清單上 | `tool_catalog*.go` | `tool_catalog_test.go:83` | ✅ conforms |
| BR-9 | 即時更新只看一眼 | 第一則即回、最長等上限 | `trading_service_proxy.go:156` | `trading_service_proxy_test.go:238`、`:258` | ✅ conforms |

### Non-Functional Requirements

| ID | Clause | Oracle | Implementation | Test | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| NFR-1 | 外掛自己不做任何運算 | 沒有任何行情/指標計算 | domain 只有請求組裝與身分判斷 | 結構性，無專屬測試 | 🟡 partial |
| NFR-2 | 看一眼必須有等待上限，絕不無限等 | 等滿即收線 | `trading_service_proxy.go:157` | `trading_service_proxy_test.go:258` | ✅ conforms |
| NFR-3 | 密碼只在換登入那一瞬間存在 | 之後不留 | `authentication_service.go:48`（不存 `SignInDto`） | `mcp_controller_test.go:196` | ✅ conforms |
| NFR-4 | 憑證不寫檔、不寫紀錄、不放進回覆 | 三者皆無 | `signed_in_session_repository.go`（純記憶體）、`attempt_record.go`（只寫三個欄位） | `mcp_controller_test.go`（`EveryAttemptLeavesATraceThatNamesNoSecret`：明確斷言紀錄不含兩份憑證）＋`:196` | ✅ conforms |
| NFR-5 | 執行紀錄裡不得出現密碼或憑證 | 紀錄不含兩者 | `attempt_record.go`（只寫能力名稱、成敗、失敗類別） | `mcp_controller_test.go`（`EveryAttemptLeavesATraceThatNamesNoSecret`、`ATraceSaysWhichKindOfFailureItWas`） | ✅ conforms |
| NFR-6 | 一個連線讀不到另一個連線的身分 | 讀不到 | `signed_in_session_repository.go:31` | `api_tool_application_test.go:147`、`mcp_controller_test.go:303` | ✅ conforms |
| NFR-7 | 一般的助理用戶端連得上，不限某一家 | 標準 MCP over HTTP | `dependencies.go:60`（`StreamableHTTPHandler`） | `mcp_controller_test.go`（真的 MCP client）、`dependencies_test.go:21` | ✅ conforms |
| NFR-8 | 每次代辦留得下「做了哪件事、成不成功、失敗屬於哪一類」，且不含憑證與密碼 | 每次代辦留下一筆含能力名稱與結果類別的紀錄 | `attempt_record.go`、`assistant_reply.go`（在唯一的出口處留下） | `mcp_controller_test.go`（`ATraceSaysWhichKindOfFailureItWas`） | ✅ conforms |

---

## Orphans

| # | Behaviour | Clause? | 判定 |
| :--- | :--- | :--- | :--- |
| O-1 | `trading_renew_session`（手動立刻換一份新登入） | BRIEF 的能力清單有「續用登入」，PRD 無專屬 scenario | ⚠️ orphan（良性：功能在需求共識裡，只是沒被寫成驗收情境。建議補一條 AC） |
| O-2 | 外掛自己的 `GET /health` | 無 | ⚠️ orphan（良性：維運用，不是業務能力，且不在 Out of Scope 之列） |

**Out of Scope 反向檢查**：PRD 列的六項（不存行情、不改規則、不做持續推送、不記密碼、不做自己的資料庫、**不代理行情對話助手**）逐項核對，**沒有任何一項被實作** — 無違規。最後一項另有專屬測試守著（`tool_catalog_test.go` 的 `NoAbilityLetsOneAssistantSpendAnothersBudget`），因為它是一條關於「不存在」的規定，而不存在的東西沒有別的方式驗得出來。

---

## Summary

| | |
| :--- | :--- |
| ✅ conforms | 47 |
| ➖ withdrawn | 1（AC-25，隨範圍變更移除） |
| 🟡 partial | 1（NFR-1） |
| ❌ gap | 0 |
| 🔴 violation | 0 |
| 🟠 mis-asserted | 0 |
| ❔ unclear | 0 |
| ⚠️ orphan | 2（皆良性） |

**Conformance: 98%（48 條中 47 條完全符合）**

### 第一輪的發現與處置

| 第一輪 | 處置 |
| :--- | :--- |
| ❌ NFR-8 — 沒有任何逐次代辦的紀錄 | 已補：在**唯一的回覆出口**留下能力名稱、成敗與失敗類別（`attempt_record.go`）。刻意**不寫**欄位內容、回覆內容、憑證、密碼與使用者——紀錄是唯一會被複製進 bug report 的產物，進去了就是外洩了。連帶讓 NFR-4／NFR-5 從結構性宣稱變成有測試守著 |
| 🟡 AC-29 — 分不到即時名額沒有專屬測試 | 已補，同時驗「轉述 `unavailable`」與「能力說明講出它不會自己好」 |
| 🟡 AC-03 — 交易服務簽不出登入 | 已補，並額外斷言它**不被說成「連不到交易服務」**——它答話了，只是答不出憑證 |
| 🟡 AC-08 — 未登入建立使用者 | 已補端到端案例 |

### 仍然未竟

- **NFR-1（外掛自己不做任何運算）** 🟡 — 這是一條關於「沒有什麼」的規定，只能由結構保證（domain 裡沒有任何行情或指標運算），沒有一條測試寫得出它。維持 partial 是誠實的說法，不是待辦。
- **O-1（`trading_renew_session`）** ⚠️ — 功能在需求共識（BRIEF）的能力清單裡，但 PRD 沒有寫成驗收情境。行為已有測試覆蓋（`RenewingOnDemandReachesTheAssistantAsOneOfThreeDifferentAnswers`）。建議下次動這塊時補一條 AC，不必現在改程式。

---

## 範圍變更紀錄

**2026-09-18 — 移除行情對話助手的三件能力（提問、列出對話、讀一段對話）。**

理由不是成本，是誰在做決定：代理它等於讓**一個 AI 自己決定去花另一個 AI 的錢**。
那個助手一次回答來回數十趟、要好幾分鐘、按 token 計費，而中間沒有任何人判斷過
這一次值不值得。使用者自己去用它沒有問題；不該存在的是「助理可以自己伸手拿」這件事。

處置：從能力清單移除（48 件轉達能力），BRIEF／PRD 的 Out of Scope 各補一條，
AC-25 隨之撤回，並補一條**專屬的反向測試**——不存在的能力沒有辦法被誤呼叫，
也沒有哪次重構能不小心把它放回來。
