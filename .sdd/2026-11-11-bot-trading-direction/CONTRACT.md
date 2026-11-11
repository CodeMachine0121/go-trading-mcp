# Contract Traceability Matrix — 交易方向從交易策略搬到機器人與回測（外掛）

Contract: PRD.md · Design map: ARCH.md · Implementation: `cmd/server/`（`internal/` 未動）
Oracle: Acceptance Criteria (28 scenarios) + Business Rules (6) + NFR (2) = 36 clauses

Path abbreviations: `strat` = `cmd/server/tool_catalog_strategy.go`, `auto` = `cmd/server/tool_catalog_automation.go`,
`binance` = `cmd/server/tool_catalog_binance_trading_key.go`, `screening` = `cmd/server/tool_catalog_symbol_screening.go`,
`BT` = `cmd/server/tool_catalog_bot_trading_direction_test.go`, `CB` = `cmd/server/tool_catalog_contract_backtest_test.go`, `T` = `cmd/server/tool_catalog_test.go`,
`AT` = `internal/application/tests/api_tool_application_test.go`（拒絕原話轉交）, `TA` = `internal/domain/models/domains/tool_arguments_domain.go:60`（`EncodedSubset`：只送宣告過且填了的名字）。
行號以稽核當下的 commit 為準。

## Clauses

