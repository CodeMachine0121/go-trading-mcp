# go-trading-mcp

讓 **AI 助理**代你操作 [`go-trading`](../go-trading) 的一層外掛（MCP server，純 Go）。

交易服務會的每一件事——查行情、算指標、寫策略腳本、組交易策略、重演一段行情、
養策略機器人、推通知——都成為助理呼叫得到的一個**能力**，目前 **67 件**。

你在助理裡把帳號交出去一次，之後說人話就能操作自己的東西，不必離開對話去點畫面，
也不必每次重新登入：**登入過期外掛會自己換新，你不會被打斷。**

**已經有一台跑在 `https://trading-mcp.coding-afternoon.com/mcp`**，不必自己架：

```bash
claude mcp add --scope user --transport http \
  go-trading https://trading-mcp.coding-afternoon.com/mcp
```

（`/mcp` 與 `--scope user` 兩樣都不能省，理由見〈快速開始〉。）

## 它做什麼、不做什麼

| | |
| :--- | :--- |
| **做** | 保管你的身分、把助理的要求轉達給交易服務、把交易服務的回答（含拒絕的原話）帶回來 |
| **不做** | 不自己存行情、不自己算指標、不下單、**不改變任何一條業務規則** |
| **刻意不做** | **不代理交易服務內建的行情對話助手**——見下 |

交易服務拒絕的，外掛不放行；交易服務允許的，外掛不多問。
**它沒有自己的資料庫**——身分只活在行程裡，重啟就要重新登入，這是刻意的。

### 為什麼少了「行情對話助手」

交易服務有一個內建的 AI 對話助手，而這個外掛**刻意不把它接出來**。

代理它等於讓**一個 AI 自己決定去花另一個 AI 的錢**：那個助手一次回答來回數十趟、
要好幾分鐘、按 token 計費，而中間沒有任何人判斷過這一次值不值得。

你自己去用它完全沒問題（`POST /chat`）。不該存在的是「助理可以自己伸手拿」這件事，
所以它不在清單上——而**不在清單上的能力，沒有辦法被誤呼叫**。
`cmd/server/tool_catalog_test.go` 守著這條界線。

## 快速開始

有兩條路：**接到已經跑著的那一台**，或**自己在本機跑一台**。
日常用前者，改這個外掛的程式碼才需要後者。

### 接到已經部署好的那一台

```bash
claude mcp add --scope user --transport http \
  go-trading https://trading-mcp.coding-afternoon.com/mcp
```

然後 `/mcp` 應該看到 `go-trading ✔ Connected`。

這一行有**兩個地方踩過坑**，都不會給你看得懂的錯誤訊息：

**一、網址結尾一定要有 `/mcp`。**

MCP 掛在 `/mcp`，根路徑上什麼都沒有。少打這三個字，助理會拿到 404，
接著去試 OAuth 自動註冊，那條路一樣 404，於是畫面上寫的是
`Dynamic Client Registration rejected (HTTP 404)`——
一句完全指不到「你網址打短了」的話。

```
https://trading-mcp.coding-afternoon.com/mcp   →  200   ✅
https://trading-mcp.coding-afternoon.com/      →  404   ❌
```

**二、`--scope user` 不要省。**

Claude Code 把所有設定放在**同一個** `~/.claude.json`，而每個開著的 session
各自握著一份自己讀進來的副本。存檔時是整份寫回去，所以**比你晚存檔、但比你早開啟的
那個 session，會把你剛改好的設定連同它的舊值一起蓋掉**。

專案層那一格特別容易被這樣輾過。寫進 user scope 至少換一格。
真的又被蓋掉的話：先把其他 Claude Code 視窗關掉，再跑一次上面那行。

> **這個端點沒有帳號密碼那一關**，這是刻意的。曾經有過一道 basic auth，
> 拿掉的理由是它擋不住它宣稱要擋的東西：交易服務的 API 本來就對外開著，
> 想猜密碼的人直接打那裡，不會繞路走這裡。擋住事情的是**登入本身**——
> 見下面〈誰都連得到，為什麼還是安全的〉。

