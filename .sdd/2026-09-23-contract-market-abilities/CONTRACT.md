# 契約追溯矩陣 — contract-market-abilities（能力清單接上永續合約那一條線）

Contract: PRD.md（Section 3 Gherkin → `AC-xx`；Section 4 → `BR-xx`；Section 6 為 N/A）
Design map: ARCH.md（Section 7 Traceability）
Implementation: `cmd/server/tool_catalog_contract.go`、`cmd/server/tool_catalog.go`、`internal/domain/models/domains/api_tool_domain.go`、`internal/domain/service/api_tool_service.go`
Tests: `cmd/server/tool_catalog_contract_test.go`、`cmd/server/tool_catalog_test.go`、`internal/application/tests/api_tool_application_test.go`
Branch: `feat/contract-market-tools`（交易服務對照：`go-trading` `feat/contract-market-context` 的 `cmd/server/dependencies.go`）
Oracle: Acceptance Criteria + Business Rules（24 條）

> 靜態契約稽核：oracle 只由 PRD 文字推導（Phase 2 先於讀程式碼完成），再分別判斷測試是否斷言 oracle、程式碼是否產出 oracle。
> 僅執行對應到條款的單一測試做佐證（8 支皆綠），判定不依 pass/fail。不撰寫新探針、不執行自創情境。

## 橋接（Phase 3 Step 0）

- 「合約那一條線」→ 交易服務以 `/contract-` 開頭的位址（UL-MAP「合約能力」、ARCH §3 表）；「現貨那一條線」→ 非 `/contract-` 位址（如 `/k-candles`）。
- 名字「帶合約」→ 能力名含 `contract`（UL-MAP：「名字一律帶 contract（合約）」）。
- 「送到交易服務之前就被擋下，並說出缺的那一格」→ `ApiToolDomain.BuildRequest` 回 `ErrRequiredArgumentMissing：<欄位名>`，`ApiToolService.CallApiTool` 在 `Send` 前即以 `ToolOutcomeInvalidArguments` 回覆（`internal/domain/service/api_tool_service.go:76-83`）。
- 指數價格 → `indexOpen/High/Low/Close`；溢價指數 → `premiumIndexOpen/High/Low/Close`；成交筆數 → `tradeCount`；合約標的 → `symbol`。
- 「要不要身分」→ `requiresSignIn`。

## 交易服務路由對照（位址與動詞）

外掛動詞對應 HTTP：`read→GET`、`submit→POST`、`replace→PUT`、`remove→DELETE`（`internal/infrastructure/tradingservice/trading_service_proxy.go:38-41`）。

| 能力 | 外掛（tool_catalog_contract.go） | 交易服務（go-trading dependencies.go） | 結果 |
|---|---|---|---|
| trading_create_contract_k_candle | POST `/contract-k-candles` (:69) | POST `/contract-k-candles` (:251) | 一致 |
| trading_list_contract_k_candles | GET `/contract-k-candles` (:84) | GET `/contract-k-candles` (:252) | 一致 |
| trading_get_contract_k_candle | GET `/contract-k-candles/{symbol}/{openTime}` (:90) | GET `/contract-k-candles/:symbol/:openTime` (:267) | 一致 |
| trading_update_contract_k_candle | PUT `/contract-k-candles/{symbol}/{openTime}` (:98) | PUT `/contract-k-candles/:symbol/:openTime` (:269) | 一致 |
| trading_delete_contract_k_candle | DELETE `/contract-k-candles/{symbol}/{openTime}` (:107) | DELETE `/contract-k-candles/:symbol/:openTime` (:271) | 一致 |
| trading_backfill_contract_k_candles | POST `/contract-k-candles/backfill` (:116) | POST `/contract-k-candles/backfill` (:253) | 一致 |
| trading_sync_contract_k_candle_history | POST `/contract-k-candles/history` (:127) | POST `/contract-k-candles/history` (:263) | 一致 |
| trading_get_contract_k_candle_history_sync | GET `/contract-k-candles/history/{id}` (:136) | GET `/contract-k-candles/history/:id` (:265) | 一致 |
| trading_list_contract_trading_symbols | GET `/contract-trading-symbols` (:148) | GET `/contract-trading-symbols` (:310) | 一致 |
| trading_add_to_contract_watchlist | POST `/contract-watchlist` (:158) | POST `/contract-watchlist` (:312) | 一致 |
| trading_remove_from_contract_watchlist | DELETE `/contract-watchlist/{symbol}` (:166) | DELETE `/contract-watchlist/:symbol` (:313) | 一致 |
| trading_list_contract_funding_rate_settlements | GET `/contract-funding-rate-settlements` (:177) | GET `/contract-funding-rate-settlements` (:278) | 一致 |
| trading_list_contract_position_statistics | GET `/contract-position-statistics` (:189) | GET `/contract-position-statistics` (:286) | 一致 |
| trading_get_contract_maintenance_margin_tiers | GET `/contract-maintenance-margin-tiers` (:200) | GET `/contract-maintenance-margin-tiers` (:283) | 一致 |

