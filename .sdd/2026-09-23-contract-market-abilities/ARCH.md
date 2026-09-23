# 能力清單接上永續合約那一條線 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-09-23-contract-market-abilities/PRD.md`
**Tech context:** Go · MCP server · 能力清單在組裝根（`cmd/server/tool_catalog*.go`），每一件能力是一個 `domains.ApiToolDomain`

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 在能力清單補上十五件合約能力，每一件只問得到交易服務的合約位址，且沒有一件現貨能力的行為改變。
- **Guiding principle:** **清單就是功能。** 既有的轉達流程對每一件能力都一樣（依欄位位置組請求、必填沒填就擋、原話帶回拒絕），
  新增一件事只是在清單補一列——不加 handler、不加分支。所以這一刀**只新增一個清單檔**，並把它掛進 `apiToolCatalog`。

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog_contract.go` | **Add** | `contractApiTools()`：十五件合約能力，與兩組共用的欄位清單 |
| `cmd/server/tool_catalog.go` | **Modify** | `apiToolCatalog` 多掛一行 `contractApiTools()` |
| `cmd/server/tool_catalog_test.go` | **Modify** | `everyAbilityTheTradingServiceOffers` 多十四列——這份清單本來就是用來在服務新增能力時轉紅的 |
| `cmd/server/tool_catalog_contract_test.go` | **Add** | 合約能力專屬的守衛測試 |
| `README.md` | **Modify** | 能力數 51 → 66，補一段合約的說明 |
| 轉達流程（`internal/`） | **Not touched** | 合約能力與現貨能力走同一條路；沒有新的行為要加 |
| 現貨能力 | **Not touched** | PRD US-06 |

## 3. New Classes / Modules

| Name | Kind | Responsibility | Satisfies |
| :--- | :--- | :--- | :--- |
| `contractApiTools()` | 清單函式 | 十五件合約能力，名字一律帶 `contract`、路徑一律 `/contract-` 開頭、全部 `requiresSignIn=false` | US-01, US-03, US-05, US-06 |
| `contractKCandleFigureParameters()` | 共用欄位清單 | 一根合約 K 線的二十一個數字欄位，**全部必填**；新增與修改共用，免得一份漏了另一份有的欄位 | US-02 |
| `contractRangeParameters()` | 共用欄位清單 | 合約每一種資料的「合約標的＋起訖」三格，查合約 K 線、資金費率結算、持倉統計共用 | US-01, US-03 |

十五件能力：

| 能力 | 動作 | 位址 |
| :--- | :--- | :--- |
| `trading_create_contract_k_candle` | 提交 | `/contract-k-candles` |
| `trading_list_contract_k_candles` | 讀 | `/contract-k-candles` |
| `trading_get_contract_k_candle_series` | 讀 | `/contract-k-candles/series` |
| `trading_get_contract_k_candle` | 讀 | `/contract-k-candles/{symbol}/{openTime}` |
| `trading_update_contract_k_candle` | 取代 | `/contract-k-candles/{symbol}/{openTime}` |
| `trading_delete_contract_k_candle` | 移除 | `/contract-k-candles/{symbol}/{openTime}` |
| `trading_backfill_contract_k_candles` | 提交 | `/contract-k-candles/backfill` |
| `trading_sync_contract_k_candle_history` | 提交 | `/contract-k-candles/history` |
| `trading_get_contract_k_candle_history_sync` | 讀 | `/contract-k-candles/history/{id}` |
| `trading_list_contract_trading_symbols` | 讀 | `/contract-trading-symbols` |
| `trading_add_to_contract_watchlist` | 提交 | `/contract-watchlist` |
| `trading_remove_from_contract_watchlist` | 移除 | `/contract-watchlist/{symbol}` |
| `trading_list_contract_funding_rate_settlements` | 讀 | `/contract-funding-rate-settlements` |
| `trading_list_contract_position_statistics` | 讀 | `/contract-position-statistics` |
| `trading_get_contract_maintenance_margin_tiers` | 讀 | `/contract-maintenance-margin-tiers` |

## 6. Extensibility & Handoff Notes

- **下一個需求**：合約重演。它是**另一件能力、自己一份欄位清單**，不是在現貨重演上加一個開關（`backtestParameters` 的註解早已寫明）。放進 `contractApiTools()` 旁邊即可。
- **合約新增一種資料**：在 `contractApiTools()` 補一列；若是「標的＋區間」查詢就重用 `contractRangeParameters()`。
- **守衛**：`TestContractAbilitiesOnlyEverReachTheContractLine` 同時檢查名字與位址——新增合約能力卻忘了名字帶 `contract`，或新增現貨能力卻指到合約位址，都會轉紅。

## 7. Traceability

| PRD | Fulfilled by | Test |
| :--- | :--- | :--- |
| US-01 全部 | `contractApiTools()` 的名字與位址 | `TestContractAbilitiesOnlyEverReachTheContractLine` |
| US-02 全部 | `contractKCandleFigureParameters()` 全必填 ＋ 既有轉達流程的必填檢查 | `TestWritingAContractCandleAsksForEveryFigure`、`TestAWrittenContractCandleCarriesItsPriceLines`、`TestAContractCandleMissingALineNeverLeavesTheConnector` |
| US-03 全部 | 三件查詢能力 ＋ `contractRangeParameters()` | `TestTheContractSeriesAreAskedForOneContractOverOneStretch` |
| US-04 全部 | 各能力說明 | `TestTheContractAbilitiesSayWhatCannotBeDiscoveredBySending` |
| US-05 全部 | 加入／移出合約追蹤名單 | 同上 ＋ 位址測試 |
| US-06 全部 | `apiToolCatalog` | `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse` |

## 8. Risks & Open Decisions

- 資金費率結算、持倉統計、維持保證金分級要交易服務 PR #64 部署後才有東西可問；在那之前這三件會收到交易服務的「找不到」。