### 自己在本機跑一台

先把 `go-trading` 跑起來（預設 `localhost:8080`），然後：

```bash
go mod download
make start        # 啟動於 :8090，MCP 掛在 /mcp
curl localhost:8090/health   # {"status":"Healthy"}
```

接到助理：

```bash
claude mcp add --scope user --transport http go-trading http://localhost:8090/mcp
```

### 接上之後怎麼用

先登入：在對話裡說一句「幫我登入，帳號是 …」。然後就直接問人話：

> 「BTCUSDT 最近一天每小時的形狀？」
> 「把我那支布林通道在台積電上用 1 小時刻度回測 2026 年，起始資金 100 萬。」
> 「幫我開一台機器人，每 15 分鐘跑一次那份交易策略。」
> 「同一段再跑一次，這次停損放 2%，差很多嗎？」
> 「BTCUSDT 永續這兩天的資金費率是正還是負？大戶偏多還是偏空？」

**登入會過期，但不會打擾你**——外掛自己換一份新的，把原本那件事做完（見〈登入是怎麼回事〉）。
**重啟外掛就要重新登入**：身分只活在行程裡，這是刻意的。

### 接不上的時候

| 看到什麼 | 多半是 |
| :--- | :--- |
| `Dynamic Client Registration rejected (HTTP 404)` | 網址少了結尾的 `/mcp` |
| `MCP endpoint not found` | 同上 |
| 明明改好了、重開又變回舊網址 | 別的 Claude Code session 把 `~/.claude.json` 整份蓋回去了。關掉其他視窗再設定一次 |
| `✔ Connected` 但每件事都說要先登入 | 還沒登入，或外掛重啟過。說一次「幫我登入」 |
| 登入回 **429** | 密碼連錯三次被鎖了，**一週**。訊息裡有可以再試的時間；改密碼也會解鎖 |
| 登入成功但每件事都被擋 | 帳號還沒被放行（`isEnabled`）。放行只發生在系統外面 |

先確認端點本身活著：

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  https://trading-mcp.coding-afternoon.com/mcp \
  -X POST -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}'
# 200 = 端點沒問題，問題在你的設定那一行
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

連接埠**只綁在 `127.0.0.1`**。這不是因為端點危險，而是因為本機跑的這一台是給你自己
改程式用的——它指向的多半是你電腦上那份測試資料，沒有理由讓同一個網段的人碰得到。

| 指令 | 用途 |
| :--- | :--- |
| `make docker-up` | 背景啟動 |
| `make docker-down` | 停掉並移除 |
| `make docker-logs` | 跟著看紀錄 |
| `make docker-build` | 重新編映像檔 |

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

## 誰都連得到，為什麼還是安全的

部署出去的那一台**沒有帳號密碼那一關**，任何人都連得上 `/mcp`。這是想過的，不是漏的。

曾經有過一道 basic auth。拿掉它的理由是**它擋不住它宣稱要擋的東西**：交易服務自己的
API（`trading-api.coding-afternoon.com`）本來就對外開著，`POST /sessions` 全世界都打得到。
想猜密碼的人直接打那裡就好，不會繞路走這個外掛。一道只擋得住自己人的鎖，
唯一的效果是讓人以為擋住了——而它還佔用了 `Authorization` 標頭，
把「自備憑證」那條路一起堵死。

真正擋住事情的有三層，而且它們擋的是**「能做什麼」**而不是**「連不連得到」**：

