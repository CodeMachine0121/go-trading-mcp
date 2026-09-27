package main

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

const spotTradePriceNote = "實際買進或賣出價（字串形式的精確小數）。**必須是使用者說的實際價格**——" +
	"機器人訊息裡的參考價不是成交價，使用者只貼了機器人訊息、沒說實際價格時，先問他，不要拿參考價充當"

const spotTradeQuantityNote = "數量（字串形式的精確小數），必須大於零。**台股以股計、必須是整數**：" +
	"使用者說「一張」即 1,000 股，換算後說出股數請他確認；說零股就照股數。加密貨幣可以有小數"

const spotTradeFeeNote = "這一筆的手續費（字串形式的精確小數）。現貨沒有手續費率，**沒說即 0**；" +
	"台股賣出的證交稅要併入手續費一起填——使用者記台股賣出沒提手續費時，提醒他證交稅要併入"

const spotTradePlanSideNote = "**止損必須低於第一筆買進價、止盈必須高於**，放錯邊會被拒絕。沒有計畫止損就算不出 R 倍數，報酬率照算"

const spotTradeOnlyBuyThenSellNote = "**現貨只有先買後賣，沒有做空、沒有槓桿**：使用者說做空或幾倍時，說明現貨只有先買後賣，不要送出。"

const spotTradeAmbiguousSymbolNote = "**同一個代號可能同時是加密貨幣現貨與合約**（例如 BTCUSDT）：使用者沒說是哪一本時先問現貨還是合約，得到回答前不要送出。"

// spotTradeFillParameters are shared by adding and correcting a buy or sell so the two never drift apart.
func spotTradeFillParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("kind", vo.ToolParameterKindString, "buy（買進）或 sell（賣出）", true),
		bodyParameter("filledAt", vo.ToolParameterKindString, tradeJournalRecordedAtNote, false),
		bodyParameter("price", vo.ToolParameterKindString, spotTradePriceNote, true),
		bodyParameter("quantity", vo.ToolParameterKindString, spotTradeQuantityNote, true),
		bodyParameter("fee", vo.ToolParameterKindString, spotTradeFeeNote, false),
	}
}

func spotTradeIdentifierParameter() vo.ToolParameterVo {
	return pathParameter("id", "現貨交易編號")
}

