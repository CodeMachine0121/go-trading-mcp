# go-trading-mcp

讓 **AI 助理**代你操作 [`go-trading`](../go-trading) 的一層外掛（MCP server，純 Go）。

交易服務會的每一件事——查行情、算指標、寫策略腳本、組交易策略、重演一段行情、
養策略機器人、推通知——都成為助理呼叫得到的一個**能力**，目前 **65 件**。

在 Claude Code 的 `/mcp` 選單按一次連線、在瀏覽器登入並按「允許」，之後說人話就能操作自己的東西。
**密碼從不經過對話**，授權過期時 Claude Code 會自己換新。

**已經有一台跑在 `https://trading-mcp.coding-afternoon.com/mcp`**，不必自己架：

```bash
claude mcp add --scope user --transport http \
  go-trading https://trading-mcp.coding-afternoon.com/mcp
```

（`/mcp` 與 `--scope user` 兩樣都不能省，理由見〈快速開始〉。）

## 它做什麼、不做什麼

| | |
| :--- | :--- |
| **做** | 確認每一次呼叫帶來的外掛授權、帶著它把助理的要求轉達給交易服務、把交易服務的回答（含拒絕的原話）帶回來 |
| **不做** | 不自己存行情、不自己算指標、不下單、**不改變任何一條業務規則** |
| **刻意不做** | **不代理交易服務內建的行情對話助手**——見下 |

交易服務拒絕的，外掛不放行；交易服務允許的，外掛不多問。
**它沒有自己的資料庫，也不保管任何人的身分**——所以重啟不必重新授權，也可以同時跑不只一份。

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

第一次呼叫時外掛會回 401 並指出授權說明，Claude Code 看到後會在 `/mcp` 選單把
`go-trading` 標成需要授權：選它、按 Authenticate，瀏覽器會打開交易服務的網站，登入並按「允許」即可。

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

授權一次之後就直接問人話：

> 「BTCUSDT 最近一天每小時的形狀？」
> 「把我那支布林通道在台積電上用 1 小時刻度回測 2026 年，起始資金 100 萬。」
> 「幫我開一台機器人，每 15 分鐘跑一次那份交易策略。」
> 「同一段再跑一次，這次停損放 2%，差很多嗎？」
> 「BTCUSDT 永續這兩天的資金費率是正還是負？大戶偏多還是偏空？」

建立帳號與登入都在交易服務的網站上做；外掛不會、也不該向你要密碼。

### 接不上的時候

| 看到什麼 | 多半是 |
| :--- | :--- |
| `Dynamic Client Registration rejected (HTTP 404)` | 網址少了結尾的 `/mcp` |
| `MCP endpoint not found` | 同上 |
| 明明改好了、重開又變回舊網址 | 別的 Claude Code session 把 `~/.claude.json` 整份蓋回去了。關掉其他視窗再設定一次 |
| `/mcp` 裡顯示需要授權 | 還沒授權，或授權被撤銷。選 `go-trading` → Authenticate，在瀏覽器登入並允許 |
| 回覆說「請到 Claude Code 的 /mcp 選單重新連線」 | 交易服務不認得這份授權了。照做一次 |
| 瀏覽器登入回 **429** | 密碼連錯三次被鎖了，**一週**。訊息裡有可以再試的時間；改密碼也會解鎖 |
| 授權成功但每件事都被擋 | 帳號還沒被放行（`isEnabled`）。放行只發生在系統外面 |

先確認端點本身活著：

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  https://trading-mcp.coding-afternoon.com/mcp \
  -X POST -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}'
# 401 且帶 WWW-Authenticate = 端點沒問題，接下來交給 Claude Code 的授權流程
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

## 授權是怎麼回事

外掛是一個 **OAuth 受保護資源**，交易服務是**授權伺服器**（外掛授權）：

1. 沒帶授權的呼叫一律回 401，`WWW-Authenticate` 指向
   `/.well-known/oauth-protected-resource/mcp`，裡面寫著這個外掛的正式位址（`PUBLIC_BASE_URL` + `MCP_PATH`）
   與授權伺服器（`TRADING_SERVICE_PUBLIC_URL`）。
2. Claude Code 自己向交易服務註冊、打開瀏覽器；你在交易服務的網站登入並按「允許」。
3. 之後每一次呼叫都帶著那份外掛授權。外掛向交易服務確認它**有效、使用者還在、而且是發給這個外掛的**，
   以授權的 SHA-256 指紋記住結果最多一分鐘（不超過它自己的到期時刻），然後帶著同一份授權去代辦。
4. 過期由 Claude Code 自己向交易服務換新，外掛不參與。

- **外掛不經手密碼、不保管任何身分**，授權也不會出現在回答或紀錄裡。
- **可以同時跑不只一份。** 連線（`Mcp-Session-Id`）仍然存在，但不再承載身分；
  落到另一份時 Claude Code 重開一段連線即可，不必重新授權。
- **整個 `/mcp` 都要授權**，沒有匿名能力；只有 `/health` 與授權說明不需要。
- 交易服務暫時不在時回的是「連不到交易服務」，不會叫你重新授權。

## 做不成的時候，它會說清楚是哪一種

這幾句話絕不混用，因為它們要你做的事完全不同：

| 回覆 | 意思 | 你該做什麼 |
| :--- | :--- | :--- |
| `請到 Claude Code 的 /mcp 選單重新連線…` | 交易服務不認得這份外掛授權 | 在 `/mcp` 選單重新授權 |
| `連不到交易服務` | **你送的沒有錯**，是它現在不在 | 等一下，送同一件事 |
| （交易服務的原話） | 規則不通過、不是你的、額度用完… | 照它說的改一改再送 |

