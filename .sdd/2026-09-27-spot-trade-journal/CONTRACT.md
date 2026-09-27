# 現貨交易日誌（外掛）— Contract Conformance

**Contract:** `PRD.md`（同資料夾）· **Design map:** `ARCH.md` §7 · **Glossary:** `.sdd/UL-MAP.md`
**Ceiling:** 靜態符合性稽核——依 PRD 驗收情境判斷測試斷言與程式路徑，不執行自創情境。

檔案簡稱：`spot.go`＝`cmd/server/tool_catalog_spot_trade_journal.go`、`spot_test.go`＝`cmd/server/tool_catalog_spot_trade_journal_test.go`、
`contract.go`／`contract_test.go`＝`cmd/server/tool_catalog_trade_journal(_test).go`。

---

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AC-01 | 口述一張台股 | 送往新增現貨交易；說明要助理把一張換成 1,000 股並說出股數請確認、回覆編號 | spot.go:44 | `TestEverySpotAbilitySaysWhatWillGetItRefusedAndWhatToDo`、`TestSpotQuantitiesPricesAndFeesSayWhatTheAssistantMustKnow` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-02 | 口述加密貨幣現貨 | 送往新增現貨交易，第一筆買進巢狀送出 | spot.go:44 | `TestRecordingASpotTradeSendsTheFirstBuyNested`、路由表 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03 | 現貨沒有做空 | 說明寫明只有先買後賣、說做空不送出 | spot.go:20 | 同 AC-01 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-04 | 現貨沒有槓桿 | 說明寫明沒有槓桿；任何現貨能力都沒有槓桿、方向一格 | spot.go:20 | `TestASpotTradeOffersNoDirectionAndNoLeverage` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-05 | 現貨與合約都有的代號先問 | 說明寫明先問現貨還是合約、得到回答前不送 | spot.go:22 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-06 | 只貼機器人訊息不當成成交 | 說明與價格欄寫明參考價不是成交價、先問 | spot.go:10 | 說明測試、`TestSpotQuantitiesPricesAndFeesSayWhatTheAssistantMustKnow` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-07 | 同一標的已有持有中 | 拒絕原話帶回；說明指向加買進 | spot.go:44 | 說明測試、`TestARefusalComesBackInTheTradingServicesOwnWords`（internal/application/tests） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-08 | 沒說時間以現在為準 | 省略時不送時間；說明要以台北時間回報並請確認 | spot.go:29 | `TestSpotDefaultsAreLeftToTheTradingService`、`TestSpotQuantitiesPricesAndFeesSayWhatTheAssistantMustKnow` | shallow（台北時間那句沒被釘住） | produces-oracle | 🟠 mis-asserted → 已修正（4e1bc16） |
| AC-09 | 分批賣出 | 送往加買賣 | spot.go:72 | 路由表 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 全部賣出即平倉 | 說明要回覆淨損益與報酬率並提醒寫檢討 | spot.go:72 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-11 | 賣出超過持有 | 說明寫明會被拒絕；拒絕原話帶回 | spot.go:72 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-12 | 台股必須整數股 | 數量欄寫明台股整數股 | spot.go:13 | 數量說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-13 | 沒說手續費就是 0 | 省略時不送手續費；說明寫明沒說即 0 並提醒證交稅 | spot.go:16 | `TestSpotDefaultsAreLeftToTheTradingService`、手續費說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-14 | 說了含稅手續費就照填 | 給的手續費原樣送出 | spot.go:33 | — | no-test | produces-oracle | 🟡 partial → 已補 `TestAFeeTheUserGaveIsForwardedAsGiven`（4e1bc16） |
| AC-15 | 修正持有中的一筆買賣 | 說明要先讀出再帶上 | spot.go:81 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-16 | 平倉後不能修正買賣 | 說明寫明平倉後鎖定、加附註或刪除重記 | spot.go:81 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-17 | 止損必須低於買進價 | 計畫欄寫明止損低於第一筆買進價 | spot.go:24 | `TestASpotPlanSaysWhichSideTheStopBelongsOn` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-18 | 只改一項計畫也不清掉其他項 | 說明要先讀出現在的計畫再帶上 | spot.go:99 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-19 | 平倉後計畫鎖定 | 說明寫明鎖定並建議加附註 | spot.go:99 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-20 | 平倉後寫檢討 | 送往檢討；說明寫明平倉後才能寫、失誤標籤與合約共用 | spot.go:118 | 說明測試、路由表 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-21 | 看現貨統計依市場分組 | 說明寫明兩組分開、金額不跨幣別加總 | spot.go:170 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-22 | 平均 R 以有止損的筆數計 | 說明要說出以幾筆計 | spot.go:170 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-23 | 期間內沒有平倉 | 說明寫明勝率不適用、不是 0% | spot.go:170 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-24 | 查看持有中的現貨交易 | 說明寫明浮動損益是估算、沒有資金費用與強平價 | spot.go:153 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-25 | 現貨實盤 vs 回測 | 送往現貨對照、等重演；說明寫明重演不計成本 | spot.go:179 | 說明測試、`TestOnlyTheSpotBacktestComparisonWaitsLikeAReplay` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-26 | 列出現貨交易可依市場篩選 | 市場篩選原樣送出；說明寫明台股即 market=taiwanStock、說出總數 | spot.go:140 | `TestSpotDefaultsAreLeftToTheTradingService`、說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-27 | 加倉 | 合約加紀錄說明稱加倉、開倉價 | contract.go:79 | `TestContractRecordsAreCalledByTheirPositionNames`、`TestEveryAbilitySaysWhatWillGetItRefused` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-28 | 減倉與平倉 | 稱減倉、平倉價 | contract.go:21 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-29 | 合約能力的說明不再出現成交 | 合約日誌能力說明與欄位說明都沒有「成交」 | contract.go | `TestNoContractJournalAbilitySpeaksOfFills` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-30 | 確認後刪除 | 送往刪除；說明寫明一併刪除 | spot.go:160 | 路由表、說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-31 | 指的不只一筆先問 | 說明寫明先列出請指定、不猜 | spot.go:160 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-32 | 拒絕原話帶回 | 交易服務拒絕原話轉達 | 既有 `ApiToolService` | `TestARefusalComesBackInTheTradingServicesOwnWords` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-33 | 別人的交易 | 查看說明寫明別人的是找不到；原話帶回 | spot.go:153 | 說明測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-34 | 沒有欄位可以指定擁有者 | 任何現貨能力都沒有擁有者一格 | spot.go | `TestNoSpotAbilityLetsTheAssistantNameAnOwner` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-1 | 現貨與合約能力各自一組、名字帶現貨或合約 | 清單同時有兩組、名稱不重複 | tool_catalog.go:206 | `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse`、`TestNoTwoAbilitiesShareAName` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-2 | 標籤兩本共用 | 標籤能力說明寫明共用、刪除兩本一起算 | contract.go:192 | `TestEveryAbilitySaysWhatWillGetItRefused` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-3 | 手續費率只用於合約 | 費率能力說明寫明只用於合約 | contract.go:184 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-1 | 實盤 vs 回測沿用重演等待上限、逾時不重送 | 對照能力等 120 秒；說明寫明不要立刻重送 | spot.go:186 | `TestOnlyTheReplaysWaitLonger`、說明測試 | asserts-oracle | produces-oracle | ✅ conforms |

---

## Orphans

| Behavior | Site | Classification |
| :--- | :--- | :--- |
| 合約第一筆紀錄 `kind` 省略即開倉（說明改寫） | contract.go:63 | PRD Appendix 已記載的決定；對齊交易服務，非範圍外 |

---

## Summary

- 38 clauses（34 AC、3 BR、1 NFR）。
- 初次稽核：36 ✅ · 1 🟠（AC-08）· 1 🟡（AC-14）· 0 🔴 · 0 ❌。
- 修正後：38 ✅，符合率 100%（`4e1bc16` 補上兩處斷言）。
