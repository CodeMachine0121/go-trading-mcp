# 幣安自動下單設定（外掛只讀） — Architecture Design

**Status:** Confirmed
**Source PRD:** `.sdd/2026-10-01-binance-auto-order-setup/PRD.md`
**Tech context:** Go · MCP 外掛 · 能力清單即功能（`cmd/server/tool_catalog*.go` 宣告，`ApiToolDomain` 單一路徑轉達）

---

## 1. Design Goal & Guiding Principle

- **In one sentence:** 多一件只讀能力 `trading_get_binance_trading_key_status`（`GET /users/me/binance-trading-key/status`），
  機器人能力的說明寫出回覆帶 `autoOrderEnabled` 與「目前開著也還不會下單」，伺服器說明寫出寫入只能在網頁上做、不收也不複述金鑰。
- **Guiding principle:** **「做不到」靠能力清單裡沒有，不靠擋。** 外掛的寫入能力只轉送宣告過的欄位（`ApiToolDomain.BuildRequest`
  只把宣告過的 body 欄位交給 `ToolArgumentsDomain.EncodedSubset`），所以建立／修改機器人**不宣告** `autoOrderEnabled` 就保證它永遠送不出去；
  存金鑰、開關自動下單**不宣告成能力**，就沒有任何一條路可以呼叫。外掛不需要新增任何檢查程式碼，`internal/` 一行不動。

---

## 2. Change Scope

| Area | Action | What / Why |
| :--- | :--- | :--- |
| 新檔 `cmd/server/tool_catalog_binance_trading_key.go` | **Add** | `binanceTradingKeyApiTools()`：唯一一件 `trading_get_binance_trading_key_status`（Read，無欄位）；`binanceAutoOrderWebOnlyNote`：「設定金鑰、開關自動下單只能在網頁上做、不要求也不複述金鑰」，狀態能力與機器人寫入能力共用這一段 |
| `cmd/server/tool_catalog.go` `apiToolCatalog` | **Modify** | 加一行 `binanceTradingKeyApiTools()` |
| `cmd/server/tool_catalog_automation.go` `strategyBotApiTools` | **Modify** | 新常數 `strategyBotAutoOrderNote`（回覆帶 `autoOrderEnabled`、目前不下單、只送 Telegram）掛在 list / get / start / run_now；create / update 說明加「沒有自動下單這一格、建立一律關、修改保留原本的」+ 網頁才能做 |
| `cmd/server/dependencies.go` 伺服器說明 | **Modify** | 加一段：幣安交易金鑰與自動下單外掛只讀；寫入只在網頁設定頁／機器人詳細頁；不要求貼金鑰、貼了不複述不轉送；自動下單目前不下單 |
| `cmd/server/tool_catalog_test.go` 能力全集 | **Modify** | 補 `trading_get_binance_trading_key_status`；同時明列**不得出現**的寫入能力名 |
| 新測試 `cmd/server/tool_catalog_binance_auto_order_test.go` | **Add** | 釘住 PRD 每一條情境 |
| `README.md` | **Modify** | 新一節「幣安自動下單（只讀）」 |
| `internal/` 全部 | **Not touched** | 轉達路徑已經「沒宣告就不送」、回覆原樣轉交（`autoOrderEnabled` 自然出現在每個機器人回覆裡） |
| 「我是誰」能力 `trading_get_current_user` | **Not touched** | 決策：**另立一件只讀能力**而不併入。併入要外掛為一次詢問打兩個入口並自己拼回覆，破壞「一件能力＝交易服務一件事」的單一路徑；「我是誰」回答身分，與交易設定無關；另立也讓助理只在被問到幣安時才去讀 |

---

## 3. New Classes / Modules

無新型別。新增的是能力宣告與說明常數：

| Name | Kind | Responsibility (purpose) | Collaborators | Satisfies (PRD scenario) |
| :--- | :--- | :--- | :--- | :--- |
| `binanceTradingKeyApiTools()` | 能力宣告 | 宣告讀取幣安交易金鑰狀態這一件 | `ApiToolDomain` | 已設定、尚未設定 |
| `binanceAutoOrderWebOnlyNote` | 說明常數 | 一份「只能在網頁上做」的話，狀態能力與兩件機器人寫入能力共用，避免三處說法漂移 | — | 要求打開自動下單、貼出金鑰、建立時不能順帶打開 |
| `strategyBotAutoOrderNote` | 說明常數 | 一份「回覆帶 autoOrderEnabled、目前不下單」的話，所有讀得到機器人的能力共用 | — | 列出／查看機器人 |

