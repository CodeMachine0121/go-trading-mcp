# Contract Traceability Matrix — 回測改為送出後查看（外掛）

Contract: PRD.md
Design map: ARCH.md
Implementation: `cmd/server/tool_catalog_strategy.go`（四件重演＋`backtestRunApiTools`）、`cmd/server/tool_catalog.go`、`internal/domain/models/domains/replay_result_domain.go`、`cmd/server/dependencies.go`、`cmd/server/config.go`
Oracle: Acceptance Criteria (19 clauses)

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-1 | 合約腳本回測送出就回 | 立刻拿到輪次（id、status）；助理告訴使用者編號、可隨時回來看 | 四件重演宣告（不再 `.Waiting`）＋ `backgroundReplayNote`「它不等算完」「把輪次編號說給使用者」 | `tool_catalog_backtest_run_test.go` `TestEveryReplayTeachesSubmittingThenLookingAtTheRun`；`TestOnlyWhatIsReplayedOnTheSpotOrLookedAtWithAWaitWaitsLonger` | asserts-oracle（說明句＋ResponseWaitLimit=0） | produces-oracle（回覆原樣轉交交易服務的輪次） | ✅ conforms |
| AC-2 | 四個回測能力的輸入一字不改 | 四件的路由、動詞、必填欄位與今天相同 | 四件宣告的參數未動 | `TestEveryReplayStillSendsWhatItSentBefore`；既有 `TestEveryReplayForwardsTheFillTimingAndTheValidationStart` 等 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-3 | 不合法的要求拒絕原話帶回、沒有輪次 | 交易服務的拒絕原話 | `ApiToolDomain.Relayed` 只精簡成功回覆（未改）；四件不再精簡 | `TestOnlyLookingAtABacktestRunCondensesWhatItHandsOn`（四件原封）；既有 refusal 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-4 | 回測能力不再比別的能力等更久 | ResponseWaitLimit 為一般等待 | 四件不再 `.Waiting` | `TestOnlyWhatIsReplayedOnTheSpotOrLookedAtWithAWaitWaitsLonger` | asserts-oracle（植入 `.Waiting` 會紅） | produces-oracle | ✅ conforms |
| AC-5 | 完成的輪次帶精簡成績單 | result 內最近 100 筆、200 點，並說總數；成績單數字不動 | `ReplayResultDomain` 把 `result` 當一個部分精簡；`trading_get_backtest_run` `.CondensingReplayResults()` | `replay_result_domain_test.go` `TestABacktestRunDetailIsCondensedInsideItsResult`、`TestTheReportCardIsNeverTouched`（既有）；`TestOnlyLookingAtABacktestRunCondensesWhatItHandsOn` | asserts-oracle（植入拿掉 result 會紅） | produces-oracle | ✅ conforms |
| AC-6 | 順便等一下、提前結束就提前回 | waitSeconds 送到交易服務 | `queryParameter("waitSeconds")` | `TestLookingAtABacktestRunCanWaitAWhile`（Query 帶 waitSeconds=30） | asserts-oracle | produces-oracle（等待本身由交易服務執行） | ✅ conforms |
| AC-7 | 等滿仍未結束回當時狀態 | 回進行中與進度，沒有成績單 | 交易服務回什麼就轉什麼；沒有 result 的回覆原封 | `TestABacktestRunWithoutAResultIsHandedOnUntouched` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-8 | 最多等六十秒 | 助理給 600 時只等 60 | 說明寫「最多 60」；上限由交易服務裁 | `TestLookingAtABacktestRunTeachesHowToReadIt`（「最多 60」） | asserts-oracle（說明） | partial：外掛不自己裁到 60，靠交易服務 | ⚠️ partial（交易服務裁，外掛說明＋等待上限 120 秒兜底） |
| AC-9 | 失敗原因照說 | 回覆失敗與原因；助理不猜成腳本壞了 | 說明「failureReason 照原話說」「不是腳本壞了」 | `TestLookingAtABacktestRunTeachesHowToReadIt` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 別人的輪次「找不到」 | 404 原話帶回 | 說明「不是自己的輪次回 404」；拒絕原話既有路徑 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-11 | 列出新到舊並說總數 | 20 筆、說總數 25 | `trading_list_backtest_runs`，說明「回覆時說出總數」 | `TestListingBacktestRunsFiltersOnlyWhatWasAsked` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-12 | 只看失敗的 | status=failed 送出 | query `status` | 同上（Query 帶 status=failed） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-13 | 取消排隊中的 | #42 已取消 | `trading_cancel_backtest_run` Submit `/backtest-runs/{id}/cancel` | `TestCancellingABacktestRunNamesTheRun` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-14 | 已結束的拒絕原話帶回 | 交易服務原話 | 說明「會被拒絕並說它已經結束」；拒絕原話既有路徑 | 同上（「已經結束」） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-15 | 使用者問昨天跑過的 | 助理用列出能力找、不重跑 | 說明「不要憑記憶重跑」 | `TestListingBacktestRunsFiltersOnlyWhatWasAsked` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-16 | 排隊中的說法 | 「前面還有在跑的，會自動開始」、不叫重送 | `backgroundReplayNote`、get 說明 | `TestEveryReplayTeachesSubmittingThenLookingAtTheRun`、`TestLookingAtABacktestRunTeachesHowToReadIt` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-17 | 每人上限滿了 | 原話轉達並建議等一筆結束或取消一筆 | `backgroundReplayNote`「你同時最多 N 筆…取消一筆再送」 | `TestEveryReplayTeachesSubmittingThenLookingAtTheRun` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-18 | 連續兩次呼叫落在不同分身 | 兩次都正常回覆 | `StreamableHTTPOptions{Stateless: true}` | `dependencies_test.go` `TestTheAssembledConnectorAnswersACallThatCarriesNoSession`（initialize 後 tools/list 不帶 session 仍 200） | asserts-oracle（植入 `Stateless: false` 會紅） | produces-oracle | ✅ conforms |
| AC-19 | 重新連線的指引不變 | 回覆請到 /mcp 重新連線 | 既有 `FailureReasonDomain`、Instructions 未動 | 既有 `TestTheAssembledConnectorOffersEveryAbilityOverTheWire`（Instructions 含 /mcp）與 failure reason 測試 | asserts-oracle | produces-oracle | ✅ conforms |

## Orphans (code with no clause)

- 移除 `IDLE_CONNECTION_TIMEOUT_MINUTES`：AC-18 的必然結果（無狀態下沒有連線可放掉），非違反。

## Summary

- Conforms: 18/19 clauses ✅
- Partial: 1（AC-8：60 秒上限由交易服務裁，外掛只在說明寫明並以 120 秒等待兜底；外掛不複製交易服務的規則，與既有「宣告不帶規則」原則一致）
- Violations / Mis-asserted / Gaps: —
- Orphans: 0

Note: static conformance audit against the Acceptance Criteria — it judges test assertions and code paths against the spec's expected outcome, not by running the full suite.
