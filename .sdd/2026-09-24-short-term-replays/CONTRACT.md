# Contract Conformance — 助理做得了短線重演

**Oracle:** `PRD.md` §3 Acceptance Criteria（16 scenarios）· **Audit:** static, by reading tests and code against each scenario's expected outcome; mapped tests were run as corroboration only.

## Clauses

| ID | Scenario | Oracle (from spec) | Code | Test | Test audit | Code audit | Status |
|---|---|---|---|---|---|---|---|
| AC-1 | 成交時點照轉 | 送出的內容帶著下一格開盤成交 | `cmd/server/tool_catalog.go`（`fillTiming` 在 `backtestParameters()`）+ `ApiToolDomain.BuildRequest` | `TestEveryReplayForwardsTheFillTimingAndTheValidationStart` | asserts-oracle | produces-oracle | ✅ |
| AC-2 | 驗證起點照轉 | 送出的內容帶著 1/21 | 同上（`validationStartTime`） | 同上 | asserts-oracle | produces-oracle | ✅ |
| AC-3 | 兩格都沒給 | 送出的內容沒有這兩格 | `ToolArgumentsDomain.EncodedSubset` 只帶給了的 | `TestAReplayWithNeitherSendsWhatItSentBefore` | asserts-oracle | produces-oracle | ✅ |
| AC-4 | 拒絕照轉 | 助理收到的拒絕與交易服務一字不差 | `ApiToolDomain.Relayed`（只精簡成功） | `TestARefusedReplayReachesTheAssistantInTheTradingServicesWords`、`TestAReplayAbilityWaitsLongerAndCondensesOnlyWhatSucceeded` | asserts-oracle | produces-oracle | ✅ |
| AC-5 | 說明引導 | 說明說出 nextOpen、只調調參段只判驗證段、五格與不適用、時間上限與怎麼改 | `shortTermReplayNote`（四件都接上） | `TestEveryReplayTeachesShortTermReplaying` | asserts-oracle | produces-oracle | ✅ |
| AC-6 | 重演在等待時間內回覆 | 助理收到成績單 | `TradingServiceProxy.Send` 依 `ResponseWaitLimit` 等 | `TestAnAskThatWaitsLongerOutlastsTheUsualWait/a_longer_wait_receives_it` | asserts-oracle | produces-oracle | ✅ |
| AC-7 | 重演超過等待時間 | 助理收到「連不到交易服務」 | 同上 | `…/a_longer_wait_still_gives_up_past_its_own_limit` | asserts-oracle | produces-oracle | ✅ |
| AC-8 | 其他能力照舊 | 一般能力仍用三十秒 | `ResponseWaitLimit` 只由 `Waiting` 設 | `TestOnlyTheReplaysWaitLonger`、`…/the_usual_wait_gives_up_on_a_slow_answer` | asserts-oracle | produces-oracle | ✅ |
| AC-9 | 資金曲線取樣 | 剩 200 點，頭尾都在 | `ReplayResultDomain.ToCondensedContent` | `TestALongEquityCurveIsSampledKeepingBothEnds` | asserts-oracle | produces-oracle | ✅ |
| AC-10 | 資金曲線不多時不動 | 原封 | 同上（未縮短即回原文） | `TestAShortReplayResultIsHandedOnExactlyAsItArrived` | asserts-oracle | produces-oracle | ✅ |
| AC-11 | 交易明細只留最近一百筆 | 最近 100 筆、說出共 350 筆 | 同上 | `TestOnlyTheMostRecentTradesAreKeptAndTheTotalIsSaid` | asserts-oracle | produces-oracle | ✅ |
| AC-12 | 交易明細不多時不動 | 原封 | 同上 | `TestAShortReplayResultIsHandedOnExactlyAsItArrived` | asserts-oracle | produces-oracle | ✅ |
| AC-13 | 分段成績單也精簡 | 驗證段剩 200 點 | 同上（遞迴 `inSample`／`validation`） | `TestTheInSampleAndValidationPartsAreCondensedByTheSameRules` | asserts-oracle | produces-oracle | ✅ |
| AC-14 | 成績單數字不動 | 每個數字一字不差 | 同上（只動兩個清單） | `TestTheReportCardIsNeverTouched` | asserts-oracle | produces-oracle | ✅ |
| AC-15 | 被拒絕的回覆不動 | 原封 | `ApiToolDomain.Relayed` | `TestARefusedReplayReachesTheAssistantInTheTradingServicesWords` | asserts-oracle | produces-oracle | ✅ |
| AC-16 | 看不懂的回覆原封（§4） | 原封 | `ReplayResultDomain`（解不開即原文） | `TestWhatIsNotAReplayResultIsHandedOnUntouched` | asserts-oracle | produces-oracle | ✅ |

## Orphans

| Behavior | Site | Note |
|---|---|---|
| 精簡時不做 HTML 跳脫 | `replay_result_domain.go` | 不在 PRD；讓策略名稱裡的 `<`、`&` 原樣到達助理。已由 `TestCondensingDoesNotEscapeWhatTheTradingServiceWrote` 釘住 |
| 現貨重演一支腳本的說明拿掉「沒有資金曲線」 | `tool_catalog_strategy.go` | 那句原本就與交易服務不符；`TestTheSpotScriptReplayNoLongerSaysItHasNoEquityCurve` |

## Summary

✅ 16 conforms · 🔴 0 · 🟠 0 · 🟡 0 · ❌ 0 · ❔ 0 · ⚠️ 2 orphans（皆為說明性、非 Out of Scope）· Conformance 100%。
第一輪發現 AC-5 的測試只驗到指標名稱、沒驗「不適用」與「怎麼改」（shallow），已補上斷言。