`Test audit` / `Code audit` 是 Claude 的 shadow verdict（在 Jev 之前記下、之後不改）；`Jev`、`Status`、`Route` 照 `jev_audit.py` 原樣抄錄。

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev test | Jev code | Status | Route |
|----|--------|------------------------|------|------|------------|------------|----------|----------|--------|-------|
| AC-1 | 建立合約交易策略不提交易方向 | 送出內容沒有交易模式；建立能力沒有交易模式這一項 | strat:114-142 | CB WritingATradingStrategySaysItsKindAndOffersNoTradingMode | asserts-oracle | produces-oracle | asserts-oracle 0.93 | produces-oracle 0.89 | ✅ conforms | accept |
| AC-2 | 使用者要求交易策略只做多 | 說明告訴助理方向改由機器人交易方向或回測交易模式決定 | strat:191-199 | BT EveryTradingStrategyAbilitySays… | asserts-oracle | produces-oracle | asserts-oracle 0.75 | produces-oracle 0.85 | ✅ conforms | recheck (mid confidence) |
| AC-3 | 修改時仍送交易模式被拒 | 拒絕原話帶回；修改說明不再寫保留交易模式 | strat:170-180；格子已刪（送不出） | CB RewritingATradingStrategySaysOnlyTheKindIsKept；AT RefusalComesBack… | asserts-oracle | produces-oracle | asserts-oracle 0.89 | produces-oracle 0.29 | ✅ conforms | escalate (low confidence) |
| AC-4 | 讀交易策略看不到交易模式 | 說明不說帶著交易模式；說同一份可養做多與做空的機器人 | strat:160-170、191-199 | BT EveryTradingStrategyAbilitySays… | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.80 | ✅ conforms | recheck (mid confidence) |
| AC-5 | 建立時選做多 | 送出內容帶交易方向「做多」 | auto:105-108；TA | BT ForwardsItsDirectionAndNotificationsAsGiven | asserts-oracle | produces-oracle | asserts-oracle 0.63 | produces-oracle 0.85 | ✅ conforms | recheck (mid confidence) |
| AC-6 | 建立時沒說方向 | 不送交易方向；說明說這時是做多加做空 | auto:105-108 | BT …WithoutDirectionOrNotificationsSendsNeither；BT DeclaresItsDirection… | asserts-oracle | produces-oracle | asserts-oracle 0.57 | unclear 0.35 | ❔ unclear | escalate (disagree) |
| AC-7 | 一個方向都沒選 | 空清單原樣送出、拒絕原話帶回；說明預先寫至少一個 | auto:105-108 | BT Forwards…（[]）；BT Declares… | asserts-oracle | produces-oracle | shallow 0.16 | unclear 0.19 | ❔ unclear | escalate (disagree) |
| AC-8 | 現貨機器人送來方向 | 說明預先寫現貨機器人沒有交易方向 | auto:105-108 | BT Declares… | asserts-oracle | produces-oracle | asserts-oracle 0.77 | unclear 0.09 | ❔ unclear | escalate (disagree) |
| AC-9 | 同一份交易策略兩台方向不同 | 說明讓助理說 A 平多、B 做空並列三種對照 | auto:177-185 | BT SaysWhatItsDirectionMeans… | asserts-oracle | produces-oracle | asserts-oracle 0.85 | produces-oracle 0.91 | ✅ conforms | accept |
| AC-10 | 只改名字時方向保留 | 修改說明說沒送方向即保留 | auto:260-263 | BT RewritingAStrategyBotSaysTheDirection…Kept | asserts-oracle | produces-oracle | asserts-oracle 0.79 | unclear 0.40 | ❔ unclear | escalate (disagree) |
| AC-11 | 讀機器人看得到方向 | 七件回覆機器人的能力說合約帶交易方向、現貨沒有 | auto:169-173 | BT EveryAbilityAnsweringWithABot… | asserts-oracle | produces-oracle | asserts-oracle 0.89 | produces-oracle 0.85 | ✅ conforms | accept |
| AC-12 | 安全時改方向 | 修改送出內容帶交易方向「做空」 | auto:105-108；TA | BT Forwards…（update × ["short"]） | asserts-oracle | produces-oracle | asserts-oracle 0.60 | produces-oracle 0.73 | ✅ conforms | recheck (mid confidence) |
| AC-13 | 自動下單開著時改方向 | 說明說外掛關不了自動下單，請使用者到網頁詳細頁先關、不要重送 | binance:18-24；auto:185 | BT SaysWhatItsDirectionMeans…；BT TheInstructionsSay… | asserts-oracle | produces-oracle | asserts-oracle 0.91 | produces-oracle 0.73 | ✅ conforms | recheck (mid confidence) |
| AC-14 | 監聽標的還有機器人持倉 | 說明說照原話說出那個標的還有持倉，要先讓它空手 | auto:181-183 | BT SaysWhatItsDirectionMeans… | shallow | produces-oracle | asserts-oracle 0.45 | produces-oracle 0.81 | ✅ conforms | escalate (disagree) |
| AC-15 | 監聽標的還有沒做完的自動下單 | 說明說要等它做完再改 | auto:181-183 | BT SaysWhatItsDirectionMeans… | shallow | produces-oracle | asserts-oracle 0.27 | produces-oracle 0.81 | ✅ conforms | escalate (disagree) |
| AC-16 | 方向不變只改間隔 | 說明說方向沒變不受自動下單限制 | auto:184 | BT SaysWhatItsDirectionMeans… | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.72 | ✅ conforms | recheck (mid confidence) |
| AC-17 | 執行中改方向或通知 | 說明照舊寫只有已停止的機器人改得動 | auto:255-262 | BT RewritingAStrategyBot…Kept | asserts-oracle | produces-oracle | asserts-oracle 0.97 | produces-oracle 0.52 | ✅ conforms | escalate (low confidence) |
| AC-18 | 關掉買入通知 | 送出通知只帶買入關；說明說沒說的維持原本 | auto:109-112；TA | BT Forwards…（{"buy":false}）；BT Declares… | asserts-oracle | produces-oracle | asserts-oracle 0.41 | unclear 0.51 | ❔ unclear | escalate (disagree) |
| AC-19 | 建立時沒提通知 | 說明說這時買入開、賣出開、持有關 | auto:109-112 | BT Declares…；BT …SendsNeither | asserts-oracle | produces-oracle | asserts-oracle 0.92 | unclear 0.28 | ❔ unclear | escalate (disagree) |
| AC-20 | 關掉通知仍會下單 | 說明說只是不寄訊息，信號、去重、自動下單照舊 | auto:190-196 | BT SaysANotificationOnlySilencesTheMessage | asserts-oracle | produces-oracle | asserts-oracle 0.91 | produces-oracle 0.51 | ✅ conforms | escalate (low confidence) |
| AC-21 | 回到持有 | 說明讓助理認得回到持有那一則：沒有建議部位、不下單 | auto:190-196 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.96 | produces-oracle 0.49 | ✅ conforms | escalate (low confidence) |
| AC-22 | 連續持有 | 說明說不會再寄 | auto:190-196 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.95 | produces-oracle 0.65 | ✅ conforms | recheck (mid confidence) |
| AC-23 | 三個通知都關 | 說明說機器人照跑、輪次訊息都不寄、已啟動已停止照寄 | auto:190-196 | 同上 | asserts-oracle | produces-oracle | asserts-oracle 0.69 | produces-oracle 0.67 | ✅ conforms | recheck (mid confidence) |
| AC-24 | 讀機器人看得到通知 | 說明說每台帶三個通知開關 | auto:169-173 | BT EveryAbilityAnsweringWithABot… | asserts-oracle | produces-oracle | asserts-oracle 0.72 | produces-oracle 0.91 | ✅ conforms | recheck (mid confidence) |
| AC-25 | 回測只做空 | 送出內容帶交易模式「只做空」；說明說成績單記著 | strat:418-430、450-453 | CB …IsToldItsTradingModeByTheRun | asserts-oracle | produces-oracle | asserts-oracle 0.79 | produces-oracle 0.72 | ✅ conforms | recheck (mid confidence) |
| AC-26 | 回測沒說方向 | 送出內容沒有交易模式；說明說照多空反手 | strat:450-453；TA | CB …IsToldItsTradingModeByTheRun | asserts-oracle | produces-oracle | asserts-oracle 0.80 | produces-oracle 0.29 | ✅ conforms | escalate (low confidence) |
| AC-27 | 現貨交易策略回測說方向 | 現貨回測沒有交易模式格；說明說要做空改用合約重演 | strat:286 | T NoAbilityOffersATradingModeToChoose；T BothReplaysSayTheyOnlyEverTradeSpot | asserts-oracle | produces-oracle | asserts-oracle 0.93 | produces-oracle 0.69 | ✅ conforms | recheck |
| AC-28 | 標的篩選設定不受影響 | 篩選設定照舊有自己的交易模式可填 | screening:36-37 | BT AScreeningProfileStillSaysItsOwnTradingMode | asserts-oracle | produces-oracle | asserts-oracle 0.73 | produces-oracle 0.93 | ✅ conforms | recheck (mid confidence) |
| BR-1 | 交易策略沒有交易模式；送來即整份被拒 | 不宣告、送不出 | strat:114-142 | CB WritingATradingStrategy… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-2 | 交易方向對照、預設、現貨沒有 | 說明含三種對照、建立預設兩者、修改保留、現貨拒絕 | auto:105-108、177-185 | BT Declares…、SaysWhatItsDirectionMeans… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-3 | 改方向的檢查順序；方向沒變不受後兩項 | 說明含三道限制與「方向沒變不受限」 | auto:181-185 | BT SaysWhatItsDirectionMeans… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-4 | 結論通知的預設與語意 | 說明含預設、保留、只管訊息、回到持有、三個都關 | auto:109-112、190-196 | BT Declares…、SaysANotificationOnly… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-5 | 合約交易策略回測的交易模式由那一次說 | 格子在、不給不送 | strat:418-430 | CB …IsToldItsTradingModeByTheRun | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| BR-6 | 外掛原樣轉送、拒絕原話帶回 | body 原樣、拒絕原話 | TA；既有轉交路徑 | BT Forwards…；AT RefusalComesBack… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| NFR-1 | 與交易服務同時上線（交易策略送交易模式即拒） | 外掛不再送交易模式 | strat:114-142 | CB WritingATradingStrategy… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |
| NFR-2 | 外掛不帶自動下單開關，改不了真錢機器人方向 | 機器人寫入不宣告 autoOrderEnabled；總說明寫明 | auto；dependencies.go:73 | 既有 `tool_catalog_binance_auto_order_test.go`；BT TheInstructionsSay… | asserts-oracle | produces-oracle | not run | not run | ✅ conforms | — |

