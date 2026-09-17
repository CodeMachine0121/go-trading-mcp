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

### 或者用 Docker

```bash
docker compose up -d
curl localhost:8090/health   # {"status":"Healthy"}
```

映像檔 7 MB、非 root、沒有掛任何 volume——這個外掛不寫檔案，掛上去也不會讓它記得任何事。

**唯一要注意的一格：容器裡的 `localhost` 是容器自己。** 所以預設值是
`http://host.docker.internal:8080`（交易服務跑在你的機器上），不是 `localhost:8080`。
交易服務之後也搬進 Docker 時，改成 `http://go-trading:8080` 並讓兩邊在同一個網路上，
改法寫在 `docker-compose.yml` 最下面。

連接埠**只綁在 `127.0.0.1`**，這是刻意的——MCP 端點沒有任何門鎖，連得到的人就開得了
一段連線並以自己的帳號登入。要對外開放之前請先讀下面那段。

| 指令 | 用途 |
| :--- | :--- |
| `make docker-up` | 背景啟動 |
| `make docker-down` | 停掉並移除 |
| `make docker-logs` | 跟著看紀錄 |
| `make docker-build` | 重新編映像檔 |

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

## 誰是誰，怎麼分的

MCP over HTTP 在連上時發一個**連線識別碼**（`Mcp-Session-Id`），之後每次請求都要帶著。
外掛就用它當抽屜的鑰匙：一個連線一個抽屜，彼此讀不到。

**那把鑰匙來自傳輸層，不來自助理填的任何欄位**——所以模型沒有辦法要求打開別人的抽屜，
它根本說不出別人的抽屜叫什麼。

兩件要知道的事：

- **重開 Claude Code 就是新的連線識別碼**，舊抽屜變孤兒，要重新登入。外掛重啟也一樣。
- **這道門沒有鎖。** 連得到 `:8090` 的人就開得了一段連線並以自己的帳號登入。
  在本機自己用沒問題；**不要把這個埠對外開放**——連線識別碼本身就等於一把鑰匙，
  走在沒有 TLS 的標頭裡。

**預設只聽 `127.0.0.1`。** 要聽在別張網卡上得自己設 `SERVER_BIND_ADDRESS`，
而且啟動時會印一行警告——開這扇門應該是個明確的動作，不該是預設值。

真的要給多個人用時，別用登入能力，改讓每個人自己帶
`Authorization: Bearer <他自己的登入憑證>`：身分跟著每一次請求走，不放在外掛身上，
重啟不影響，伺服器上也沒有東西可偷。代價是外掛**不替它續用**（續用的那一半不在它手上）。

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
| `SERVER_BIND_ADDRESS` | `127.0.0.1` | 聽在哪張網卡上。**預設只聽本機**——這個端點沒有門鎖 |
| `SERVER_PORT` | `8090` | 外掛自己聽在哪個埠 |
| `MCP_PATH` | `/mcp` | 掛在哪個路徑上 |
| `TRADING_SERVICE_REQUEST_TIMEOUT_SECONDS` | `30` | 單次請求逾時 |
| `LIVE_UPDATE_WAIT_LIMIT_SECONDS` | `10` | 「看一眼即時更新」最長等幾秒 |
| `IDLE_CONNECTION_TIMEOUT_MINUTES` | `60` | 連線閒置多久就放掉（放掉要重新登入） |

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
Dockerfile               多階段建置：靜態編譯 → alpine，非 root
docker-compose.yml       預設交易服務在宿主機上

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