欄位名亦對照交易服務：K 線內文二十一個數字欄位 ＋ `symbol`/`openTime` 與 `KCandleContractRequest` json tag 一致（`go-trading internal/controller/models/k_candle_contract_request.go`）；查詢字串 `symbol`/`startTime`/`endTime` 與三個查詢 controller 的 `ginContext.Query(...)` 一致；backfill `symbol`、history `symbol`/`lookbackDays`、watchlist `symbol` 皆一致。**14 件無任何位址或動詞不符。**

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-01 | US-01 查合約 K 線問到合約那一條線：Given 使用者要看 BTCUSDT 永續合約這一小時的 K 線 / When 助理用查合約 K 線的能力 / Then 問到合約那一條線,回來的是帶標記價格、指數價格、溢價指數的合約 K 線 | 查合約 K 線的那一件問的是合約的 K 線（合約那一條線），且帶著 BTCUSDT 與這一小時的起訖；回來的內容就是合約 K 線（外掛原樣轉達） | tool_catalog_contract.go:76-86（GET `/contract-k-candles` ＋ `contractRangeParameters` :47-53） | tool_catalog_contract_test.go:68 `TestContractAbilitiesOnlyEverReachTheContractLine` | shallow — 只斷言位址以 `/contract-` 開頭（:79），不釘 `/contract-k-candles` 本身、也不釘 symbol/startTime/endTime 被放進查詢字串；指到 `/contract-funding-rate-settlements` 或漏帶起訖仍會綠 | produces-oracle | 🟠 mis-asserted |
| AC-02 | US-01 查現貨 K 線仍問到現貨那一條線：Given 使用者要看 BTCUSDT 現貨這一小時的 K 線 / When 助理用查 K 線的能力 / Then 問到現貨那一條線,與合約無關 | 查 K 線（現貨）那一件問到的是現貨 K 線，不碰合約 | tool_catalog_market.go:29（GET `/k-candles`，本分支未改動） | tool_catalog_contract_test.go:79 同上 | shallow — 只斷言「不是 `/contract-`」，未釘「問到現貨那一條線」（`/k-candles`）；位址改成任何非合約位址都會綠 | produces-oracle | 🟠 mis-asserted |
| AC-03 | US-01 名字分得出合約與現貨：Given 能力清單上的每一件能力 / When 看它的名字 / Then 合約的每一件都帶「合約」,現貨的沒有一件帶 | 十四件合約能力名字都帶 contract；其餘任何一件都不帶 | tool_catalog_contract.go:65-193（14 個名字）；tool_catalog.go:182 | tool_catalog_contract_test.go:81-82 | asserts-oracle（逐件掃整份清單，雙向） | produces-oracle | ✅ conforms |
| AC-04 | US-01 合約的能力不會問到現貨那一條線：Given 任何一件合約的能力 / When 它被送出 / Then 它只問到合約那一條線 | 每一件合約能力送出時位址都在合約那一條線 | tool_catalog_contract.go:69-200 | tool_catalog_contract_test.go:68-88 | asserts-oracle（所有欄位填滿後實際 BuildRequest，逐件檢查 `/contract-`，並計數 14） | produces-oracle | ✅ conforms |
| AC-05 | US-02 每一項都填了就整根轉達：Given 新增合約 K 線時每一項都填了 / And 溢價指數收盤是 -0.0005 / When 送出 / Then 交易服務收到的那一根帶著溢價指數收盤 -0.0005 與其他每一項 | 送出的那一根含溢價指數收盤 -0.0005（原值、負號保留）與其他每一項 | tool_catalog_contract.go:64-75, 15-43；api_tool_domain.go:134-143；tool_arguments_domain.go:60-77 | tool_catalog_contract_test.go:123 `TestAWrittenContractCandleCarriesItsPriceLines` | asserts-oracle（斷言 `"premiumIndexClose":"-0.0005"` 與 21 個欄位都在內文） | produces-oracle | ✅ conforms |
| AC-06 | US-02 沒填指數價格還沒送出就被擋下：Given 新增合約 K 線時沒填指數價格 / When 送出 / Then 在送到交易服務之前就被擋下,並說出缺的是指數價格那一格 | 沒送出；回覆指出缺的是指數價格那一格 | tool_catalog_contract.go:32-36（index* 必填）；api_tool_domain.go:114-117；api_tool_service.go:76-83 | tool_catalog_contract_test.go:138 `TestAContractCandleMissingALineNeverLeavesTheConnector`（create / indexOpen） | asserts-oracle（錯誤含 `indexOpen`；BuildRequest 失敗即不產生請求，服務層在 Send 前返回） | produces-oracle | ✅ conforms |
| AC-07 | US-02 成交筆數填零照常轉達：Given 新增合約 K 線時成交筆數填 0,其他每一項都填了 / When 送出 / Then 交易服務收到成交筆數 0 | 送出，且成交筆數為 0（不被當成沒填） | tool_catalog_contract.go:25-26；tool_arguments_domain.go:28-32（以「有無鍵」判斷，不看值） | tool_catalog_contract_test.go:130 | asserts-oracle（`"tradeCount":0`） | produces-oracle | ✅ conforms |
| AC-08 | US-02 修改時沒填溢價指數被擋下：Given 修改合約 K 線時沒填溢價指數 / When 送出 / Then 在送到交易服務之前就被擋下,並說出缺的是溢價指數那一格 | 沒送出；回覆指出缺的是溢價指數那一格 | tool_catalog_contract.go:94-103（共用 `contractKCandleFigureParameters`，premiumIndex* 必填 :37-41） | tool_catalog_contract_test.go:144（update / premiumIndexClose） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-09 | US-03 查資金費率結算：Given BTCUSDT 兩天內有六次結算 / When 助理指定 BTCUSDT 與這兩天查資金費率結算 / Then 問到合約那一條線的資金費率結算,帶著合約標的與起訖時間 | 問的是合約的資金費率結算，帶 BTCUSDT 與兩天的起訖 | tool_catalog_contract.go:169-179 | tool_catalog_contract_test.go:162 `TestTheContractSeriesAreAskedForOneContractOverOneStretch` | asserts-oracle（位址全等、symbol/startTime/endTime 值全等） | produces-oracle | ✅ conforms |
| AC-10 | US-03 查持倉統計：Given BTCUSDT 一小時內有十二筆持倉統計 / When 助理指定 BTCUSDT 與這一小時查持倉統計 / Then 問到合約那一條線的持倉統計,帶著合約標的與起訖時間 | 問的是合約的持倉統計，帶 BTCUSDT 與這一小時的起訖 | tool_catalog_contract.go:180-191 | tool_catalog_contract_test.go:162（position_statistics 子測試） | asserts-oracle | produces-oracle | ✅ conforms |
| AC-11 | US-03 沒設定帳戶金鑰時分級是空的,說明早已告知：Given 交易服務沒有設定帳戶金鑰 / When 助理讀查維持保證金分級的說明 / Then 說明寫著沒有帳戶金鑰時回空的,而且那不是錯誤 | 說明同時寫出三件事：沒有帳戶金鑰時、回來是空的、那不是錯誤 | tool_catalog_contract.go:198-199（「沒設定時回空陣列，那不是錯誤」） | tool_catalog_contract_test.go:225 | shallow — 只斷言含「帳戶金鑰」與「不是錯誤」，未釘「回空」；說明若改成「沒有帳戶金鑰時會拒絕，那不是錯誤」仍會綠 | produces-oracle | 🟠 mis-asserted |
| AC-12 | US-03 沒指定合約標的被擋下：Given 查資金費率結算、持倉統計或維持保證金分級時沒指定合約標的 / When 送出 / Then 在送到交易服務之前就被擋下,並說出缺的是合約標的那一格 | 三件都沒送出；回覆指出缺的是合約標的 | tool_catalog_contract.go:49, 201（symbol 必填）；api_tool_domain.go:114-117 | tool_catalog_contract_test.go:186-197 | asserts-oracle（三件各自刪 symbol 後錯誤含 `symbol`） | produces-oracle | ✅ conforms |
| AC-13 | US-04 資金費率說明寫出誰付給誰：When 助理讀查資金費率結算的說明 / Then 說明寫著費率為正時做多的人付給做空的人 / And 寫著結算時間不要自行取整 | 說明含「費率為正時做多付給做空」與「結算時間不要自行取整」 | tool_catalog_contract.go:173, 175 | tool_catalog_contract_test.go:220-221 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-14 | US-04 持倉統計說明寫出只有三十天：When 助理讀查持倉統計的說明 / Then 說明寫著來源只留最近三十天 | 說明含「來源只留最近三十天」 | tool_catalog_contract.go:187 | tool_catalog_contract_test.go:223 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-15 | US-04 合約 K 線說明寫出空的不是零：When 助理讀查合約 K 線的說明 / Then 說明寫著指數價格與溢價指數空的那幾根是舊資料,並指出同步那一段就會補上 | 說明寫出指數/溢價指數空的是舊資料（不是零），並指出用同步那一段補上 | tool_catalog_contract.go:81-82 | tool_catalog_contract_test.go:227-228 | asserts-oracle（「舊資料」＋指名 `trading_sync_contract_k_candle_history`） | produces-oracle | ✅ conforms |
| AC-16 | US-04 合約標的清單說明寫出只是最小那一級：When 助理讀列出合約標的的說明 / Then 說明寫著交易規格裡的維持保證金率只是最小那一級,並指出完整分級要另外查 | 說明寫出規格內的維持保證金率只是最小一級，並指出完整分級另查 | tool_catalog_contract.go:146-147 | tool_catalog_contract_test.go:230-231 | asserts-oracle（「最小那一級」＋指名 `trading_get_contract_maintenance_margin_tiers`） | produces-oracle | ✅ conforms |
| AC-17 | US-04 加入合約追蹤名單說明寫出要等與代號不對應：When 助理讀加入合約追蹤名單的說明 / Then 說明寫著要等二十秒左右 / And 寫著現貨 SHIBUSDT 在合約叫 1000SHIBUSDT | 說明（含欄位說明）寫出要等約二十秒，並寫出 SHIBUSDT ↔ 1000SHIBUSDT | tool_catalog_contract.go:156, 160 | tool_catalog_contract_test.go:233 | asserts-oracle（「二十秒」＋「1000SHIBUSDT」） | produces-oracle | ✅ conforms |
| AC-18 | US-04 合約同步進度說明寫出編號自己一串：When 助理讀看合約同步進度的說明 / Then 說明寫著合約的同步輪次編號是自己那一串 | 說明寫出合約同步輪次編號是自己那一串 | tool_catalog_contract.go:134 | tool_catalog_contract_test.go:235 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-19 | US-05 加入合約追蹤名單：Given 使用者要追 BTCUSDT 永續 / When 助理加入合約追蹤名單 / Then 問到合約那一條線的追蹤名單,帶著 BTCUSDT | 問的是合約追蹤名單，帶 BTCUSDT | tool_catalog_contract.go:150-161 | tool_catalog_contract_test.go:202 `TestJoiningTheContractWatchlistNamesTheContract` | asserts-oracle（位址全等、內文 `{"symbol":"BTCUSDT"}`） | produces-oracle | ✅ conforms |
| AC-20 | US-05 移出只停止追蹤：When 助理讀移出合約追蹤名單的說明 / Then 說明寫著只停止追蹤、已存下的一筆都不刪、現貨不受影響 | 說明寫出三件事：只停止追蹤、已存下的一筆都不刪、現貨不受影響 | tool_catalog_contract.go:164-165 | tool_catalog_contract_test.go:237 | shallow — 斷言「只停止追蹤」「不刪」「現貨」；「現貨」一詞不釘「不受影響」（說明若寫「現貨那邊也會一併移出」仍會綠） | produces-oracle | 🟠 mis-asserted |
| AC-21 | US-06 能力清單剛好多了這十四件：Given 這一刀之前的五十一件能力 / When 看能力清單 / Then 清單是那五十一件加上合約的十四件,不多也不少,而且每一件要不要身分與原本相同 | 清單 = 原 51 件 ＋ 14 件合約，恰好 65 件；每一件身分需求與原本相同 | tool_catalog.go:176-191（只多 :182 一行；`git diff main` 顯示現貨清單檔零改動） | tool_catalog_test.go:26-100, 102 `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse` | asserts-oracle（字面清單 62 件 ＋ 各自 requiresSignIn 全等比對；外掛自有的登入/登出/續用 3 件不在此清單，但本分支未動） | produces-oracle | ✅ conforms |
| AC-22 | US-06 合約能力都不需要身分：When 看合約的十四件能力 / Then 沒有一件需要身分 | 十四件全部不需要身分 | tool_catalog_contract.go（14 處 `requiresSignIn=false`） | tool_catalog_test.go:49-62 | asserts-oracle（14 件皆 `false`） | produces-oracle | ✅ conforms |
| BR-01 | 外掛只轉達、不把關：必填欄位沒填在送出前擋下，其餘一切合法性由交易服務判斷，拒絕的原話原樣帶回。 | 缺必填 → 不送出並指出那一格；其餘值（負的溢價、零筆成交…）照送；交易服務的拒絕一字不改帶回 | api_tool_domain.go:111-147（只檢查必填）；api_tool_service.go:76-99；合約清單無任何額外驗證 | api_tool_application_test.go:256 `TestARefusalComesBackInTheTradingServicesOwnWords`、:290 `TestAHalfFilledFormNamesTheMissingBoxAndIsNotSent`；tool_catalog_contract_test.go:123（負值、零照送） | asserts-oracle（共用路徑；合約能力走同一條，無另立分支） | produces-oracle | ✅ conforms |
| BR-02 | 每一件能力的說明寫給助理讀：做什麼，以及**什麼會讓它被拒絕、回來的東西怎麼讀**。 | 每一件合約能力的說明都寫出：做什麼、什麼會讓它被拒絕、回來的東西怎麼讀 | tool_catalog_contract.go:170-176（資金費率結算）、:181-188（持倉統計） | 無針對「會被拒絕」的斷言；tool_catalog_test.go:122 只檢查說明非空 | shallow | **diverges** — 交易服務對這兩件查詢會以「時間區間過大，請縮小區間（單次最多 N 筆）」及「結束時間不得早於開始時間」拒絕（go-trading `internal/domain/service/contract_funding_rate_service.go:109-131`、`contract_position_statistic_service.go:107-124`、`internal/domain/models/domains/k_candle_query_domain.go:28`），但兩件說明都沒寫任何拒絕條件；同型的 `trading_list_contract_k_candles` 則有寫「單次筆數有上限，超過即拒絕」(:83)。次要：`trading_delete_contract_k_candle`(:106) 未說不存在時的拒絕（交易服務回 404，`k_candle_contract_controller.go:194-195`）、`trading_get_contract_maintenance_margin_tiers`(:193-199) 未說代號不合法會被拒絕 | 🔴 violation |