### Recheck / escalate 的複查結果

- **AC-14、AC-15（shadow：shallow；Jev：asserts-oracle 0.45／0.27，disagree）**：`TestWritingAStrategyBotSaysWhatItsDirectionMeansAndWhenItCannotChange` 只斷言兩個 reason 名稱，
  沒有斷言「照原話說出是哪個標的」與「要先讓它空手或等它做完」——拿掉這半句測試仍綠。程式說明本身正確（auto:182）。**判定測試太淺，須補斷言**（見下方「稽核後修正」）。
- **AC-6、7、8、10、18、19（Jev：code unclear 0.09–0.51）**：重讀 auto:105-112、255-263，每一句都在且被 BT 逐字斷言；格子轉送走共用的 `EncodedSubset`，
  Jev 的低信心來自「外掛不實作規則、只說明並原樣轉送」這個設計（與前一刀資金分配的判定同類）。所有句子與格子都經 `/implement` 的突變驗證（拿掉即轉紅）。不改。
- **AC-27**：第一次跑時給錯了程式行號（現貨說明不在那段），改正後 Jev 為 conforms。
- 其餘 `recheck`／`escalate (low confidence)`：重讀對應行，句子在、測試逐字斷言，維持 conforms。

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| `contractTradingModeNote` 也掛在合約腳本重演 | 合約腳本重演本來就由那一次說交易模式；這一刀只是把那句話搬成兩件共用 | 純重構，行為不變 |
| README〈交易方向與結論通知〉 | 文件 | 非行為 |

無任何程式碼落在 Out of Scope（外掛沒有關自動下單、沒有逐結論下單、沒有自己檢查方向）。

## Summary

- Shadow verdict: 34/36 conforms；AC-14、AC-15 shallow（測試沒斷言處置的那半句）。
- Jev status (as returned, 28 AC): ✅ 22 conforms · ❔ 6 unclear
- Violations: none · Mis-asserted (shadow): AC-14、AC-15
- Orphans: 2（皆非行為）
- Jev: ran on 28 AC clauses; BR/NFR not run（與 AC 重疊）
- Needs James: AC-3、6、7、8、10、14、15、17、18、19、20、21、26（route escalate；AC-14／15 已補測試，其餘複查後判定 conforms）

> Static conformance audit against the Acceptance Criteria — it judges test assertions and code paths against the spec's expected outcome, not by running the full suite.

## 稽核後修正

- AC-14、AC-15：在 `TestWritingAStrategyBotSaysWhatItsDirectionMeansAndWhenItCannotChange` 補斷言「照原話說出是哪個標的，要先讓它空手或等它做完」，並以突變驗證（拿掉那半句即轉紅）。
