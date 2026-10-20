# Contract Traceability Matrix — 多標的機器人的資金分配（外掛）

Contract: PRD.md · Design map: ARCH.md · Implementation: `cmd/server/`（`internal/` 未動）
Oracle: Acceptance Criteria (21 scenarios) + Business Rules (7) + NFR (2) = 30 clauses

Path abbreviations: `auto` = `cmd/server/tool_catalog_automation.go`, `CA` = `cmd/server/tool_catalog_capital_allocation_test.go`,
`AT` = `internal/application/tests/api_tool_application_test.go`（`TestEveryAbilityTravelsUnder…` 釘回覆原樣轉交、`TestARefusalComesBackInTheTradingServicesOwnWords` 釘拒絕原話轉交），
`TA` = `internal/domain/models/domains/tool_arguments_domain.go:60`（`EncodedSubset`：只送宣告過且填了的名字，物件原樣轉送）。
`auto:50-86` 是 `positionPlan` 說明；`auto:123-135` 是 `strategyBotWatchedSymbolsNote`；`auto:137-147` 是 `strategyBotCapitalAllocationNote`；`auto:219-225` 是改寫說明的分配段落。

## Clauses

`Test audit` / `Code audit` 是 Claude 的 shadow verdict；`Jev` 欄照 `jev_audit.py` 原樣抄錄（test / code），`Status`／`Route` 亦同。

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Jev | Status | Route |
|----|--------|------------------------|------|------|------------|------------|-----|--------|-------|
| AC-1 | 兩個標的各分到多少 | 說明讓助理說出 600／400（總金額 × 比例 ÷ 100） | auto:137-147；auto:126 | CA SaysTheCapitalIsTheBotsTotal…；CA EveryAbilityAnswering… | asserts-oracle | produces-oracle | asserts 0.91 / produces 0.77 | ✅ conforms | recheck |
| AC-2 | 平均分配留下零頭 | 各分到 333.3、合計 99.99%、0.01% 沒動用 | auto:144 | CA SaysTheCapitalIsTheBotsTotal…（「3 個各 33.33%（總金額 1000 時各分到 333.3，0.01% 沒動用）」） | asserts-oracle | produces-oracle | asserts 0.83 / not-implemented 0.20 | ❌ gap | escalate (disagree) |
| AC-3 | 合約機器人的總金額是保證金 | 保證金 50、名目 250，不是名目 100 | auto:69；auto:140 | CA SaysHowToGive…；CA SaysTheCapitalIsTheBotsTotal… | asserts-oracle | produces-oracle | asserts 0.97 / produces 0.77 | ✅ conforms | recheck |
| AC-4 | 每個標的分到的不夠押固定金額 | 每個只分到 25、押不下 50、「這個標的分到的資金不足」、提醒調大總金額或改倉位大小 | auto:146-147 | CA SaysTheCapitalIsTheBotsTotal… | asserts-oracle | produces-oracle | asserts 0.97 / produces 0.75 | ✅ conforms | recheck |
| AC-5 | 不再說每個標的各自照整份算 | 舊兩句不在；同時開滿不超過總金額、不借給別的標的 | auto:137-147（舊常數已刪） | CA SaysTheCapitalIsTheBotsTotal…（NotContains） | asserts-oracle | produces-oracle | asserts 0.99 / produces 0.95 | ✅ conforms | accept |
| AC-6 | 給每一個標的分配比例 | 送出帶 60／40；回覆的分配比例原樣交給助理 | auto:50；TA | CA ForwardsTheCapitalAllocationsAsGiven；AT TestEveryAbilityTravelsUnder…（回覆原樣） | asserts-oracle | produces-oracle | shallow 0.36 / unclear 0.51 | ❔ unclear | escalate (disagree) |
| AC-7 | 合計少於 100 | 送出 30／30；說明寫可少於 100、剩下不動用 | auto:50、61-65 | CA Forwards…；CA SaysHowToGive… | asserts-oracle | produces-oracle | asserts 0.85 / produces 0.33 | ✅ conforms | escalate (low conf) |
| AC-8 | 不給分配比例 | 不帶任何分配比例；說明讓助理知道會各 33.33% | auto:50；TA；auto:143-144 | CA WithoutCapitalAllocationsSendsNone；CA SaysTheCapitalIsTheBotsTotal… | asserts-oracle | produces-oracle | asserts 0.85 / produces 0.61 | ✅ conforms | recheck |
| AC-9 | 合計超過 100 | 原樣送出；拒絕原話帶回；說明預先寫合計 ≤ 100、照原話轉述 | auto:61-65；既有拒絕轉交路徑 | CA Forwards…（70／40）；CA SaysHowToGive…；AT TestARefusalComesBack… | asserts-oracle | produces-oracle | asserts 0.53 / unclear 0.19 | ❔ unclear | escalate (disagree) |
| AC-10 | 只給一部分標的 | 說明寫每個監聽標的都要給；拒絕原話帶回 | auto:63 | CA SaysHowToGive…；AT TestARefusalComesBack… | asserts-oracle | produces-oracle | asserts 0.72 / produces 0.07 | ✅ conforms | escalate (low conf) |
| AC-11 | 超過兩位小數 | 原樣送出 33.333；說明寫最多兩位小數 | auto:64 | CA Forwards…（33.333）；CA SaysHowToGive… | asserts-oracle | produces-oracle | asserts 0.28 / not-implemented 0.21 | ❌ gap | escalate (disagree) |
| AC-12 | 指到不在清單上的標的 | 原樣送出 SOLUSDT；說明寫只能指名監聽標的 | auto:63 | CA Forwards…（SOLUSDT）；CA SaysHowToGive… | asserts-oracle | produces-oracle | asserts 0.27 / unclear 0.20 | ❔ unclear | escalate (disagree) |
| AC-13 | 說明列出分配比例的規則 | 可不給、每個都給、不重複、> 0 ≤ 100、兩位小數、合計 ≤ 100 | auto:61-65 | CA SaysHowToGive… | asserts-oracle | produces-oracle | asserts 0.88 / produces 0.28 | ✅ conforms | escalate (low conf) |
| AC-14 | 只改名字時不必重帶分配 | 修改說明：標的不變即保留、只改名字不必重帶 | auto:222-223 | CA RewritingAStrategyBotSaysWhen… | asserts-oracle | produces-oracle | asserts 0.92 / produces 0.39 | ✅ conforms | escalate (low conf) |
| AC-15 | 只改順序也算標的不變 | 修改說明：順序不論 | auto:222 | CA Rewriting… | asserts-oracle | produces-oracle | asserts 0.92 / produces 0.47 | ✅ conforms | escalate (low conf) |
| AC-16 | 加標的而沒給分配會重新平均 | 重新平均、調過的會不見、先說明或一起送 | auto:223-224 | CA Rewriting… | asserts-oracle | produces-oracle | asserts 0.95 / produces 0.61 | ✅ conforms | recheck |
| AC-17 | 改成各半 | 修改送出帶 50／50 | auto:50；TA | CA Forwards…（update × 50／50） | asserts-oracle | produces-oracle | mis-asserted 0.59 / not-implemented 0.21 | ❌ gap | escalate (disagree) |
| AC-18 | 每個標的都帶著分配比例 | 說明讓助理說出 60%（600）、30%（300）、10% 沒動用；回覆原樣 | auto:126-128；回覆原樣轉交 | CA EveryAbilityAnswering…；AT TestEveryAbilityTravelsUnder… | asserts-oracle | produces-oracle | no-test 0.16 / not-implemented 0.27 | ❌ gap | escalate (disagree) |
| AC-19 | 只盯一個標的 | 100%、訊息不多一行 | auto:129 | CA EveryAbilityAnswering…（「100% 時訊息不多這一行」） | asserts-oracle | produces-oracle | asserts 0.80 / produces 0.31 | ✅ conforms | escalate (low conf) |
| AC-20 | 分配不是 100% 時訊息多一行 | 說明含「資金分配 33.33%：總金額 1000 中的 333.3」 | auto:129 | CA EveryAbilityAnswering… | asserts-oracle | produces-oracle | asserts 0.87 / not-implemented 0.25 | ❌ gap | escalate (disagree) |
| AC-21 | 合約機器人開著自動下單 | 每個標的仍真的下單、用分到的那一份、與訊息同一個數字 | auto:147 | CA SaysTheCapitalIsTheBotsTotal… | asserts-oracle | produces-oracle | asserts 0.97 / produces 0.81 | ✅ conforms | recheck |
| BR-1 | 總金額；合約是保證金 | 同 AC-3 | auto:56、69、137-141 | CA SaysHowToGive…、SaysTheCapital… | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| BR-2 | 分到的資金公式、押不下措辭 | 同 AC-1、AC-4 | auto:138、146-147 | CA SaysTheCapital… | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| BR-3 | 分配固定、不借用、合計可 < 100 | 說明含兩者 | auto:142、65 | CA SaysTheCapital…、SaysHowToGive… | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| BR-4 | 分配比例規則、外掛原樣轉送不檢查 | 說明列規則；body 原樣 | auto:61-65；TA | CA Forwards…、SaysHowToGive… | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| BR-5 | 不給：建立平均、修改保留或重新平均 | 說明含三者 | auto:143-144、222-224 | CA SaysTheCapital…、Rewriting… | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| BR-6 | 修改只能在已停止時 | 修改說明照舊寫只有已停止的改得動 | auto:216（未變） | 既有 `tool_catalog_test.go` 改寫說明斷言 | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| BR-7 | 讀機器人：分配比例、外掛不另算 | 說明含 allocationPercentage、回覆沒有這個數字要自己算 | auto:126-128 | CA EveryAbilityAnswering… | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| NFR-1 | 交易服務向下相容；外掛之後合併 | 不給分配比例照樣能寫（不送） | TA | CA WithoutCapitalAllocationsSendsNone | asserts-oracle | produces-oracle | not run | ✅ conforms | — |
| NFR-2 | 自動下單仍只能在網頁切換 | 建立／修改不宣告 autoOrderEnabled | auto（未變） | 既有 `tool_catalog_binance_auto_order_test.go` | asserts-oracle | produces-oracle | not run | ✅ conforms | — |

