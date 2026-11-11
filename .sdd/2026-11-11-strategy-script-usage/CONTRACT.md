# Contract Traceability Matrix — 策略腳本的用途（外掛）

Contract: PRD.md · Design map: ARCH.md · Implementation: `cmd/server/`（`internal/` 未動）
Oracle: Acceptance Criteria (18 scenarios) + Business Rules (5) + NFR (2) = 25 clauses

Path abbreviations: `strat` = `cmd/server/tool_catalog_strategy.go`, `catalog` = `cmd/server/tool_catalog.go`, `market` = `cmd/server/tool_catalog_market.go`,
`screening` = `cmd/server/tool_catalog_symbol_screening.go`, `UT` = `cmd/server/tool_catalog_strategy_script_usage_test.go`,
`AT` = `internal/application/tests/api_tool_application_test.go`（拒絕原話轉交）。行號以稽核當下的 commit 為準。

## Clauses

`Test audit` / `Code audit` 是 Claude 的 shadow verdict（在 Jev 之前記下、之後不改）；`Jev`、`Status`、`Route` 照 `jev_audit.py` 原樣抄錄。

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | 建立圖表用腳本 | 送出內容帶用途「圖表用」 | strat:12-26；catalog:96-104 | UT CreatingAStrategyScriptForwardsTheUsageOnlyWhenGiven | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.84 | ✅ conforms | recheck (mid confidence) |
| AC-2 | 建立時沒說用途 | 送出內容沒有用途；說明說這時是計算用 | 同上 | UT Forwards…（沒說）；UT CreatingAStrategyScriptSaysWhatTheUsageIs | asserts-oracle | produces-oracle | asserts-oracle 0.80 | produces-oracle 0.63 | ✅ conforms | recheck (mid confidence) |
| AC-3 | 用途不在兩種之內 | 拒絕原話帶回；說明預先寫只能圖表用或計算用 | catalog:96-104 | UT Forwards…（backtest）；UT EveryUsageBoxSpellsTheSameTwoUsages；AT RefusalComesBack… | asserts-oracle | produces-oracle | asserts-oracle 0.36 | unclear 0.07 | ❔ unclear | escalate (disagree) |
| AC-4 | 改寫內容不帶用途 | 改寫能力沒有用途格、送出內容沒有用途；說明指向切換用途 | strat:40-56 | UT RewritingAStrategyScriptNeverCarriesAUsage | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.68 | ✅ conforms | recheck (mid confidence) |
| AC-5 | 沒人在用的腳本切成圖表用 | 送出切換、內容只有用途圖表用；帶回切換後的腳本 | strat:90-104 | UT SwitchingAStrategyScriptsUsageSendsOnlyTheUsage；UT …SaysWhenItIsRefused… | asserts-oracle | produces-oracle | shallow 0.27 | unclear 0.48 | ❔ unclear | escalate (disagree) |
| AC-6 | 交易策略在用時不能切成圖表用 | 原話帶回說出「4H 趨勢」；說明請使用者先換掉、不要重送 | strat:90-99；既有拒絕轉交 | UT …SaysWhenItIsRefused…；AT RefusalComesBack… | asserts-oracle | produces-oracle | asserts-oracle 0.63 | unclear 0.35 | ❔ unclear | escalate (disagree) |
| AC-7 | 篩選設定在用時不能切成圖表用 | 原話帶回說出「新幣」 | 同上 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.37 | unclear 0.25 | ❔ unclear | escalate (disagree) |
| AC-8 | 圖表用切回計算用 | 說明說切成計算用沒有限制 | strat:90-99 | UT …SaysWhenItIsRefused… | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.55 | ✅ conforms | escalate (low confidence) |
| AC-9 | 切成原本的用途 | 說明說不算失敗、什麼也不變 | strat:90-99 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | unclear 0.48 | ❔ unclear | escalate (disagree) |
| AC-10 | 市集副本也能切換 | 說明說副本也能切、算式仍看不到改不了 | strat:90-99 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.77 | ✅ conforms | recheck (mid confidence) |
| AC-11 | 別人的腳本 | 說明說被拒、說法同找不到 | strat:90-99 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.89 | unclear 0.24 | ❔ unclear | escalate (disagree) |
| AC-12 | 圖表用的腳本當信號來源 | 交易策略寫入說明說只有計算用能當信號來源、圖表用被拒、可先切 | strat:208-211 | UT BothTradingStrategyWritesSayOnlyCalculationScripts… | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.83 | ✅ conforms | recheck (mid confidence) |
| AC-13 | 篩選設定指名圖表用的腳本 | 篩選設定說明說只接受計算用 | screening:29-32 | UT BothScreeningProfileWritesSay… | asserts-oracle | produces-oracle | asserts-oracle 0.76 | unclear 0.15 | ❔ unclear | escalate (disagree) |
| AC-14 | 直接回測圖表用的腳本 | 兩件直接回測說明說指名的須為計算用、自帶算式不受影響 | strat:213-216、380、473 | UT BothScriptReplaysSay… | asserts-oracle | produces-oracle | asserts-oracle 0.88 | produces-oracle 0.64 | ✅ conforms | recheck |
| AC-15 | 試算不受用途影響 | 兩件算指標說明說兩種用途都能算 | market:172-173 | UT BothIndicatorCalculationsSay… | asserts-oracle | produces-oracle | asserts-oracle 0.91 | produces-oracle 0.81 | ✅ conforms | recheck (mid confidence) |
| AC-16 | 查看與列出腳本 | 說明說每一支帶著用途 | strat:26-40、200-206 | UT EveryAbilityAboutAScriptsUsage… | asserts-oracle | produces-oracle | asserts-oracle 0.81 | produces-oracle 0.68 | ✅ conforms | recheck (mid confidence) |
| AC-17 | 市集帶著用途 | 瀏覽說帶用途；加入說副本帶加入當下用途、之後不受發佈者影響 | strat:72-90 | UT TheMarketplaceSays… | asserts-oracle | produces-oracle | asserts-oracle 0.80 | unclear 0.20 | ❔ unclear | escalate (disagree) |
| AC-18 | 圖表上找不到我的指標 | 說明說圖表選單只列圖表用、升級前都是計算用、徵得同意可切 | strat:200-206 | UT EveryAbilityAboutAScriptsUsage… | asserts-oracle | produces-oracle | asserts-oracle 0.63 | produces-oracle 0.77 | ✅ conforms | recheck (mid confidence) |
| BR-1 | 用途二選一；建立不說即計算用；改寫不能帶 | 建立格子選填、改寫無格子 | catalog:96-104；strat:12-56 | UT Creating…、Rewriting… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-2 | 切換用途的規則 | 切換說明含副本、被引用拒絕、計算用無限制、同用途不變、別人同找不到 | strat:90-104 | UT Switching… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-3 | 只接受計算用的三處；自帶算式、試算不受影響 | 三處說明＋兩個例外 | strat:208-216；screening:29-32；market:172 | UT BothTradingStrategyWrites…、BothScriptReplays…、BothScreening…、BothIndicator… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-4 | 圖表選單只列圖表用；既有腳本為計算用 | 說明含兩者 | strat:200-206 | UT EveryAbilityAboutAScriptsUsage… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-5 | 外掛原樣轉送、拒絕原話帶回 | body 原樣 | `EncodedSubset`；既有轉交路徑 | UT Forwards…；AT RefusalComesBack… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| NFR-1 | 改寫能力不得帶用途 | 改寫不宣告、送不出 | strat:40-56 | UT RewritingAStrategyScriptNeverCarriesAUsage | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| NFR-2 | 以使用者授權給外掛的帳號進行 | 新能力走同一條授權路徑 | 單一路徑（`internal/` 未動） | 既有 `TestEveryAbilityTravelsUnderTheConnectorAuthorizationTheCallerBrought` | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |

