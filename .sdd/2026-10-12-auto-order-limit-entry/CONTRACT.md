# Contract Traceability Matrix — 自動下單限價進場（外掛）

Contract: PRD.md · Design map: ARCH.md · Implementation: `cmd/server/`（`internal/` 未動）
Oracle: Acceptance Criteria (15 scenarios) + Business Rules (5) + NFR (1) = 21 clauses

Path abbreviations: `auto` = `cmd/server/tool_catalog_automation.go`, `key` = `cmd/server/tool_catalog_binance_trading_key.go`,
`dep` = `cmd/server/dependencies.go`, `LE` = `cmd/server/tool_catalog_auto_order_limit_entry_test.go`,
`BA` = `cmd/server/tool_catalog_binance_auto_order_test.go`, `DT` = `cmd/server/dependencies_test.go`,
`PX` = `internal/infrastructure/tradingservice/tests/trading_service_proxy_test.go`.

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-01-1 | 建立時沒提容忍度 | 送出內容不含容忍度；說明寫沒給就是 0.1 | auto `strategyBotWriteParameters`（非必填）；`ApiToolDomain.BuildRequest` 只送填了的 | LE TestAnUngivenEntrySlippageTolerance…；TestTheEntrySlippageToleranceSaysWhatItMeans…「建立時不給就是 0.1」 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-01-2 | 給了容忍度 | 送出內容帶 0.05 | auto 容忍度 body 欄位 | LE TestAGivenEntrySlippageToleranceTravelsAsGiven（"0.05"） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-01-3 | 容忍度 0 | 送出帶 0；說明寫 0 可能常常不成交 | 同上 + 欄位說明 | LE 同上（"0"）＋說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-01-4 | 超出範圍 | 助理拿到交易服務原話；說明事先寫 0 到 5（含）、百分點 | `TradingServiceProxy.responseOf` 原樣轉交；欄位說明 | PX TestTheVerdictFollowsWhatTheTradingServiceAnswered（400 原文）；LE 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-01-5 | 改寫時沒給就保留 | 改寫內容不含容忍度；說明列為整份改寫的例外 | auto update 說明「例外有三樣」 | LE TestAnUngiven…（update）＋ TestRewritingAStrategyBotSaysTheEntrySlippageToleranceIsKeptWhenNotGiven | asserts-oracle | produces-oracle | ✅ conforms |
| AC-02-1 | 查看一台 | 回覆原樣；查看與列出說明提到容忍度 | auto `strategyBotAutoOrderNote` | BA TestEveryAbilityThatAnswersWithAStrategyBotSaysAutoOrderReallyTrades（含 list/get） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03-1 | 全部成交 | 說明讓助理說出全部成交與滑價正數＝不利 | auto `contractAutoOrderResultNote` | LE TestReadingABotsRoundsSaysHowEachAutoOrderEntryWent | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03-2 | 部分成交 | 說明讓助理說出「部分成交 X／Y，其餘已撤單」 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03-3 | 沒有成交 | 說明讓助理說出沒有成交、已撤單、沒有開倉 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03-4 | 等待成交中 | 說明讓助理說出還在等、最多一個觸發間隔 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-04-1 | 進場單還掛著 | 拒絕原話與原因原樣；說明要網頁關自動下單、等撤單、外掛不能關 | proxy 原樣轉交；auto delete 說明 | PX（409 原文）；LE TestDeletingABotSaysWhatToDoWhileItsEntryOrderStillRests | asserts-oracle | produces-oracle | ✅ conforms |
| AC-04-2 | 沒有掛著的進場單 | 照常刪除 | auto delete（路徑與動作未變） | 無針對本情境的新測試；既有轉達路徑 | no-test | produces-oracle | 🟡 partial（未變的既有路徑，接受） |
| AC-05-1 | 列出機器人說實情 | 說明說會真的開倉平倉；任何說明與整體指引不再出現「還不會下單」 | key `autoOrderReallyTradesNote`；auto；dep | BA TestEveryAbility…ReallyTrades、TestNoAbilitySaysAutoOrderDoesNotPlaceOrdersYet（含欄位說明）；DT TestTheConnectorInstructions… | asserts-oracle | produces-oracle | ✅ conforms |
| AC-05-2 | 要求打開自動下單 | 說明只能網頁做；不送改動 | 無對應能力；`binanceAutoOrderWebOnlyNote` | BA TestNoAbilityWritesABinanceTradingKeyOrSwitchesAutoOrder、TestWritingAStrategyBotNeverForwardsAutoOrder | asserts-oracle | produces-oracle | ✅ conforms |
| BR-1 | 容忍度以文字精確小數、不送就不填預設、不檢查範圍 | 字串欄位、非必填、外掛無範圍檢查 | auto | LE TestWritingAStrategyBotOffersAnOptionalEntrySlippageTolerance | asserts-oracle | produces-oracle | ✅ conforms |
| BR-2 | 整份改寫三個例外 | 說明列 marketDataKind、容忍度、自動下單 | auto update 說明 | LE + `tool_catalog_contract_strategy_bot_test.go` TestRewritingAStrategyBotSaysTheKindIsKept + BA TestWritingAStrategyBotSaysAutoOrderIsSwitchedOnlyOnTheWeb | asserts-oracle | produces-oracle | ✅ conforms |
| BR-3 | 下單結果原樣轉交 | 外掛不改寫成功回覆 | `ApiToolDomain.Relayed`（只精簡重演結果） | PX 成功回覆原文 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-4 | 刪除被擋原樣轉交 | 同 AC-04-1 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-5 | 自動下單的說法 | 合約開倉限價、平倉市價、掛止損止盈；現貨仍不下單；開關與金鑰只在網頁 | key `autoOrderReallyTradesNote`、`binanceAutoOrderWebOnlyNote` | BA、DT | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-1 | 外掛不能下單、撤單、切換、經手金鑰 | 清單裡沒有這類能力 | catalog | BA TestNoAbilityWritesABinanceTradingKeyOrSwitchesAutoOrder；`tool_catalog_test.go` 能力全集 | asserts-oracle | produces-oracle | ✅ conforms |

## Orphans

| Behavior | Site | Classification |
| :--- | :--- | :--- |
| 下單結果說明也解釋 `protectionMissing`、`action`、平倉／開倉欄位 | auto `contractAutoOrderResultNote` | 交易服務上一個切片就有的欄位，外掛先前沒說明；與合約不衝突，屬補全 |

Out-of-scope checklist：下單／撤單／切換自動下單／存取金鑰能力——均未新增 ✔；外掛範圍檢查——未實作 ✔

## Summary

✅ 20 conforms · 🔴 0 violations · 🟠 0 mis-asserted · 🟡 1 partial · ❌ 0 gaps · ❔ 0 unclear · ⚠️ 1 orphan
Conformance: 95%（20/21）

AC-04-2 is the unchanged existing delete path; accepted without a new test.
Static conformance audit against the Acceptance Criteria; it judges assertions and code paths against the spec's expected outcomes and does not execute invented scenarios.
