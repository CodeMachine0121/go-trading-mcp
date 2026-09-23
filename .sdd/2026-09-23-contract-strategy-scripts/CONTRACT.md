# 契約追溯矩陣 — contract-strategy-scripts（助理寫得出、算得動吃合約行情的策略腳本）

Contract: PRD.md（Section 3 Gherkin → `AC-xx`；Section 4 → `BR-xx`；Section 6 為 N/A）
Design map: ARCH.md（Section 7 Traceability）
Implementation: `cmd/server/tool_catalog.go`、`cmd/server/tool_catalog_strategy.go`、`cmd/server/tool_catalog_market.go`；轉達流程 `internal/domain/models/domains/api_tool_domain.go`（未改動）
Tests: `cmd/server/tool_catalog_contract_script_test.go`、`cmd/server/tool_catalog_contract_test.go`、`cmd/server/tool_catalog_test.go`、`cmd/server/dependencies_test.go`
Branch: `feat/contract-strategy-scripts`（交易服務對照：`go-trading` 的 `cmd/server/dependencies.go`、`internal/controller/models/strategy_script_request.go`、`internal/domain/models/domains/market_data_kind_domain.go`、`internal/controller/indicator_calculation_controller.go`）
Oracle: Acceptance Criteria + Business Rules（28 條）

> 靜態契約稽核：oracle 只由 PRD 文字推導（Phase 2 先於讀程式碼完成），再分別判斷測試是否斷言 oracle、程式碼是否產出 oracle。
> 僅執行對應到條款的單一測試做佐證，判定不依 pass/fail。不撰寫新探針、不執行自創情境。

## 橋接（Phase 3 Step 0）