func spotTradeJournalApiTools(replayWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_record_spot_trade",
			"記一筆使用者**自己在交易所或券商手動買進**的現貨交易（加密貨幣現貨或台股，系統不下單）。新增後狀態是持有中，回覆帶出編號與記下的時間；"+
				"台股一張換算成股數時把股數說出來請使用者確認。"+
				"\n\n"+spotTradeOnlyBuyThenSellNote+
				"\n\n"+spotTradeAmbiguousSymbolNote+
				"\n\n**價格必須是使用者說的實際價格**：機器人訊息的參考價不是成交價，沒說就先問。"+
				"\n\n**同一標的已有持有中會被拒絕**，回覆會給出那一筆的編號——改用 trading_add_spot_trade_fill 在那一筆加買進。"+
				"\n\n這一筆一律屬於授權給外掛的那位使用者，**沒有欄位可以指定擁有者**。",
			vo.RequestVerbSubmit, "/spot-trade-records",
			bodyParameter("symbol", vo.ToolParameterKindString, "現貨標的，如 2330（台股）或 BTCUSDT（加密貨幣現貨）；不認得的會被拒絕", true),
			bodyParameter("firstBuyFill", vo.ToolParameterKindObject,
				"第一筆買進，形狀：{\"kind\":\"buy\", \"filledAt\":\"2026-09-27T10:15:00+08:00\", \"price\":\"1050\", \"quantity\":\"1000\", \"fee\":\"1496\"}。"+
					"price 與 quantity 必填；kind 省略即買進。price："+spotTradePriceNote+"。quantity："+spotTradeQuantityNote+
					"。filledAt："+tradeJournalRecordedAtNote+"。fee："+spotTradeFeeNote, true),
			bodyParameter("tradingStrategyId", vo.ToolParameterKindInteger,
				"這筆是依哪一份交易策略做的。**只能是使用者自己的 K 線（現貨）交易策略**，合約交易策略會被拒絕；不給即自行判斷", false),
			bodyParameter("setupTagIds", vo.ToolParameterKindArray,
				"型態標籤編號（整數陣列），與合約日誌共用同一組標籤。不知道編號時先用 trading_list_trade_tags", false),
			bodyParameter("plan", vo.ToolParameterKindObject,
				"進場計畫，皆選填，形狀：{\"plannedStopLossPrice\":\"1000\", \"plannedTakeProfitPrice\":\"1150\", "+
					"\"entryReason\":\"突破前高\", \"confidence\":3}。價格為字串形式的精確小數，信心 1 到 5。"+spotTradePlanSideNote, false),
		),
		domains.NewApiToolDomain(
			"trading_add_spot_trade_fill",
			"對一筆持有中的現貨交易加一筆買進（加碼）或賣出（分批賣出、全部賣出）。"+
				"\n\n**賣出超過目前持有會被拒絕**。全部賣出即變成已平倉，回覆淨損益與報酬率，並提醒使用者可以寫檢討。**已平倉的交易不能再加買賣**。"+
				"一句話說了多筆時依序一筆一筆送，中途被拒就停下並帶回原因。",
			vo.RequestVerbSubmit, "/spot-trade-records/{id}/fills",
			append([]vo.ToolParameterVo{spotTradeIdentifierParameter()}, spotTradeFillParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_update_spot_trade_fill",
			"修正一筆記錯的買進或賣出。**整筆取代，沒給的項目不會保留**：省略時間會變成現在、省略手續費會變成 0——"+
				"只改其中一項時，先用 trading_get_spot_trade 讀出這一筆，其餘原樣帶上。"+
				"**只有持有中的交易可以修正**；平倉後已鎖定，只能加附註或刪除整筆重記。",
			vo.RequestVerbReplace, "/spot-trade-records/{id}/fills/{fillId}",
			append([]vo.ToolParameterVo{
				spotTradeIdentifierParameter(),
				pathParameter("fillId", "買賣編號"),
			}, spotTradeFillParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_remove_spot_trade_fill",
			"刪除一筆記錯的買進或賣出。**持有中才可刪**；不能刪到沒有買進——要整筆放棄請用 trading_delete_spot_trade。",
			vo.RequestVerbRemove, "/spot-trade-records/{id}/fills/{fillId}",
			spotTradeIdentifierParameter(),
			pathParameter("fillId", "買賣編號"),
		),
		domains.NewApiToolDomain(
			"trading_update_spot_trade_plan",
			"修改現貨交易的進場計畫（止損、止盈、理由、信心）。**整份取代，沒給的項目會被清空**——只改其中一項時，先用 trading_get_spot_trade 讀出現在的計畫，其餘原樣帶上。"+
				"**平倉後計畫已鎖定**，被拒時建議改用 trading_add_spot_trade_note 加附註。",
			vo.RequestVerbReplace, "/spot-trade-records/{id}/plan",
			spotTradeIdentifierParameter(),
			bodyParameter("plannedStopLossPrice", vo.ToolParameterKindString,
				"計畫止損價（字串形式的精確小數）。"+spotTradePlanSideNote, false),
			bodyParameter("plannedTakeProfitPrice", vo.ToolParameterKindString, "計畫止盈價（字串形式的精確小數）", false),
			bodyParameter("entryReason", vo.ToolParameterKindString, "進場理由", false),
			bodyParameter("confidence", vo.ToolParameterKindInteger, "信心，1 到 5", false),
		),
		domains.NewApiToolDomain(
			"trading_add_spot_trade_note",
			"在一筆現貨交易上加一則附註。任何狀態都能加；**附註不會改動原本的計畫**，平倉後要更正計畫就用這一支。",
			vo.RequestVerbSubmit, "/spot-trade-records/{id}/notes",
			spotTradeIdentifierParameter(),
			bodyParameter("content", vo.ToolParameterKindString, "附註內容", true),
		),
		domains.NewApiToolDomain(
			"trading_write_spot_trade_review",
			"寫或修改一筆現貨交易的檢討。**只能在平倉後寫**，持有中會被拒絕；寫完即已檢討，之後還能改。"+
				"**整份取代，沒給的項目會被清空**（含失誤標籤）——修改時先用 trading_get_spot_trade 讀出現在的檢討，其餘原樣帶上。"+
				"失誤標籤與合約日誌共用同一組。",
			vo.RequestVerbReplace, "/spot-trade-records/{id}/review",
			spotTradeIdentifierParameter(),
			bodyParameter("wentWell", vo.ToolParameterKindString, "哪裡做對", false),
			bodyParameter("wentWrong", vo.ToolParameterKindString, "哪裡做錯", false),
			bodyParameter("nextTime", vo.ToolParameterKindString, "下次怎麼做", false),
			bodyParameter("executionScore", vo.ToolParameterKindInteger, "執行評分，1 到 5", true),
			bodyParameter("mistakeTagIds", vo.ToolParameterKindArray,
				"失誤標籤編號（整數陣列）。不知道編號時先用 trading_list_trade_tags", false),
		),
		domains.NewApiToolDomain(
			"trading_set_spot_trade_setup_tags",
			"設定一筆現貨交易的型態標籤，**整組取代**；空陣列即全部拿掉。",
			vo.RequestVerbReplace, "/spot-trade-records/{id}/setup-tags",
			spotTradeIdentifierParameter(),
			bodyParameter("setupTagIds", vo.ToolParameterKindArray, "型態標籤編號（整數陣列）", true),
		),
		domains.NewApiToolDomain(
			"trading_list_spot_trades",
			"列出使用者的現貨交易，依第一筆買進時間由新到舊。**省略 limit 即最近 20 筆**，使用者指定筆數才填；回覆時說出總數。"+
				"「待檢討」就是 status=closed；「我的台股交易」就是 market=taiwanStock。",
			vo.RequestVerbRead, "/spot-trade-records",
			queryParameter("status", vo.ToolParameterKindString, "open（持有中）、closed（已平倉、待檢討）或 reviewed（已檢討）", false),
			queryParameter("symbol", vo.ToolParameterKindString, "只列這個現貨標的", false),
			queryParameter("market", vo.ToolParameterKindString, "taiwanStock（台股）或 crypto（加密貨幣現貨）", false),
			queryParameter("period", vo.ToolParameterKindString,
				"7d、30d、90d 或 all，以第一筆買進時間篩選；**省略即不限期間**", false),
			queryParameter("limit", vo.ToolParameterKindInteger, "最多幾筆，最多 200", false),
		),
		domains.NewApiToolDomain(
			"trading_get_spot_trade",
			"查看一筆現貨交易：買賣、計畫、附註、檢討，以及交易服務算好的淨損益、報酬率、R 倍數、最大不利／最大有利、浮動損益、進場滑點。"+
				"**現貨沒有資金費用與強平價**，不要提。"+
				"\n\n回覆裡「算不出／無法計算／估算」的項目照交易服務的說法帶回，**不要說成 0**；浮動損益是以最新價估算。別人的交易會是找不到。",
			vo.RequestVerbRead, "/spot-trade-records/{id}",
			spotTradeIdentifierParameter(),
		),
		domains.NewApiToolDomain(
			"trading_delete_spot_trade",
			"刪除一筆現貨交易，買賣、附註、檢討會一併刪除。"+
				"\n\n**刪除前必須先向使用者確認是哪一筆**（說出標的、編號）；使用者說「刪掉 2330 那筆」但有多筆時，先列出請他指定，不要猜。",
			vo.RequestVerbRemove, "/spot-trade-records/{id}",
			spotTradeIdentifierParameter(),
		),
		domains.NewApiToolDomain(
			"trading_get_spot_trade_statistics",
			"看一段期間的現貨績效統計。**台股與加密貨幣分成兩組**，各自有勝率、平均報酬率、獲利因子、累積損益、失誤成本、有關聯策略 vs 自行判斷；"+
				"**金額不跨幣別加總**（台股是新台幣、加密貨幣是 USDT）。只計期間內平倉的交易。"+
				"\n\n平均 R 只用有計畫止損的交易算，回覆時說出以幾筆計；沒有已平倉交易時勝率是不適用，**不是 0%**。",
			vo.RequestVerbRead, "/spot-trade-records/statistics",
			queryParameter("period", vo.ToolParameterKindString, "7d、30d、90d 或 all；**省略即最近 30 天**", false),
		),
		domains.NewApiToolDomain(
			"trading_compare_spot_trades_with_backtest",
			"拿一份 K 線（現貨）交易策略的已平倉現貨實單，對照同一段期間的現貨重演，並排平倉筆數與勝率。**重演不計成本**（現貨沒有手續費率），回覆時說出來。"+
				"**會替每個標的重演一次，需要時間**。某一列重演失敗時實盤照常、回測欄說出原因；策略已刪除時說無法重演。"+
				"標的多時可能超過等待上限而被當成連不到交易服務——**這時不要立刻重送**（會再重演一輪），請使用者稍後再試。",
			vo.RequestVerbRead, "/trading-strategies/{id}/spot-trade-comparison",
			pathParameter("id", "K 線交易策略識別碼"),
		).Waiting(replayWaitLimit),
	}
}
