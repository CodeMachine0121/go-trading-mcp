# go-trading-mcp

讓 **AI 助理**代你操作 [`go-trading`](../go-trading) 的一層外掛（MCP server，純 Go）。

交易服務會的每一件事——查行情、算指標、寫策略腳本、組交易策略、重演一段行情、
養策略機器人、推通知、跟行情助手對話——都成為助理呼叫得到的一個**能力**，
**一件不漏**，目前 **54 件**。

你在助理裡把帳號交出去一次，之後說人話就能操作自己的東西，不必離開對話去點畫面，
也不必每次重新登入：**登入過期外掛會自己換新，你不會被打斷。**

## 它做什麼、不做什麼

| | |
| :--- | :--- |
| **做** | 保管你的身分、把助理的要求轉達給交易服務、把交易服務的回答（含拒絕的原話）帶回來 |
| **不做** | 不自己存行情、不自己算指標、不下單、**不改變任何一條業務規則** |

交易服務拒絕的，外掛不放行；交易服務允許的，外掛不多問。
**它沒有自己的資料庫**——身分只活在行程裡，重啟就要重新登入，這是刻意的。

## 快速開始

先把 `go-trading` 跑起來（預設 `localhost:8080`），然後：

```bash
go mod download
make start        # 啟動於 :8090，MCP 掛在 /mcp
curl localhost:8090/health   # {"status":"Healthy"}
```

接到助理（以 Claude Code 為例）：

```bash
claude mcp add --transport http go-trading http://localhost:8090/mcp
```

然後在對話裡說一句「幫我登入，帳號是 …」，接著就能直接問：

> 「BTCUSDT 最近一天每小時的形狀？」
> 「把我那支布林通道在台積電上用 1 小時刻度回測 2026 年，起始資金 100 萬。」
> 「幫我開一台機器人，每 15 分鐘跑一次那份交易策略。」

## 登入是怎麼回事

交易服務發的是**一對**憑證：一份**登入憑證**（十五分鐘，每次操作都帶著）、
一份**續用憑證**（三十天，**只能用一次**，只拿來換新的一對）。

外掛替你保管這一對：

- **密碼用完即丟。** 只用來換一份登入，之後一個字都不留。
- **憑證不會出現在助理看得到的回答裡。** 登入成功只告訴你「你是誰、到期於何時」。
- **一個連線一份登入。** 同一台外掛同時服務很多人，彼此的身分互不相見。
- **過期不打擾人。** 外掛自己拿續用憑證換一份新的，把原本那件事做完。
- **續用只花一次。** 好幾件事同時撞上過期時，只換一次、大家共用換來的那一份——
  重複花掉會讓交易服務把整條換發鏈作廢，**把真正的你登出**。

三個外掛自己的能力：`trading_sign_in`、`trading_sign_out`、`trading_renew_session`。

**已經有憑證的人**可以跳過登入：在請求標頭帶 `Authorization: Bearer <你的登入憑證>`，
外掛會照用。自備的身分外掛**不替你續用**——那一半不在它手上。

## 做不成的時候，它會說清楚是哪一種

四句話絕不混用，因為它們要你做的事完全不同：

| 回覆 | 意思 | 你該做什麼 |
| :--- | :--- | :--- |
| `請先登入` | 這件事要知道是誰在做，而還沒有人說 | 先登入 |
| `登入已失效，請重新登入` | 登入過，但那一份救不回來了 | 重新登入 |
| `連不到交易服務` | **你送的沒有錯**，是它現在不在 | 等一下，送同一件事 |
| （交易服務的原話） | 規則不通過、不是你的、額度用完… | 照它說的改一改再送 |

**拒絕的原因一字不改地帶回來**，助理才知道該把回溯天數改成多少、該去哪裡拿權限。

## 設定

全部有預設值，`.env` 可整份省略。見 `.env.example`。

| 變數 | 預設值 | 用途 |
| :--- | :--- | :--- |
| `TRADING_SERVICE_BASE_URL` | `http://localhost:8080` | 交易服務的位址。**填錯這一個，其餘全部不會動** |
| `SERVER_PORT` | `8090` | 外掛自己聽在哪個埠 |
| `MCP_PATH` | `/mcp` | 掛在哪個路徑上 |
| `TRADING_SERVICE_REQUEST_TIMEOUT_SECONDS` | `30` | 單次請求逾時 |
| `LIVE_UPDATE_WAIT_LIMIT_SECONDS` | `10` | 「看一眼即時更新」最長等幾秒 |

打錯的數字會**退回預設值而不是讓外掛起不來**——一個時間打錯，不該讓整個東西掛掉。

## 即時跟盤只給「看一眼」

`trading_peek_live_k_candle` 收到第一則更新就回，最多等 `LIVE_UPDATE_WAIT_LIMIT_SECONDS` 秒。
**等滿沒收到東西是正常結果，不是錯誤**——凌晨三點的台股本來就是這樣。
要連續看就重複呼叫；外掛不維持一條永遠開著的通道。

## 新增一件能力

交易服務多了一個端點時，**只要在 `cmd/server/tool_catalog*.go` 的清單補一筆**：

```go
domains.NewApiToolDomain(
    "trading_do_the_new_thing",
    "它做什麼、哪些欄位必填、什麼情況會被拒絕（這段是寫給助理看的）",
    vo.RequestVerbSubmit, "/the/new/{id}", true,
    pathParameter("id", "哪一個"),
    bodyParameter("amount", vo.ToolParameterKindString, "多少", true),
)
```

**不必新增 handler、不必改 service、不必改 controller、不必改 proxy。**
五十幾件事共用同一段流程，差異全部收在這份清單裡。

`cmd/server/tool_catalog_test.go` 有一份**手寫的**「交易服務有什麼」清單守著這件事：
那邊多一個端點，測試就會紅，直到這邊也補上為止。

## Commands

| 指令 | 用途 |
| :--- | :--- |
| `make start` | 啟動（`go run ./cmd/server`） |
| `make build` | 編譯到 `bin/server` |
| `make test` | 跑全部測試 |
| `make mock` | 重新產生 mock（`go generate ./...`） |
| `make vet` | 靜態檢查 |

`mockgen` 已用 `tool` 指示詞釘在 `go.mod`，不需要另外全域安裝。

commit 前請自行跑過 `go build ./... && go vet ./... && go test ./...`。

## 專案結構

```
cmd/server/
├── main.go              進入點與有序關機
├── config.go            環境變數
├── dependencies.go      組裝根：手動 DI、掛上 HTTP
└── tool_catalog*.go     ★ 能力清單——唯一列出五十幾件事的地方

internal/
├── controller/          MCP 請求/回應轉換（一次呼叫是誰打來的、答案怎麼呈現）
├── application/         用例編排
├── domain/              核心，不認識 HTTP、不認識 MCP
│   ├── models/
│   │   ├── domains/     行為所在地：一件能力怎麼變成一個請求、一份身分還能不能用
│   │   ├── dto/         對 application 的回傳形狀
│   │   └── vo/          不可變值
│   ├── service/         代辦一件事、保管身分
│   └── interface/       對外介面（mocks/ 放產生的 mock）
└── infrastructure/
    ├── tradingservice/  ★ 唯一知道 HTTP 存在的地方（含即時跟盤那條線）
    ├── persistence/     身分的記憶體保管
    └── clock/           系統時鐘
```

依賴方向一律指向 `domain/`。規範見 [`.claude/rules/`](.claude/rules/)，入口是 [`CLAUDE.md`](CLAUDE.md)。
需求與設計見 [`.sdd/`](.sdd/)。