- 行情種類 → 欄位 `marketDataKind`；K 線 → `kCandle`；合約行情 → `contractKCandle`（與交易服務 `vo.MarketDataKindKCandle`／`MarketDataKindContractKCandle` 拼法一致）。
- 「交易服務收到的那一支帶著／沒有行情種類」→ `BuildRequest` 產出的內文含／不含 `marketDataKind` 鍵。
- 合約指標計算 → `trading_calculate_contract_indicator`，POST `/contract-indicator-calculations`（交易服務 `dependencies.go:397` 同位址、同樣要登入）。
- 合約那一條線 → `/contract-` 開頭的位址；名字帶「合約」→ 名字含 `contract`（UL-MAP「合約能力」）。
- 需要身分 → `requiresSignIn`／請求 `CarriesIdentity`。
- 「在送到交易服務之前就被擋下，並說出缺的那一格」→ `BuildRequest` 回 `ErrRequiredArgumentMissing：<欄位名>`。
- 合約行情格的教學 → 常數 `contractKCandleScriptNote`（UL-MAP「合約行情格教學」）；欄位名對照交易服務 `vo.ContractKCandleVo`／`vo.PriceLineVo` 與腳本匯出名 `indicator.ContractKCandle`／`indicator.PriceLine`（`yaegi_contract_indicator_script_proxy.go:27-30`）。

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-01 | US-01 建立一支吃合約行情的策略腳本 | 送出的那一支帶行情種類「合約行情」 | tool_catalog.go:80-83 | contract_script_test:26（建立時說合約行情） | asserts-oracle（內文 `marketDataKind` = `"contractKCandle"`） | produces-oracle | ✅ |
| AC-02 | US-01 沒說行情種類照舊轉達 / 說明寫著不說就是 K 線 | 送出的那一支沒有行情種類；說明寫著建立時不說即 K 線 | tool_catalog.go:80-83；api_tool_domain.go（只送有填的） | :26（建立時沒說）、:71 | asserts-oracle（鍵不存在、其餘照送；說明含「建立時不給就是 kCandle」） | produces-oracle | ✅ |
| AC-03 | US-01 行情種類不是必填 / 只能是兩者之一 | 那一格非必填；說明寫著只能是 K 線或合約行情兩者之一 | tool_catalog.go:80-81 | :71 | 第一輪 **shallow**：只斷言兩個拼法都出現，未釘「只能兩者之一」——說明改成「例如 kCandle（現貨 K 線）或 contractKCandle 等」仍會綠 | produces-oracle（「二選一」） | 🟠 → ✅ |
| AC-04 | US-01 不認得的行情種類照轉 | 外掛照原樣送出「選擇權」這種值，不擋 | tool_catalog.go:80（無列舉驗證） | :26（`"options"`） | asserts-oracle | produces-oracle | ✅ |
| AC-05 | US-02 只改名字不提行情種類 | 送出的修改沒有行情種類；修改的說明寫著留白是保留 | tool_catalog.go:80；tool_catalog_strategy.go:36 | :26（修改時沒提）、:89 | asserts-oracle | produces-oracle | ✅ |
| AC-06 | US-02 照抄原本的行情種類 | 送出的修改帶「合約行情」 | 同上 | :26（修改時照抄） | asserts-oracle | produces-oracle | ✅ |
| AC-07 | US-02 要換成另一種會被拒絕，說明早已寫明 | 修改的說明寫著建立後不得更換、要換另建一支 | tool_catalog_strategy.go:36-37 | :89 | asserts-oracle | produces-oracle | ✅ |
| AC-08 | US-03 算式欄位教兩種入口 | 算式那一格寫出 K 線入口收一串 K 線、合約入口收一串合約行情格 | tool_catalog.go:66-69 | :101 | asserts-oracle（建立與修改兩件） | produces-oracle | ✅ |
| AC-09 | US-03 合約行情格的每一項都叫得出名字 | 合約指標計算的說明寫出同名的現貨各項，以及成交筆數、三組價格、資金費率、結算與否、持倉量、持倉價值、兩組多空比各自的名字 | tool_catalog_strategy.go:157-168；market.go:230 | :114、:145 | asserts-oracle（二十一個名字逐一斷言，且說明逐字含這段） | produces-oracle（與交易服務 `ContractKCandleVo` 欄位逐一相符） | ✅ |
| AC-10 | US-03 沒有值一律為零 | 說明寫著一律為零、分不出沒錄到與零，並列三種情形 | tool_catalog_strategy.go:170-171 | :114 | asserts-oracle | produces-oracle | ✅ |
| AC-11 | US-03 延續的費率不是每一格都收付 | 說明寫著每一格延續上一次結算的費率、只有這一格內有結算的才是收付 | tool_catalog_strategy.go:169 | :114 | asserts-oracle | produces-oracle | ✅ |
| AC-12 | US-03 兩處教的是同一段話 | 建立策略腳本與合約指標計算的說明含逐字相同的教學 | strategy.go:16, 39；market.go:230 | :145 | asserts-oracle（三件都含同一常數） | produces-oracle | ✅ |
| AC-13 | US-04 指名一支合約策略腳本算最近一天 | 問到合約那一條線的指標計算，帶編號 7、BTCUSDT、起訖、一小時 | market.go:217-236 | :185 | asserts-oracle（位址、動詞、內文全等） | produces-oracle | ✅ |
| AC-14 | US-04 自帶一段合約算式 | 算式、指標值種類、旋鈕、旋鈕值原樣送出 | market.go:165-186 | :204 | asserts-oracle | produces-oracle | ✅ |
| AC-15 | US-04 填的東西與現貨指標計算一模一樣 | 兩件的欄位名、必填與否、放的位置相同 | market.go:165-186, 208-211, 232-235 | :227 | asserts-oracle（逐格比對、全部在內文、查詢字串為空） | produces-oracle | ✅ |
| AC-16 | US-04 沒填合約標的被擋下 | 沒送出，指出缺合約標的 | market.go:233-234 | :259（symbol） | asserts-oracle | produces-oracle | ✅ |
| AC-17 | US-04 沒填觀察區間起點被擋下 | 沒送出，指出缺起點 | market.go:171-172 | :259（startTime） | asserts-oracle | produces-oracle | ✅ |
| AC-18 | US-04 需要身分 | 合約指標計算需要身分 | market.go:231 | :185；tool_catalog_test.go 清單（`true`） | asserts-oracle | produces-oracle | ✅ |
| AC-19 | US-04 說明寫出會被拒絕的情況 | 說明寫著：指名 K 線的會被拒絕並應改用現貨計算；湊不出最少可算根數會被拒絕並說出可用根數；照現貨收 K 線的算式算不動 | market.go:223-229 | :275 | asserts-oracle | produces-oracle（與交易服務 `indicator_calculation_controller.go` 的 404／400／422 對應一致） | ✅ |
| AC-20 | US-05 現貨指標計算不吃合約策略腳本 | 現貨計算的說明寫著指名合約的會被拒絕，並指向合約指標計算 | market.go:196-197 | :287 | asserts-oracle | produces-oracle | ✅ |
| AC-21 | US-05 重演、交易策略、機器人還不吃合約行情 | 建立策略腳本的說明寫著目前只能用在合約指標計算、那三者還不能、直接說做不到 | strategy.go:172-175, 16 | :136、:145 | asserts-oracle | produces-oracle | ✅ |
| AC-22 | US-05 讀回來看得到行情種類 | 列出、讀一支、市集三件的說明都寫著回來的帶著行情種類 | strategy.go:23, 28, 66 | :159 | 第一輪 **shallow**：只斷言出現 `marketDataKind` 這個字——說明改成「這裡不回 marketDataKind」仍會綠 | produces-oracle | 🟠 → ✅ |
| AC-23 | US-05 現貨指標計算其餘不變 | 現貨計算的位址、欄位與之前完全相同 | market.go:190-212 | :297 | 第一輪 **shallow**：只釘欄位名與必填，未釘型別——把某一格型別改掉（如 strategyScriptId 改成字串）仍會綠 | produces-oracle（`git diff main` 顯示九格名稱、型別、必填、說明與位址皆未變，只有順序與來源清單變了） | 🟠 → ✅ |
| AC-24 | US-06 能力清單剛好多了這一件 | 清單＝原 66 件＋合約指標計算，每一件身分需求不變 | tool_catalog.go:188-203；market.go:217 | tool_catalog_test.go:103；dependencies_test.go:21 | asserts-oracle（字面清單全等；線上列出數量＝清單＋3） | produces-oracle | ✅ |
| AC-25 | US-06 合約指標計算只問得到合約那一條線 | 送出時位址在合約那一條線、名字帶合約 | market.go:218, 231 | tool_catalog_contract_test.go:68、:96 | asserts-oracle（`everyContractAbility` 含它；位址表全等） | produces-oracle | ✅ |
| BR-01 | 外掛只轉達、不把關 | 缺必填不送；行情種類的合法性、不得更換、不混用全由交易服務判斷；拒絕原話帶回 | 無任何新驗證；api_tool_service.go（未改動） | :26（不認得的照送）、:259；api_tool_application_test.go（原話帶回） | asserts-oracle | produces-oracle | ✅ |
| BR-02 | 說明寫給助理：做什麼、什麼會被拒絕、回來怎麼讀 | 新增／改動的每件說明寫出這三者 | strategy.go:13-16, 33-39；market.go:196-197, 219-230；tool_catalog.go:80-83 | :71、:89、:159、:275、:287 | asserts-oracle | produces-oracle | ✅ |
| BR-03 | 合約行情格的教學只寫一份 | 三件能力引用同一段文字 | strategy.go:157 | :145 | asserts-oracle | produces-oracle | ✅ |