**拒絕的原因一字不改地帶回來**，助理才知道該把回溯天數改成多少、該去哪裡拿權限。

## 設定

全部有預設值，`.env` 可整份省略。見 `.env.example`。

| 變數 | 預設值 | 用途 |
| :--- | :--- | :--- |
| `TRADING_SERVICE_BASE_URL` | `http://localhost:8080` | 外掛自己連交易服務用的位址（可以是內部位址）。**填錯這一個，其餘全部不會動** |
| `TRADING_SERVICE_PUBLIC_URL` | `http://localhost:8080` | 交易服務對外的位址，寫進授權說明當作授權伺服器（正式環境 `https://trading-api.coding-afternoon.com`） |
| `PUBLIC_BASE_URL` | `http://localhost:8090` | 外掛對外的位址（不含路徑）。外掛授權必須發給它 + `MCP_PATH`；**由設定給、不從請求推算**，因為前面的轉送會改寫協定 |
| `SERVER_BIND_ADDRESS` | `127.0.0.1` | 聽在哪張網卡上。**預設只聽本機**（在 Docker 裡必須是 `0.0.0.0`，見 `docker-compose.yml`） |
| `SERVER_PORT` | `8090` | 外掛自己聽在哪個埠 |
| `MCP_PATH` | `/mcp` | 掛在哪個路徑上 |
| `TRADING_SERVICE_REQUEST_TIMEOUT_SECONDS` | `30` | 單次請求逾時 |
| `TRADING_SERVICE_REPLAY_TIMEOUT_SECONDS` | `120` | 四件重演自己的逾時。**要比交易服務整次重演的允許時間（預設 90 秒）長**，讓交易服務自己先說「沒跑完」，而不是外掛先放棄 |
| `LIVE_UPDATE_WAIT_LIMIT_SECONDS` | `10` | 「看一眼即時更新」最長等幾秒 |
| `IDLE_CONNECTION_TIMEOUT_MINUTES` | `60` | 連線閒置多久就放掉（Claude Code 會自己重開，不必重新授權） |

打錯的數字會**退回預設值而不是讓外掛起不來**——一個時間打錯，不該讓整個東西掛掉。

## 即時跟盤只給「看一眼」

`trading_peek_live_k_candle` 收到第一則更新就回，最多等 `LIVE_UPDATE_WAIT_LIMIT_SECONDS` 秒。
**等滿沒收到東西是正常結果，不是錯誤**——凌晨三點的台股本來就是這樣。
要連續看就重複呼叫；外掛不維持一條永遠開著的通道。

## 永續合約是另一條線

交易服務的**永續合約**與現貨從來源到儲存零交集：自己的 K 線、自己的追蹤名單、自己的同步輪次，
另外還有資金費率結算、持倉統計、交易規格、完整的維持保證金分級。外掛把它們接成 **19 件能力**：

- **名字一律帶 `contract`**，而且只問得到合約那一條線；現貨的能力一件都不會問到合約。
  同一個代號在兩邊是兩種商品——問錯邊不會被拒絕，只會拿回一份看起來正常、講的是另一件事的數字。
- 合約 K 線**每一項都必填**（含標記價格、指數價格、溢價指數；溢價指數可以是負的）。
- **完整維持保證金分級要交易服務設定了幣安帳戶金鑰才有**，沒有時回空的——那不是錯誤。
- **加入合約追蹤名單要等二十秒左右**，它當場補齊 K 線、完整資金費率、三十天持倉統計與交易規格。
- **`trading_peek_live_contract_k_candle`** 看一眼合約標的即時的那一根（與現貨的看一眼同一個等待上限）：內容是最新價、不含標記價格；走完的那一根由每分鐘那一輪存入；只看得到合約追蹤名單上的合約標的。

### 吃合約行情的策略腳本

策略腳本多了一格 **`marketDataKind`**：`kCandle`（現貨 K 線，建立時不給就是它）或 `contractKCandle`（合約行情格）。
**修改時不給就是保留原本的**——這是「整份改寫、沒帶到就變空」的唯一例外；**建立後不得更換**。

吃合約行情的策略腳本用 **`trading_calculate_contract_indicator`** 算：
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
- 合約機器人的建議部位**照交易所規則算**（數量步進取整、價格跳動單位、預估強平價、資金費率估算，交易所不收時說出原因），
  說明照交易服務的規則寫，外掛不算任何一個數字。
- 合約機器人的執行紀錄多帶 `suggestedDirection`、`suggestedLeverage`、`suggestedNotional`（只在有建議的那一輪；現貨沒有），
  外掛原樣轉交，讀輪次的兩個能力說明裡寫出它們的意思。

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
    vo.RequestVerbSubmit, "/the/new/{id}",
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
│   │   ├── domains/     行為所在地：一件能力怎麼變成一個請求、一份外掛授權算不算數
│   │   ├── dto/         對 application 的回傳形狀
│   │   └── vo/          不可變值
│   ├── service/         代辦一件事、確認外掛授權
│   └── interface/       對外介面（mocks/ 放產生的 mock）
└── infrastructure/
    ├── tradingservice/  ★ 唯一知道 HTTP 存在的地方（含即時跟盤那條線）
    ├── persistence/     外掛授權確認結果的短暫記憶（只記指紋）
    └── clock/           系統時鐘
```

依賴方向一律指向 `domain/`。規範見 [`.claude/rules/`](.claude/rules/)，入口是 [`CLAUDE.md`](CLAUDE.md)。
需求與設計見 [`.sdd/`](.sdd/)。