---

## 4. Modified Components

| Component | Current role | Change needed |
| :--- | :--- | :--- |
| `strategyBotApiTools()` | 九件機器人能力 | 說明附上兩個新常數；欄位**不變**（刻意不加 `autoOrderEnabled`） |
| `buildMcpServer` 的 `Instructions` | 外掛總說明 | 加一段幣安自動下單只讀規矩 |
| `apiToolCatalog` | 能力全集 | 多一組 |

---

## 5. Component Relationships

```mermaid
flowchart LR
    Assistant --> Catalog[apiToolCatalog]
    Catalog --> KeyTools[binanceTradingKeyApiTools]
    Catalog --> BotTools[strategyBotApiTools]
    KeyTools --> Note1[binanceAutoOrderWebOnlyNote]
    BotTools --> Note1
    BotTools --> Note2[strategyBotAutoOrderNote]
    KeyTools --> ApiTool[ApiToolDomain.BuildRequest]
    BotTools --> ApiTool
    ApiTool --> Status[/users/me/binance-trading-key/status]
    ApiTool --> Bots[/strategy-bots]
```

---

## 6. Extensibility & Handoff Notes

- **Most likely next requirement:** 下一刀交易服務真的會下單——「目前開著也還不會下單」這句要改成下單後的規則，且可能多出「查看下單紀錄」之類的只讀能力。
- **Where it lands:** `strategyBotAutoOrderNote` 一個常數（所有機器人讀取能力共用）與伺服器說明的那一段；新的只讀能力在 `binanceTradingKeyApiTools()` 或機器人那一組補一列。
- **How to add it:** 改那個常數的文字、補一列能力宣告與能力全集測試一行；不碰 `internal/`。
- **Do not hardcode:** 不要在外掛裡判斷可交易市場是否涵蓋機器人種類——那是交易服務的規則。
- **Known debt / deferred:** 「不複述金鑰」只能靠說明約束助理，外掛無法技術上阻止；若日後要更強，可考慮在 `ToolArgumentsDomain` 層拒收像金鑰的字串，但目前沒有任何能力收金鑰，不值得。

---

## 7. Traceability

| PRD Scenario | Fulfilled by |
| :--- | :--- |
| 列出機器人時標出自動下單（例 1） | `strategyBotAutoOrderNote` on `trading_list_strategy_bots`；回覆原樣轉交 |
| 查看單一台也看得到 | `strategyBotAutoOrderNote` on `trading_get_strategy_bot` |
| 已設定（例 2） | `trading_get_binance_trading_key_status` → `GET /users/me/binance-trading-key/status`（回覆本身不含任何金鑰內容） |
| 尚未設定（例 3） | 同上的說明：configured 為 false 時只說尚未設定、指引網頁設定頁、不說出設定時刻 |
| 要求打開自動下單（例 4） | 能力清單沒有開關能力（能力全集測試）+ `binanceAutoOrderWebOnlyNote` + 伺服器說明 |
| 在對話裡貼出金鑰（例 5） | 能力清單沒有存金鑰能力 + 伺服器說明「不要求、不複述、不轉送」 |
| 改機器人不動自動下單（例 6） | `trading_update_strategy_bot` 不宣告 `autoOrderEnabled` → `BuildRequest` 不轉送；交易服務修改時保留原本的開關 |
| 建立機器人時不能順帶打開 | `trading_create_strategy_bot` 不宣告 `autoOrderEnabled` + 說明「建立一律關」+ `binanceAutoOrderWebOnlyNote` |

---

## 8. Risks & Open Decisions

- **Risks / trade-offs:** 舊版交易服務沒有狀態入口時，那一件會回交易服務的原話（找不到）；可接受，外掛與交易服務同步上線。
- **Open decisions:** 無。