### Recheck / escalate 的複查結果

- **AC-3、5、6、7、9、11、13、17（Jev：code unclear 0.07–0.48）**：重讀 strat:72-104、catalog:96-104、screening:29-32，句子與格子都在且被 UT 逐字斷言；
  切換的路徑、動詞與內文由 `TestSwitchingAStrategyScriptsUsageSendsOnlyTheUsage` 以 `JSONEq` 斷言只有 `usage`。Jev 的低信心來自「拒絕是交易服務的行為、外掛只轉送與預先說明」。
  所有句子與格子都經 `/implement` 的突變驗證。**AC-5 的 Jev test shallow 0.27** 不成立：路徑、動詞、內文三者皆被斷言，「帶回切換後的腳本」由說明斷言（回覆本身走共用的原樣轉交）。不改。
- **AC-14**：第一次跑時給錯了程式行號（回測格子在 strat:380、473），改正後 Jev 為 conforms。

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| create 說明「畫在圖上看的建立成 chart」 | 屬 US-01 建立圖表用腳本的引導句 | 屬 AC-1 |
| README〈策略腳本的用途〉 | 文件 | 非行為 |

無任何程式碼落在 Out of Scope（外掛不檢查用途、不推測用途、不碰圖表選單）。

## Summary

- Shadow verdict: 25/25 conforms.
- Jev status (as returned, 18 AC): ✅ 10 conforms · ❔ 8 unclear
- Violations: none · Mis-asserted: none
- Orphans: 2（皆非行為）
- Jev: ran on 18 AC clauses; BR/NFR not run（與 AC 重疊）
- Needs James: AC-3、5、6、7、8、9、11、13、17（route escalate，複查後判定 conforms）

> Static conformance audit against the Acceptance Criteria — it judges test assertions and code paths against the spec's expected outcome, not by running the full suite.