## Orphans（沒有條款對應的程式碼）

已對照 Out of Scope 負面清單：

- 合約的重演、交易策略、策略機器人吃合約行情 — **不存在**：`backtestParameters()`、`tradingStrategyWriteParameters()`、`strategyBotWriteParameters()` 零改動，沒有任何一件多出行情種類。
- 更換行情種類 — **不存在**：外掛沒有任何「換種類」的能力，說明反而寫明不得更換。
- 其他現貨能力行為改變 — `git diff main` 顯示 `tool_catalog_automation.go`、`tool_catalog_contract.go`、`internal/` 零改動；現貨指標計算只多一段說明（AC-20）。

| Code | Description | Verdict |
|------|-------------|---------|
| market.go:222 | 合約計算說明寫「合約沒有交易時段，所以不會有『觀察區間沒有交易』這種拒絕」 | 無對應條款，但與交易服務 PRD Edge Cases 一致，屬 BR-02「回來怎麼讀」的說明，非新行為 |
| market.go:226-227 | 說明指點「從來沒存過合約 K 線時先用 trading_backfill_contract_k_candles 補」 | 同上，屬 BR-02 的說明；指向既有能力 |

## Summary（第一輪）

- Conforms: 25/28 ✅（89%）
- Violations: —
- Mis-asserted: AC-03、AC-22、AC-23（程式碼正確，測試斷言弱於 oracle）
- Partial / Gaps / Unclear: —
- Orphans: 0 條行為孤兒（兩處說明補充已判定屬 BR-02）
- 位址、動詞、欄位名對照交易服務：一致。

---

## 第二輪（修正後）

| ID | 第一輪 | 第二輪 | 怎麼解決的 |
|----|--------|--------|------------|
| AC-03 | 🟠 | ✅ | `TestTheMarketKindBoxSaysWhatLeavingItOutMeans` 改斷言「二選一：kCandle（現貨 K 線）或 contractKCandle」 |
| AC-22 | 🟠 | ✅ | `TestEveryStrategyScriptReadSaysItCarriesTheKind` 逐件斷言完整宣稱（「每一支都帶著 marketDataKind」／「以及它吃哪一種行情（marketDataKind）」）；以「這裡不回 marketDataKind」變異驗證會轉紅 |
| AC-23 | 🟠 | ✅ | `TestTheSpotCalculationStillAsksWhereItAlwaysDid` 改為逐格比對型別＋必填；以「strategyScriptId 改成字串」變異驗證會轉紅 |

另修正 ARCH 一處敘述：共用欄位清單是**八格**，不是九格。

**第二輪結果：28/28 條款 conforms，無違規、無弱斷言、無孤兒。**
