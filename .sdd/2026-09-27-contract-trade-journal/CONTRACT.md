# Contract Traceability Matrix — contract-trade-journal（外掛）

Contract: PRD.md
Design map: ARCH.md
Implementation: `cmd/server/tool_catalog_trade_journal.go`（能力宣告）＋既有 `internal/`（轉達、拒絕原話、重新連線、連不到）
Oracle: Acceptance Criteria (40 clauses) + Core Business Rules (10) + NFR (2)

外掛的交付物是「能力宣告」：能力存在、送往對的地方、必填與選填正確、沒填就不送、說明寫出助理必須知道的事。
「助理實際怎麼回覆」取決於模型讀說明後的行為，超出靜態稽核範圍；以下對助理行為類情境，判定的是外掛能控制的那一半（說明裡有那句話且被測試釘住）。

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-01 | 口述記一筆 | 送往新增交易；第一筆成交以巢狀送出；說明要助理回覆編號、均價、計畫風險、記下的時間 | tool_catalog_trade_journal.go:49 | tool_catalog_trade_journal_test.go `TestRecordingATradeSendsTheFirstEntryFillNested`；contract_test 路由表 | shallow（回覆內容那句沒被釘住） | produces-oracle | 🟠 mis-asserted → 已修正 |
| AC-02 | 沒說成交時間以現在為準 | 省略時間時不送時間（交易服務以現在記下）；說明要求以台北時間回報並請確認 | :12, :75 | `TestLeavingOutTheFillTimeLeavesItToTheTradingService`、`TestTheFillTimeSaysNowIsTheDefaultAndHowToReportIt` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-03 | 說了成交時間就照他說的 | 說了時間即原樣送出（RFC3339 帶時區） | :19 | `TestGivenFillTimeIsForwarded`、FillTime 說明測試（RFC3339） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-04 | 只貼機器人訊息不當成成交 | 說明寫明參考價不是成交價、先問 | :10 | `TestTheFillPriceWarnsThatAReferencePriceIsNotAFill` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-05 | 同標的同方向已有持倉中 | 拒絕原話帶回；說明建議改用加成交 | :52 | `TestEveryAbilitySaysWhatWillGetItRefused`、`TestARefusalComesBackInTheTradingServicesOwnWords` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-06 | 關聯自己的合約交易策略 | 有「依據策略」一格、選填、填了即送出；說明只能是自己的合約交易策略 | :71 | `TestRecordingATradeRequiresOnlyWhatTheTradingServiceRequires` | shallow（只釘必填與否，沒釘說明與送出） | produces-oracle | 🟠 mis-asserted → 已修正 |
| AC-07 | 交易服務拒絕時原話帶回 | 拒絕原話帶回 | 既有 `ApiToolService` | `TestARefusalComesBackInTheTradingServicesOwnWords` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-08 | 加碼 | 送往該筆交易的加成交 | :88 | 路由表、`TestLeavingOutTheFillTimeLeavesItToTheTradingService` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-09 | 全部平倉 | 送往加成交；說明提醒可寫檢討 | :88 | 路由表、Refused 測試（可以寫檢討） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-10 | 出場超過持倉 | 拒絕原話；說明寫出反手請先平倉 | :90 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-11 | 持倉中修正打錯的成交 | 送往修正成交（整筆取代） | :96 | 路由表、requiredness 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-12 | 持倉中刪除一筆記錯的成交 | 送往刪除成交 | :105 | 路由表、Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-13 | 平倉後成交鎖定 | 拒絕原話；說明建議加附註或刪整筆 | :98 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-14 | 持倉中修改計畫 | 送往修改計畫 | :112 | 路由表、requiredness | asserts-oracle | produces-oracle | ✅ conforms |
| AC-15 | 平倉後修改計畫被拒 | 拒絕原話；說明建議加附註 | :113 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-16 | 加附註 | 送往加附註；說明附註不改計畫 | :118 | 路由表、Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-17 | 口述檢討 | 送往寫檢討；評分必填、失誤標籤選填 | :125 | 路由表、requiredness | asserts-oracle | produces-oracle | ✅ conforms |
| AC-18 | 持倉中寫檢討被拒 | 拒絕原話；說明只能平倉後寫 | :126 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-19 | 列出待檢討的交易 | 狀態篩選送出；說明待檢討＝已平倉 | :147 | `TestTheDefaultsTheDescriptionsPromiseAreLeftToTheTradingService`、Refused（status=closed） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-20 | 列出交易預設最近 20 筆 | 沒說筆數不送筆數；說明寫出最近 20 筆、說出總數 | :147 | Defaults 測試、Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-21 | 指定要幾筆 | 筆數原樣送出 | :153 | Defaults 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-22 | 算不出的數字照原話帶回 | 說明寫不要說成 0 | :160 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-23 | 看統計 | 送往統計 | :176 | 路由表 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-24 | 沒說期間就看最近 30 天 | 沒說期間不送；說明寫出最近 30 天 | :181 | Defaults 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-25 | 期間內沒有已平倉交易 | 說明寫不適用、不是 0% | :178 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-26 | 看實盤 vs 回測 | 送往對照；等待時間同重演 | :184 | 路由表、`TestOnlyTheBacktestComparisonWaitsLikeAReplay` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-27 | 別人的交易 | 找不到原話帶回 | 既有 | `TestARefusalComesBackInTheTradingServicesOwnWords` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-28 | 設定手續費率 | 送往設定費率；說明舊成交不變 | :196 | 路由表、Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-29 | 修正手續費率 | 同上 | :196 | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-30 | 費率為負被拒 | 說明寫不得為負；拒絕原話 | :197 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-31 | 新增型態標籤 | 送往新增標籤；類別與名稱必填 | :208 | 路由表、requiredness | asserts-oracle | produces-oracle | ✅ conforms |
| AC-32 | 標籤重名被拒 | 說明同類不能重名 | :209 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-33 | 標籤改名 | 送往改名；說明交易跟著改 | :215 | 路由表、Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-34 | 使用中的標籤刪除被拒 | 說明使用中不能刪、說出幾筆 | :221 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-35 | 確認後刪除 | 送往刪除；說明成交附註檢討一併刪除 | :164 | 路由表、Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-36 | 指的不只一筆時不猜 | 說明先列出請他指定、不要猜 | :166 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-37 | 未確認不刪 | 說明刪除前必須先確認 | :166 | Refused 測試 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-38 | 沒有欄位能指定擁有者 | 任何日誌能力都沒有擁有者一格 | 全檔 | `TestNoTradeJournalAbilityLetsTheAssistantNameAnOwner` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-39 | 交易服務不認得身分 | 請到外掛選單重新連線、不重試 | 既有 | `TestAConnectorAuthorizationTheTradingServiceDoesNotRecognizeAsksToReconnectWithoutRetrying` | asserts-oracle | produces-oracle | ✅ conforms |
| AC-40 | 連不到交易服務 | 說暫時連不到、晚點再試 | 既有 | `TestNotReachingTheTradingServiceIsNotToldAsTheCallersMistake` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-01 | 日誌能力全屬合約線、名字帶合約 | 交易相關能力名帶 contract、只到合約入口；標籤與費率屬使用者設定 | 全檔 | `TestContractAbilitiesOnlyEverReachTheContractLine`、`TestTheCatalogueCovers...` | asserts-oracle | produces-oracle | ✅ conforms |
| BR-02 | 業務規則全由交易服務決定 | 外掛不驗證、不重算 | 全檔（純宣告） | 設計即保證 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-03 | 身分 | 同 AC-38 | — | 同 AC-38 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-04 | 成交價 | 同 AC-04 | — | 同 AC-04 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-05 | 成交時間 | 同 AC-02、AC-03 | — | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-06 | 列出交易：可指定筆數與狀態、合約標的、**期間**篩選 | 列出交易有期間篩選一格 | — | — | no-test | not-implemented | ❌ gap（交易服務的列出入口沒有期間篩選，見下） |
| BR-07 | 統計期間 | 同 AC-24 | — | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-08 | 刪除交易 | 同 AC-36、AC-37 | — | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-09 | 算不出／不適用／估算 | 同 AC-22、AC-25 | — | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-10 | 失敗的三種說法 | 同 AC-07、AC-39、AC-40 | — | 同上 | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-01 | 一般等待、對照照重演等待 | 只有對照等比較久 | :189 | `TestOnlyTheBacktestComparisonWaitsLikeAReplay`、`TestOnlyTheReplaysWaitLonger` | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-02 | 授權不保管、沒有擁有者欄位 | 同 AC-38；授權沿用既有 | 既有 | 同 AC-38 | asserts-oracle | produces-oracle | ✅ conforms |

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| `trading_set_contract_trade_setup_tags` | 整組取代一筆交易的型態標籤 | 對應 US-05「新增型態標籤」的貼上一步，PRD 沒有獨立情境；undocumented，但屬交易服務既有路由，保留 |
| `trading_get_trade_journal_settings` | 查看手續費率 | 同上，屬「設定／修正費率」的前置查看；保留 |

## Summary

- Conforms: 49/52 clauses ✅（94%，修正後 51/52）
- Violations: —
- Mis-asserted: AC-01、AC-06（已補測試釘住說明與送出）
- Partial: —
- Gaps: BR-06（列出交易的期間篩選）——交易服務的列出入口（後端設計）只有狀態、合約標的、筆數，外掛無從轉交。
  處置：不在外掛宣告一格交易服務不認得的參數；需由交易服務補上期間篩選後，外掛再加一格。已回報主流程跨 repo 核對。
- Unclear: —
- Orphans: 2（皆保留）

靜態稽核：依驗收條件判斷測試斷言與程式路徑，不執行自創情境；助理實際回覆行為超出外掛可驗證範圍。