## Orphans（沒有條款對應的程式碼）

已對照 Out of Scope 負面清單：

- 合約的重演 — 清單中**不存在**（`contractApiTools()` 無 backtest 類能力；`tool_catalog.go:88-92` 註解明言不加）。
- 交易服務內建的行情對話助手 — **不存在**（`TestNoAbilityLetsOneAssistantSpendAnothersBudget` 仍守著）。
- 現貨能力行為改變 — `git diff main` 顯示 `tool_catalog_market.go`、`tool_catalog_strategy.go`、`tool_catalog_automation.go`、`internal/` 零改動。

| Code | Description | Verdict |
|------|-------------|---------|
| — | 未發現超出條款的行為。`get`/`delete`/`backfill`/`sync`/`get_history_sync` 五件沒有專屬 AC，但由 AC-03/AC-04/AC-21/AC-22 及 BR-01/BR-02 涵蓋，不算孤兒 | — |

## Summary

- Conforms: 19/24 clauses ✅（79%）
- Violations: BR-02（資金費率結算、持倉統計兩件說明沒寫出交易服務會拒絕的條件：區間筆數上限、起訖顛倒）
- Mis-asserted: AC-01、AC-02、AC-11、AC-20（程式碼正確，但測試斷言弱於 oracle）
- Partial: —
- Gaps: —
- Unclear: —
- Orphans: 0
- 位址／動詞對照交易服務：14/14 一致，無違規。