| | 擋什麼 |
| :--- | :--- |
| **每件能力都需要身分** | 沒登入的連線看得到工具清單，**一件事也做不了** |
| **新帳號預設沒被放進來** | `isEnabled` 預設 `false`，而且**沒有任何一條路可以自己開啟自己**——沒有欄位、沒有 repository method、沒有路由。放人進來只發生在系統外面 |
| **登入會累** | 同一個帳號連續 **3** 次密碼錯誤就鎖 **7** 天，鎖住期間**連正確的密碼也進不來**，而且再試不會把解除時刻往後延 |

所以這裡的門是**平台帳號本身**，不是一組所有人共用的固定密碼。
連得到不等於進得來；進得來不等於能用。

> 被鎖住時交易服務回 **429**，訊息裡說得出什麼時候能再試。
> 沒有自助解鎖——時間到了自己開，或者改密碼（也會一併解鎖）。

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
| `SERVER_BIND_ADDRESS` | `127.0.0.1` | 聽在哪張網卡上。**預設只聽本機**（在 Docker 裡必須是 `0.0.0.0`，見 `docker-compose.yml`） |
| `SERVER_PORT` | `8090` | 外掛自己聽在哪個埠 |
| `MCP_PATH` | `/mcp` | 掛在哪個路徑上 |
| `TRADING_SERVICE_REQUEST_TIMEOUT_SECONDS` | `30` | 單次請求逾時 |
| `TRADING_SERVICE_REPLAY_TIMEOUT_SECONDS` | `120` | 四件重演自己的逾時。**要比交易服務整次重演的允許時間（預設 90 秒）長**，讓交易服務自己先說「沒跑完」，而不是外掛先放棄 |
| `LIVE_UPDATE_WAIT_LIMIT_SECONDS` | `10` | 「看一眼即時更新」最長等幾秒 |
| `IDLE_CONNECTION_TIMEOUT_MINUTES` | `60` | 連線閒置多久就放掉（放掉要重新登入） |

打錯的數字會**退回預設值而不是讓外掛起不來**——一個時間打錯，不該讓整個東西掛掉。

## 即時跟盤只給「看一眼」

`trading_peek_live_k_candle` 收到第一則更新就回，最多等 `LIVE_UPDATE_WAIT_LIMIT_SECONDS` 秒。
**等滿沒收到東西是正常結果，不是錯誤**——凌晨三點的台股本來就是這樣。
要連續看就重複呼叫；外掛不維持一條永遠開著的通道。

## 永續合約是另一條線

交易服務的**永續合約**與現貨從來源到儲存零交集：自己的 K 線、自己的追蹤名單、自己的同步輪次，
另外還有資金費率結算、持倉統計、交易規格、完整的維持保證金分級。外掛把它們接成 **18 件能力**：

- **名字一律帶 `contract`**，而且只問得到合約那一條線；現貨的能力一件都不會問到合約。
  同一個代號在兩邊是兩種商品——問錯邊不會被拒絕，只會拿回一份看起來正常、講的是另一件事的數字。
- 合約 K 線**每一項都必填**（含標記價格、指數價格、溢價指數；溢價指數可以是負的）。
- **完整維持保證金分級要交易服務設定了幣安帳戶金鑰才有**，沒有時回空的——那不是錯誤。
- **加入合約追蹤名單要等二十秒左右**，它當場補齊 K 線、完整資金費率、三十天持倉統計與交易規格。

### 吃合約行情的策略腳本

策略腳本多了一格 **`marketDataKind`**：`kCandle`（現貨 K 線，建立時不給就是它）或 `contractKCandle`（合約行情格）。
**修改時不給就是保留原本的**——這是「整份改寫、沒帶到就變空」的唯一例外；**建立後不得更換**。

吃合約行情的策略腳本用 **`trading_calculate_contract_indicator`** 算（與兩件合約重演一樣需要登入）：
填的東西與 `trading_calculate_indicator` 一模一樣，算式入口改收 `[]indicator.ContractKCandle`。
兩種計算都會拒絕吃另一種行情的策略腳本。要讓它常駐盯盤，就把用它的合約交易策略掛上一台**合約機器人**（見〈合約機器人〉）。

