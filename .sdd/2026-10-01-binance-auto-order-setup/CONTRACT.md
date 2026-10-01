# Contract Traceability Matrix — 幣安自動下單設定（外掛只讀）

Contract: PRD.md
Design map: ARCH.md
Implementation: `cmd/server/tool_catalog_binance_trading_key.go`, `cmd/server/tool_catalog_automation.go`, `cmd/server/dependencies.go`
Oracle: Acceptance Criteria + Business Rules + NFR (13 clauses)

> Static conformance audit: test assertions and code paths were judged against the spec's expected outcome, not by running the full suite.
> The connector's "behavior" toward the assistant is the ability list and the words in descriptions/instructions; the bridge for every
> "助理說／不說" outcome is the sentence an assistant reads before acting.

## Clauses

| ID | Clause | Spec-expected (oracle) | Impl | Test | Test audit | Code audit | Status |
|----|--------|------------------------|------|------|------------|------------|--------|
| AC-1 | 列出機器人時標出自動下單（例 1） | 列出的每台帶出自動下單開／關，並說明目前開著也不會下單、只送 Telegram | tool_catalog_automation.go:97,127 | tool_catalog_binance_auto_order_test.go:25 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-2 | 查看單一台也看得到 | 單一台回覆帶出自動下單開／關 | tool_catalog_automation.go:136 | tool_catalog_binance_auto_order_test.go:25 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-3 | 已設定（例 2） | 說出已設定與可交易市場；不含金鑰任何一字 | tool_catalog_binance_trading_key.go:14-25（只讀不含金鑰的狀態入口） | tool_catalog_binance_auto_order_test.go:37,48 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-4 | 尚未設定（例 3） | 只說尚未設定，指引網頁設定頁，不說出設定時刻 | tool_catalog_binance_trading_key.go:20-21 | tool_catalog_binance_auto_order_test.go:48 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-5 | 要求打開自動下單（例 4） | 說明只能在網頁機器人詳細頁做；沒有任何改動送出 | 能力清單無開關能力；dependencies.go:69-71；tool_catalog_binance_trading_key.go:8 | tool_catalog_binance_auto_order_test.go:61；dependencies_test.go:137；tool_catalog_test.go 能力全集 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-6 | 在對話裡貼出金鑰（例 5） | 說明外掛不能存金鑰、請到網頁設定頁；不複述不轉送 | 能力清單無存金鑰能力；tool_catalog_binance_trading_key.go:8-11 | dependencies_test.go:137；tool_catalog_binance_auto_order_test.go:61 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-7 | 改機器人不動自動下單（例 6） | 觸發間隔被改；自動下單維持原樣（不被送出） | tool_catalog_automation.go:27（無 autoOrderEnabled 欄位）、146 | tool_catalog_binance_auto_order_test.go:73,89 | asserts-oracle | produces-oracle | ✅ conforms |
| AC-8 | 建立機器人時不能順帶打開 | 建立成功、自動下單關；說明只能在網頁做 | tool_catalog_automation.go:27、118-119 | tool_catalog_binance_auto_order_test.go:73,89 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-1 | 讀到的狀態是交易服務原話，外掛不推算、不補值（回覆沒有那一格時不臆測） | 原樣轉交；說明要助理沒有那一格時不猜 | tool_catalog_automation.go:99 | tool_catalog_binance_auto_order_test.go:25（修正後） | asserts-oracle | produces-oracle | ✅ conforms（原為 🟡 partial，已補斷言） |
| BR-2 | 沒有存入／更換／移除金鑰、開關自動下單的能力 | 能力清單裡沒有任何這類能力 | tool_catalog.go 能力全集 | tool_catalog_test.go 能力全集；tool_catalog_binance_auto_order_test.go:61 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-3 | 建立與修改不帶自動下單，填了也不轉送 | 送出的內容不含自動下單 | ApiToolDomain.BuildRequest 只轉送宣告欄位 | tool_catalog_binance_auto_order_test.go:73 | asserts-oracle | produces-oracle | ✅ conforms |
| BR-4 | 另立只讀能力，不併入「我是誰」 | 有一件獨立的讀取能力，「我是誰」不變 | tool_catalog_binance_trading_key.go:14 | tool_catalog_binance_auto_order_test.go:37；tool_catalog_test.go 能力全集 | asserts-oracle | produces-oracle | ✅ conforms |
| NFR-1 | 外掛不經手任何一段金鑰 | 沒有收金鑰的欄位；讀的是不含金鑰的那一份 | tool_catalog_binance_trading_key.go:25 | tool_catalog_binance_auto_order_test.go:37,61 | asserts-oracle | produces-oracle | ✅ conforms |

## Orphans (code with no clause)

| Code | Description | Verdict |
|------|-------------|---------|
| tool_catalog_automation.go:160,167,184 | 啟動、停掉、立刻跑一輪三件能力的說明也附上自動下單那段話；停掉另說「停掉不會關掉它的自動下單」 | undocumented but consistent with the backend rule (停掉只改執行狀態，不動開關) and with the brief's "每個回覆機器人的能力都帶出自動下單"；kept |

## Summary

- Conforms: 13/13 clauses ✅ (100%) after one fix
- Violations: none
- Mis-asserted: none
- Partial: BR-1 → fixed by asserting「不要猜它是開還是關」in the every-bot-ability test
- Gaps: none
- Unclear: none
- Orphans: 1 (benign, kept)