### Recheck / escalate 的複查結果

- **AC-2、AC-11、AC-12、AC-17、AC-18、AC-20（Jev：not-implemented／unclear，信心 0.19–0.27）**：逐一重讀 `auto:123-150`、`auto:50-86`，句子都在指定範圍內且被 CA 逐字斷言；AC-11／12／17 的「送出」由 CA `TestWritingAStrategyBotForwardsTheCapitalAllocationsAsGiven` 對兩件寫入能力逐案斷言（33.333、SOLUSDT、50／50 皆在案例表）。拒絕本身是交易服務的行為，外掛只轉送與預先說明——與前一刀同類情境的判定一致。Jev 的低信心來自「外掛不實作規則」這個設計，不是缺漏。全部突變驗證過（拿掉段落、倒轉改寫規則、改名 `positionPlan` 格、塞回「N 倍」皆讓對應測試轉紅）。
- **AC-6（Jev：shallow）**：送出那一半由 CA 斷言；回覆那一半是所有能力共用的原樣轉交路徑，由 AT `TestEveryAbilityTravelsUnderTheConnectorAuthorizationTheCallerBrought` 斷言內容逐字相同。外掛沒有任何針對機器人回覆的轉換，所以不另寫機器人專屬的回覆測試（application 測試拿不到 `cmd/server` 的能力清單，硬寫只會複製共用測試）。不改。

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| auto:55-58 sizingMode 三種模式改說「在分到的資金裡」 | PRD 的 BR-2 涵蓋（倉位大小模式在那一份裡運作），只是沒有獨立情境 | 屬 BR-2，保留 |
| `strategyBotWriteNote`（auto:156） | improve-codebase 把建立與改寫共用的說明尾段收成一份 | 純重構，無行為 |
| README〈資金分配〉一節 | 文件 | 非行為 |

## Summary

- Shadow verdict: 30/30 clauses conforms.
- Jev status (as returned): ✅ 13 conforms · ❌ 5 gap（AC-2、11、17、18、20）· ❔ 3 unclear（AC-6、9、12）——全部經上方複查判定為 Jev 對「說明即實作」的低信心誤判，無需改碼。
- Violations: none · Mis-asserted: none（Jev 對 AC-17 的 mis-asserted 0.59 經複查不成立：update 50／50 在案例表內）
- Orphans: 3（皆非行為或已屬 BR）
- Jev: ran on 21 AC clauses; BR/NFR not run（與 AC 重疊）
- Needs James: AC-2、6、7、9、10、11、12、13、14、15、17、18、19、20（route escalate，原因如上；Claude 複查後判定 conforms）

> Static conformance audit against the Acceptance Criteria — it judges test assertions and code paths against the spec's expected outcome, not by running the full suite.