合約行情格的教學（每一個欄位名、沒有值一律為零、延續的資金費率不是每一格都收付）只寫在一處，
建立／修改策略腳本與合約指標計算三件能力讀的是同一段話。

### 合約機器人

策略機器人也多了一格 **`marketDataKind`**：`kCandle`（**現貨機器人**，建立時不給就是它）或 `contractKCandle`（**合約機器人**）。
與策略腳本、交易策略同一條規則：**修改時不給就是保留原本的、建立後不得更換**，外掛不給就不送。

- 現貨機器人只引用吃 K 線的交易策略；合約機器人只引用吃合約行情的交易策略，盯的合約標的**必須已在合約追蹤名單上**。
- 合約機器人的 `positionPlan` 多收 **`leverage`**（不給即一倍）；現貨機器人給大於一倍照舊被拒絕。
  說明裡的形狀範例刻意不含 `leverage`——助理會照抄範例，現貨機器人抄到它就會被拒絕。
- `trading_list_strategy_bots` 可帶 `marketDataKind` 只列一種。
- 兩種共用同一組機器人能力，因為交易服務兩種共用同一組入口；是否合法一律由交易服務判斷。

### 合約重演與合約交易策略

- **`trading_backtest_contract_strategy_script`**（`POST /contract-backtests`）在**逐倉合約帳戶**上重演一支吃合約行情的策略腳本；
  條件是現貨重演那一份，加 `leverage`（留白即一倍）、`slippagePercentage`（留白即不計），以及只有這一件有的 `tradingMode`
  （`longShort`／`longOnly`／`shortOnly`，留白即多空反手）。
- **`trading_backtest_contract_trading_strategy`**（`POST /trading-strategies/{id}/contract-backtests`）重演一份合約交易策略；
  **沒有 `tradingMode` 那一格**——交易模式由那份交易策略自己說。
- 兩件共用一份條件清單（現貨那一份的延伸）與一段合約帳戶說明（`contractAccountReplayNote`）：逐倉、強平看標記價格、
  資金費率一律計入、分級只有今天那一組，以及怎麼讀強平、資金費用、多空分開的勝率、被擋下的開倉。
- 建立／修改交易策略多了 `marketDataKind`（建立後不得更換、修改不給即保留）與 `tradingMode`（只有合約交易策略有）。
- 現貨重演兩件的說明從「這個服務只重演現貨」改成「這兩件只重演現貨，合約的事改用合約重演」。

### 短線重演

四件重演（現貨與合約、策略腳本與交易策略）共用：

- **`fillTiming`**：`close`（預設，說出信號那一格收盤成交）或 `nextOpen`（下一格開盤成交）。說明要助理研發短線時用 `nextOpen`。
- **`validationStartTime`**：切出調參段與驗證段，結果多 `inSample` 與 `validation` 兩份、各自從空手開始。
  說明要助理**只拿調參段調參、只以驗證段判斷**。
- 一段共用說明（`shortTermReplayNote`）教助理讀五格新統計：`profitFactor`、`expectancy`、`averageHoldingSeconds`、
  `maximumConsecutiveLossCount`、`costToGrossProfitRatio`。
- **重演等比較久**（`TRADING_SERVICE_REPLAY_TIMEOUT_SECONDS`），其他能力照舊。
- **成功的重演結果交給助理前會先精簡**（`ReplayResultDomain`）：資金曲線超過 200 點時平均取 200 點（頭尾必留）並加
  `equityCurvePointTotalCount`；交易明細超過 100 筆時只留最近 100 筆並加 `closedTradeTotalCount`；`inSample`／`validation`
  各自照做。**成績單一個數字都不動**；被拒絕的回覆與看不懂的回覆原封轉交。

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
六十幾件事共用同一段流程，差異全部收在這份清單裡。

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
└── tool_catalog*.go     ★ 能力清單——唯一列出六十幾件事的地方

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
