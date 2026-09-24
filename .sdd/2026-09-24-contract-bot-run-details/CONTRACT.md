# Contract Conformance — 助理講得準合約機器人的建議部位與執行紀錄

**Contract:** `PRD.md` (Acceptance Criteria) · **Implementation:** `cmd/server/tool_catalog_automation.go` · **Tests:** `cmd/server/tool_catalog_contract_bot_run_details_test.go`

> Static conformance audit: oracles derived from the PRD before reading code; tests and code judged independently against them.

## Clauses

| ID | Clause | Oracle | Implementation | Test | Test audit | Code audit | Status |
|---|---|---|---|---|---|---|---|
| AC-01 | 建議部位照交易所規則算 | 建立與修改的部位規劃說明都寫出：步進往下取整、照取整後數量重算保證金名目、價格跳動單位、預估強平價（名目那一級／無分級用最小那一級）、資金費率估算（費率×名目、估算、無紀錄不估） | `tool_catalog_automation.go:60-66` | `TestWritingABotSaysAContractSuggestionFollowsTheVenuesRules` | asserts-oracle | produces-oracle | ✅ |
| AC-02 | 交易所不收的那一筆 | 寫出三種不收情況、只說不收與原因、建議調整什麼 | `:67-70` | `TestWritingABotNamesWhenTheVenueRefusesAContractSuggestion` | asserts-oracle | produces-oracle | ✅ |
| AC-03 | 還沒有交易規格 | 寫出照舊建議但不取整、不估強平價、訊息會說明 | `:71-72` | `TestWritingABotSaysWhatHappensWithoutATradingSpecification` | asserts-oracle | produces-oracle | ✅ |
| AC-04 | 強平警告換成跟預估強平價比 | 寫出止損比預估強平價遠時警告；100% 規則只當退路 | `:66`, `:72-73` | `TestWritingABotComparesTheStopWithTheEstimatedLiquidationPrice` | asserts-oracle (含舊句不再出現) | produces-oracle | ✅ |
| AC-05 | 三樣新東西只在合約機器人有建議的那一輪出現 | 讀輪次兩個能力寫出「有建議部位時」多帶方向、槓桿、名目 | `:82-86`, `:154-155`, `:161-162` | `TestReadingABotsRoundsNamesTheContractRunDetails` | asserts-oracle（follow-up 補釘「那一輪有建議部位時」） | produces-oracle | ✅ |
| AC-06 | 現貨與交易所不收的那一輪沒有 | 寫出兩者沒有這三樣 | `:85` | 同上 | asserts-oracle | produces-oracle | ✅ |
| AC-07 | 開倉金額是取整後的保證金 | 寫出開倉金額是保證金、取整後、與訊息一致 | `:86` | 同上 | asserts-oracle（follow-up 補釘「與當時送出的訊息一致」） | produces-oracle | ✅ |
| AC-08 | 現貨說明一字不差 | 現貨不收槓桿、押多少的句子照舊 | 未修改的句子 | `TestWritingABotKeepsTheSpotSentences` | asserts-oracle | produces-oracle | ✅ |
| BR-01 | 數字由交易服務算、外掛原樣轉交 | 沒有新增任何計算或轉送變化 | 只改字串 | 既有轉送測試（`tool_catalog_contract_strategy_bot_test.go`） | asserts-oracle | produces-oracle | ✅ |
| BR-02 | 新規則寫在部位規劃說明；三樣寫在讀輪次兩個能力 | 同上位置 | 同上 | 各測試依能力取說明 | asserts-oracle | produces-oracle | ✅ |
| NFR-01 | 能力名稱、輸入格子、送出內容不變 | 目錄名稱、參數、轉送不變 | — | `tool_catalog_test.go` 既有目錄測試 | asserts-oracle | produces-oracle | ✅ |

## Orphans

None.

## Summary

11 clauses · ✅ 11 conforms · 🔴 0 · 🟠 0 (2 shallow assertions tightened during the audit) · 🟡 0 · ❌ 0 · ❔ 0 · ⚠️ 0 — Conformance 100%.


## Follow-up — code review corrections

A code review found five statements that were inaccurate against the trading service; the PRD, the descriptions and their tests were corrected together:

| # | Correction | Pinned by |
|---|---|---|
| 1 | Rounded stake and stops only when the contract had a trading specification; unrounded otherwise and on older records; older records lack the three fields | `TestReadingABotsRoundsNamesTheContractRunDetails` |
| 2 | A refused order still shows the (unrounded) margin line, then one reason line with its figures; no quantity, stops, liquidation price or funding | `TestWritingABotNamesWhenTheVenueRefusesAContractSuggestion` |
| 3 | A long whose estimate is at or below zero shows 「這個槓桿下不會被強制平倉」 and no warning | `TestWritingABotSaysALongThatCannotBeLiquidatedGetsNoPrice` |
| 4 | The warning compares the stop with the unrounded estimate; a stop exactly at it counts as liquidation first | `TestWritingABotSaysTheWarningComparesTheUnroundedEstimate` |
| 5 | Run-now answers with the bot, so it points to the run history instead of describing round fields | `TestRunningABotNowPointsToTheRoundsRecord` |

All clauses conform after the corrections; each correction survived a deliberate break (6 mutants, all killed).
