# 助理寫得出、算得動吃合約行情的策略腳本 — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-09-23-contract-strategy-scripts/PRD.md`
**Tech context:** Go · MCP server · 能力清單在組裝根（`cmd/server/tool_catalog*.go`），每一件能力是一個 `domains.ApiToolDomain`

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 在能力清單補一件合約指標計算、在策略腳本的寫入欄位補一格 `marketDataKind`，並讓說明教得出合約算式——轉達流程一行不動。
- **Guiding principle:** **清單就是功能**（沿用上一刀）。轉達流程對每一件能力都一樣：只轉達宣告過、而且有填的欄位，
  所以「沒說行情種類」天然就是「不送這一格」，交易服務照舊的預設接手——外掛不必、也不該替它補 `kCandle`。
  **教學只寫一份**：合約行情格的說法是一個常數，兩處引用，與既有的 `spotOnlyReplayNote`、`costedReportCardNote` 同一種作法。

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| `cmd/server/tool_catalog.go` | **Modify** | `strategyScriptWriteParameters()` 多一格 `marketDataKind`（body、字串、非必填）；`script` 那一格寫出兩種入口 |
| `cmd/server/tool_catalog_strategy.go` | **Modify** | 新增常數 `contractKCandleScriptNote`；建立／修改策略腳本的說明引用它；修改的說明寫出「行情種類留白是保留、不得更換」；列出／讀一支／市集的說明寫出回來帶著 `marketDataKind` |
| `cmd/server/tool_catalog_market.go` | **Modify** | 抽出 `indicatorCalculationParameters()`（除了 `symbol` 以外現貨與合約共用的九格）；`indicatorApiTools()` 多一件 `trading_calculate_contract_indicator`；現貨那一件的說明補「指名合約策略腳本會被拒絕」 |
| `cmd/server/tool_catalog_test.go` | **Modify** | `everyAbilityTheTradingServiceOffers` 多一列（`true`） |
| `cmd/server/tool_catalog_contract_test.go` | **Modify** | `everyContractAbility` 與位址表各多一列——守衛本來就是為了這種時候 |
| `cmd/server/tool_catalog_contract_indicator_test.go` | **Add** | 這一刀專屬的測試 |
| `README.md` | **Modify** | 能力數 66 → 67；補行情種類與合約指標計算 |
| 轉達流程（`internal/`） | **Not touched** | 沒有新的行為；留白不送是既有規則 |
| 重演、交易策略、策略機器人的欄位 | **Not touched** | PRD Out of Scope；交易服務的下一刀才會有 |

## 3. New Classes / Modules

| Name | Kind | Responsibility | Satisfies |
| :--- | :--- | :--- | :--- |
| `contractKCandleScriptNote` | 說明常數 | 合約算式的入口、合約行情格的每一個欄位名（`indicator.ContractKCandle`／`indicator.PriceLine`）、缺值給零的三種情形、延續費率與 `FundingSettledInBar`，以及「目前只能用在合約指標計算，重演／交易策略／機器人直接說做不到」 | US-03, US-05 |
| `indicatorCalculationParameters()` | 共用欄位清單 | 兩種指標計算共用的九格（`strategyScriptId`、`startTime`、`endTime`、`aggregationInterval`、`script`、`resultType`、`parameters`、`parameterValues`）——**`symbol` 不在其中**，因為兩邊的代號是兩種商品，各自說明 | US-04 |
| `trading_calculate_contract_indicator` | 能力 | POST `/contract-indicator-calculations`，需要身分；說明寫出拒絕條件並引用 `contractKCandleScriptNote` | US-04, US-06 |

`marketDataKind` 那一格只有一份說明，同時寫出建立與修改的兩條留白規則（建立＝`kCandle`、修改＝保留），
因為它住在建立與修改共用的 `strategyScriptWriteParameters()`，兩邊不可能讀到不同說法。

## 4. Depth check

- `strategyScriptWriteParameters()` 仍是一份清單：多一格不增加呼叫端要知道的事。
- `indicatorCalculationParameters()` 把「兩種計算填同樣的東西」變成事實而不是要人記得的規則；它沒有參數，只把會各自不同的 `symbol` 留在外面。
- 沒有新的介面或流程分支。

## 5. Rejected alternatives

- **把合約指標計算放進 `contractApiTools()`**：那一組全部不需要身分、是行情那條線；計算需要身分、屬於指標那一組。
  名字帶 `contract`、位址 `/contract-` 開頭，所以仍被 `everyContractAbility` 的守衛管住。
- **外掛替留白補上 `kCandle`**：修改時留白是「保留」，補了會讓一支合約策略腳本的改名被拒絕。
- **在 `marketDataKind` 加 enum 驗證**：違反「外掛只轉達、不把關」。

## 6. Extensibility & Handoff Notes

- **下一個需求**：合約重演（以及交易策略、機器人吃合約行情）。那是**另一件能力、自己一份欄位清單**（`backtestParameters` 的註解早已寫明），
  並把 `contractKCandleScriptNote` 裡「重演／交易策略／機器人還不能用」那一句改掉——**一處改，三處說法一起變**。
- **第三種行情種類**：只要改 `marketDataKind` 那一格的說明與（若有）一份新的教學常數；新增一件對應的計算能力時重用 `indicatorCalculationParameters()`。

## 7. Traceability

| PRD | Fulfilled by | Test |
| :--- | :--- | :--- |
| US-01 全部 | `strategyScriptWriteParameters()` 的 `marketDataKind` ＋ 既有「只送有填的」 | `TestWritingAStrategyScriptForwardsTheMarketKindOnlyWhenGiven`、`TestTheMarketKindBoxSaysWhatLeavingItOutMeans` |
| US-02 全部 | 同上 ＋ 修改的說明 | 同上 ＋ `TestRewritingAStrategyScriptSaysTheKindIsKept` |
| US-03 全部 | `contractKCandleScriptNote` ＋ `script` 那一格 | `TestTheContractScriptShapeIsTaughtOnceInBothPlaces`、`TestTheContractScriptNoteNamesEveryFigure` |
| US-04 全部 | `trading_calculate_contract_indicator` ＋ `indicatorCalculationParameters()` | `TestTheContractIndicatorCalculationAsksTheContractLine`、`TestBothIndicatorCalculationsTakeTheSameBoxes`、`TestTheContractIndicatorCalculationStopsAMissingBox`、`TestTheContractIndicatorCalculationSaysWhatWillGetItRefused` |
| US-05 全部 | 現貨計算說明、`contractKCandleScriptNote`、讀取三件的說明 | `TestTheSpotCalculationSaysAContractScriptIsRefused`、`TestEveryStrategyScriptReadSaysItCarriesTheKind`、`TestTheSpotCalculationStillAsksWhereItAlwaysDid` |
| US-06 全部 | `apiToolCatalog` | `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse`、`TestContractAbilitiesOnlyEverReachTheContractLine`、`TestEveryContractAbilityAsksTheRightWayAtTheRightAddress` |

## 8. Risks & Open Decisions

- 交易服務不擋把合約策略腳本掛上重演／交易策略／機器人，只會在執行時失敗；外掛這邊只有說明這一道防線。
