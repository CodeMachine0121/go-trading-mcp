# Contract Verification — 助理重演得了合約帳戶

**Oracle:** PRD Acceptance Criteria · **Ceiling:** static conformance audit — test assertions and declarations are judged against each clause's spec outcome; no invented scenario was executed.

## Clauses

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
|---|---|---|---|---|---|---|---|
| AC-01 | 帶著合約的條件重演 | 合約重演收到槓桿、交易模式、滑點、標的、期間 | `tool_catalog_strategy.go` `contractBacktestApiTools` | `TestTheContractScriptReplayForwardsTheContractConditions` | asserts-oracle | produces-oracle | ✅ |
| AC-02 | 沒填的合約條件不送出 | 三格不出現在內容裡 | `ApiToolDomain.BuildRequest`（只送填了的） | `TestTheContractScriptReplaySendsNothingForWhatWasLeftBlank` | asserts-oracle | produces-oracle | ✅ |
| AC-03 | 必填的格子沒填 | 送出前擋下並說出 symbol | `backtestParameters` symbol 必填 | `TestTheContractReplaysHoldBackAMissingSymbol` | asserts-oracle | produces-oracle | ✅ |
| AC-04 | 交易服務拒絕 | 原話帶回 | 既有轉達路徑（`FailureReasonDomain`） | 既有 `internal/application/tests` 拒絕轉達測試 | asserts-oracle | produces-oracle | ✅ |
| AC-05 | 說明寫出拒絕情況與帳戶算法 | 兩類句子都讀得到 | `contractAccountReplayNote` + 兩件說明 | `TestBothContractReplaysSayHowTheContractAccountIsReplayed`、`TestTheContractReplaysSayWhatTheyRefuseAndWhereItGoes` | asserts-oracle | produces-oracle | ✅ |
| AC-06 | 指名交易策略重演 | 送到那份的合約重演，帶槓桿 | `trading_backtest_contract_trading_strategy` | `TestTheContractTradingStrategyReplayTakesNoTradingMode`、`TestEveryContractAbilityAsksTheRightWayAtTheRightAddress` | asserts-oracle | produces-oracle | ✅ |
| AC-07 | 沒有交易模式可以填 | 無該格，說明說出由交易策略決定 | 同上 | `TestTheContractTradingStrategyReplayTakesNoTradingMode` | asserts-oracle | produces-oracle | ✅ |
| AC-08 | 建立時說出行情種類與交易模式 | 兩格都送出 | `tradingStrategyWriteParameters` | `TestWritingATradingStrategySaysItsKindAndTradingMode` | asserts-oracle（tradingMode 送出、兩格宣告） | produces-oracle | ✅ |
| AC-09 | 修改時不提行情種類 | 不送出 | 同上 + `BuildRequest` | 同上（NotContains marketDataKind） | asserts-oracle | produces-oracle | ✅ |
| AC-10 | 說明寫出兩格規則 | 不得更換、同一種行情、只有合約有模式 | 同上 | 同上 | asserts-oracle | produces-oracle | ✅ |
| AC-11 | 現貨重演指向合約重演 | 讀得到改用合約重演、讀不到「目前不做」 | `spotOnlyReplayNote` | `TestBothReplaysSayTheyOnlyEverTradeSpot` | asserts-oracle | produces-oracle | ✅ |
| AC-12 | 現貨重演仍沒有合約格子 | 無交易模式、槓桿、滑點、維持保證金率 | `backtestParameters` | `TestNoAbilityOffersATradingModeToChoose`、`TestNeitherReplayTakesALeverage`、`TestTheSpotReplaysTakeNoneOfTheContractBoxes` | asserts-oracle | produces-oracle | ✅ |
| AC-13 | 合約算式說明只說機器人還不行 | 可去合約重演與交易策略；機器人只跑 K 線 | `contractKCandleScriptNote` | `TestTheContractScriptNoteSaysWhereItCanAndCannotGo` | asserts-oracle | produces-oracle | ✅ |
| AC-14 | 清單涵蓋新的兩件 | 多兩件、都需要身分 | `apiToolCatalog` | `TestTheCatalogueCoversEveryThingTheTradingServiceOffersAndNothingElse`、`TestTheAssembledConnectorOffersEveryAbilityOverTheWire` | asserts-oracle | produces-oracle | ✅ |

## Orphans

| Behavior | Location | Note |
|---|---|---|
| 機器人的 tradingStrategyId 說明寫出只能掛 K 線交易策略 | `tool_catalog_automation.go` | BRIEF「說法改對」有寫、PRD 未列成 Scenario；屬同一目的的說明，保留 |

## Summary

✅ 14 conforms · 🔴 0 · 🟠 0 · 🟡 0 · ❌ 0 · ❔ 0 · ⚠️ 1 orphan (benign, documented) — Conformance 100%
