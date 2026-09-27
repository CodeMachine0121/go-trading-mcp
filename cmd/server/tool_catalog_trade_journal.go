package main

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

const contractTradeFilledPriceNote = "實際成交價（字串形式的精確小數）。**必須是使用者說的實際成交價**——" +
	"機器人訊息裡的參考價不是成交價，使用者只貼了機器人訊息、沒說實際成交價時，先問他，不要拿參考價充當"

const contractTradeFilledAtNote = "成交時間（RFC3339，帶時區）。使用者說了時間就照他說的換算；" +
	"**省略即交易服務以現在記下**，回覆會帶出實際記下的時間——請用台北時間（或使用者說的時區）說出來請他確認"

// contractTradeFillParameters are shared by adding and correcting a fill so the two never drift apart.
func contractTradeFillParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("kind", vo.ToolParameterKindString, "entry（進場）或 exit（出場）", true),
		bodyParameter("filledAt", vo.ToolParameterKindString, contractTradeFilledAtNote, false),
		bodyParameter("price", vo.ToolParameterKindString, contractTradeFilledPriceNote, true),
		bodyParameter("quantity", vo.ToolParameterKindString, "成交數量（字串形式的精確小數），必須大於零", true),
		bodyParameter("liquidity", vo.ToolParameterKindString,
			"maker（掛單）或 taker（吃單），決定自動帶出的手續費用哪一個費率", false),
		bodyParameter("fee", vo.ToolParameterKindString,
			"實際手續費（字串形式的精確小數）。省略即依使用者設定的費率自動算；還沒設定費率時記為 0 並標示未設定費率", false),
	}
}

// contractTradePlanParameters are shared by recording a trade and amending its plan.
func contractTradePlanParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("plannedStopLossPrice", vo.ToolParameterKindString,
			"計畫止損價（字串形式的精確小數）。**止損止盈要在對的一邊**：做多止損低於第一筆進場價、止盈高於；做空相反，放錯邊會被拒絕。"+
				"沒有計畫止損就算不出 R 倍數", false),
		bodyParameter("plannedTakeProfitPrice", vo.ToolParameterKindString, "計畫止盈價（字串形式的精確小數）", false),
		bodyParameter("entryReason", vo.ToolParameterKindString, "進場理由", false),
		bodyParameter("confidence", vo.ToolParameterKindInteger, "信心，1 到 5", false),
	}
}

func contractTradeIdentifierParameter() vo.ToolParameterVo {
	return pathParameter("id", "交易編號")
}

func tradeJournalApiTools(replayWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_record_contract_trade",
			"記一筆使用者**自己在交易所手動下**的永續合約交易（系統不下單）。新增後狀態是持倉中，回覆帶出編號、進場均價、計畫風險與記下的成交時間。"+
				"\n\n**成交價必須是使用者說的實際成交價**：機器人訊息的參考價不是成交價，沒說就先問。"+
				"\n\n**同一標的、同一方向已有持倉中會被拒絕**，回覆會給出那一筆的編號——改用 trading_add_contract_trade_fill 在那一筆加成交。"+
				"一句話說了多筆成交時：第一筆用這一支新增，其餘依序用 trading_add_contract_trade_fill 一筆一筆送，中途被拒就停下並帶回原因。"+
				"\n\n這一筆一律屬於授權給外掛的那位使用者，**沒有欄位可以指定擁有者**。",
			vo.RequestVerbSubmit, "/contract-trade-records",
			[]vo.ToolParameterVo{
				bodyParameter("symbol", vo.ToolParameterKindString, "合約標的，如 BTCUSDT；不認得的會被拒絕", true),
				bodyParameter("direction", vo.ToolParameterKindString, "long（做多）或 short（做空）；平多、平空不是方向", true),
				bodyParameter("leverage", vo.ToolParameterKindString,
					"槓桿倍數（字串形式的精確小數）。**留白即一倍**，小於一會被拒絕", false),
				bodyParameter("firstEntryFill", vo.ToolParameterKindObject,
					"第一筆進場成交，形狀：{\"filledAt\":\"2026-09-27T14:03:00+08:00\", \"price\":\"97905\", \"quantity\":\"0.03\", "+
						"\"liquidity\":\"taker\", \"fee\":\"1.47\"}。price 與 quantity 必填。"+
						"price："+contractTradeFilledPriceNote+"。filledAt："+contractTradeFilledAtNote+
						"。fee 省略即依使用者的費率自動算", true),
				bodyParameter("tradingStrategyId", vo.ToolParameterKindInteger,
					"這筆是依哪一份交易策略做的。**只能是使用者自己的合約交易策略**（吃合約行情的那種），K 線交易策略會被拒絕；不給即自行判斷", false),
				bodyParameter("setupTagIds", vo.ToolParameterKindArray,
					"型態標籤編號（整數陣列）。不知道編號時先用 trading_list_trade_tags", false),
				bodyParameter("plan", vo.ToolParameterKindObject,
					"進場計畫，皆選填，形狀：{\"plannedStopLossPrice\":\"96380\", \"plannedTakeProfitPrice\":\"100785\", "+
						"\"entryReason\":\"突破前高\", \"confidence\":3}。價格為字串形式的精確小數，信心 1 到 5。"+
						"**止損止盈要在對的一邊**：做多止損低於第一筆進場價、止盈高於；做空相反，放錯邊會被拒絕。沒有計畫止損就算不出 R 倍數", false),
			}...,
		),
		domains.NewApiToolDomain(
			"trading_add_contract_trade_fill",
			"對一筆持倉中的交易加一筆進場（加碼）或出場（減碼、平倉）成交。"+
				"\n\n**出場超過目前持倉會被拒絕**——要反手請先平倉，再用 trading_record_contract_trade 新增一筆反方向的交易。"+
				"持倉正好歸零即變成已平倉，回覆後提醒使用者可以寫檢討。**已平倉的交易不能再加成交**。",
			vo.RequestVerbSubmit, "/contract-trade-records/{id}/fills",
			append([]vo.ToolParameterVo{contractTradeIdentifierParameter()}, contractTradeFillParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_update_contract_trade_fill",
			"修正一筆記錯的成交（整筆取代）。**只有持倉中的交易可以修正**；平倉後成交已鎖定，只能加附註或刪除整筆重記。",
			vo.RequestVerbReplace, "/contract-trade-records/{id}/fills/{fillId}",
			append([]vo.ToolParameterVo{
				contractTradeIdentifierParameter(),
				pathParameter("fillId", "成交編號"),
			}, contractTradeFillParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_remove_contract_trade_fill",
			"刪除一筆記錯的成交。**持倉中才可刪**；不能刪到沒有進場成交——要整筆放棄請用 trading_delete_contract_trade。",
			vo.RequestVerbRemove, "/contract-trade-records/{id}/fills/{fillId}",
			contractTradeIdentifierParameter(),
			pathParameter("fillId", "成交編號"),
		),
		domains.NewApiToolDomain(
			"trading_update_contract_trade_plan",
			"修改進場計畫（止損、止盈、理由、信心）。**平倉後計畫已鎖定**，被拒時建議改用 trading_add_contract_trade_note 加附註。",
			vo.RequestVerbReplace, "/contract-trade-records/{id}/plan",
			append([]vo.ToolParameterVo{contractTradeIdentifierParameter()}, contractTradePlanParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_add_contract_trade_note",
			"在一筆交易上加一則附註。任何狀態都能加；**附註不會改動原本的計畫**，平倉後要更正計畫就用這一支。",
			vo.RequestVerbSubmit, "/contract-trade-records/{id}/notes",
			contractTradeIdentifierParameter(),
			bodyParameter("content", vo.ToolParameterKindString, "附註內容", true),
		),
		domains.NewApiToolDomain(
			"trading_write_contract_trade_review",
			"寫或修改一筆交易的檢討。**只能在平倉後寫**，持倉中會被拒絕；寫完即已檢討，之後還能改。",
			vo.RequestVerbReplace, "/contract-trade-records/{id}/review",
			contractTradeIdentifierParameter(),
			bodyParameter("wentWell", vo.ToolParameterKindString, "哪裡做對", false),
			bodyParameter("wentWrong", vo.ToolParameterKindString, "哪裡做錯", false),
			bodyParameter("nextTime", vo.ToolParameterKindString, "下次怎麼做", false),
			bodyParameter("executionScore", vo.ToolParameterKindInteger, "執行評分，1 到 5", true),
			bodyParameter("mistakeTagIds", vo.ToolParameterKindArray,
				"失誤標籤編號（整數陣列）。不知道編號時先用 trading_list_trade_tags", false),
		),
		domains.NewApiToolDomain(
			"trading_set_contract_trade_setup_tags",
			"設定一筆交易的型態標籤，**整組取代**；空陣列即全部拿掉。",
			vo.RequestVerbReplace, "/contract-trade-records/{id}/setup-tags",
			contractTradeIdentifierParameter(),
			bodyParameter("setupTagIds", vo.ToolParameterKindArray, "型態標籤編號（整數陣列）", true),
		),
		domains.NewApiToolDomain(
			"trading_list_contract_trades",
			"列出使用者的合約交易，依第一筆進場時間由新到舊。**省略 limit 即最近 20 筆**，使用者指定筆數才填；回覆時說出總數。"+
				"「待檢討」就是 status=closed。",
			vo.RequestVerbRead, "/contract-trade-records",
			queryParameter("status", vo.ToolParameterKindString, "open（持倉中）、closed（已平倉、待檢討）或 reviewed（已檢討）", false),
			queryParameter("symbol", vo.ToolParameterKindString, "只列這個合約標的", false),
			queryParameter("period", vo.ToolParameterKindString,
				"7d、30d、90d 或 all，以第一筆進場時間篩選；**省略即不限期間**", false),
			queryParameter("limit", vo.ToolParameterKindInteger, "最多幾筆，最多 200", false),
		),
		domains.NewApiToolDomain(
			"trading_get_contract_trade",
			"查看一筆交易：成交、計畫、附註、檢討，以及交易服務算好的損益、R 倍數、資金費用、最大不利／最大有利、進場滑點。"+
				"\n\n回覆裡「算不出／無法計算／估算」的項目照交易服務的說法帶回，**不要說成 0**。別人的交易會是找不到。",
			vo.RequestVerbRead, "/contract-trade-records/{id}",
			contractTradeIdentifierParameter(),
		),
		domains.NewApiToolDomain(
			"trading_delete_contract_trade",
			"刪除一筆交易，成交、附註、檢討會一併刪除。"+
				"\n\n**刪除前必須先向使用者確認是哪一筆**（說出標的、方向、編號）；使用者說「刪掉 BTC 那筆」但有多筆時，先列出請他指定，不要猜。",
			vo.RequestVerbRemove, "/contract-trade-records/{id}",
			contractTradeIdentifierParameter(),
		),
		domains.NewApiToolDomain(
			"trading_get_contract_trade_statistics",
			"看一段期間的績效統計：勝率、平均 R、獲利因子、失誤成本、有關聯策略 vs 自行判斷等。只計期間內平倉的交易。"+
				"\n\n沒有已平倉交易時勝率是不適用，**不是 0%**；沒設止損的交易不計入 R 相關數字。",
			vo.RequestVerbRead, "/contract-trade-records/statistics",
			queryParameter("period", vo.ToolParameterKindString, "7d、30d、90d 或 all；**省略即最近 30 天**", false),
		),
		domains.NewApiToolDomain(
			"trading_compare_contract_trades_with_backtest",
			"拿一份合約交易策略的已平倉實單，對照同一段期間的重演。**會替每個標的重演一次，需要時間**。"+
				"某一列重演失敗時實盤照常、回測欄說出原因；策略已刪除時說無法重演。",
			vo.RequestVerbRead, "/trading-strategies/{id}/contract-trade-comparison",
			pathParameter("id", "合約交易策略識別碼"),
		).Waiting(replayWaitLimit),
		domains.NewApiToolDomain(
			"trading_get_trade_journal_settings",
			"查看使用者的掛單與吃單手續費率；沒設定時說未設定。",
			vo.RequestVerbRead, "/users/me/trade-journal-settings",
		),
		domains.NewApiToolDomain(
			"trading_save_trade_journal_settings",
			"設定或修正掛單與吃單手續費率。**不得為負**；改了只影響之後記的成交，舊成交的手續費不變。",
			vo.RequestVerbReplace, "/users/me/trade-journal-settings",
			bodyParameter("makerFeeRate", vo.ToolParameterKindString, "掛單費率，百分比（0.02 就是 0.02%，字串形式的精確小數）", false),
			bodyParameter("takerFeeRate", vo.ToolParameterKindString, "吃單費率，百分比（0.05 就是 0.05%，字串形式的精確小數）", false),
		),
		domains.NewApiToolDomain(
			"trading_list_trade_tags",
			"列出使用者的失誤標籤與型態標籤。一開始就有五個預設失誤標籤。記交易或寫檢討需要標籤編號時先用這一支。",
			vo.RequestVerbRead, "/users/me/trade-tags",
		),
		domains.NewApiToolDomain(
			"trading_create_trade_tag",
			"新增一個標籤。**同一類不能重名**，不同類可以。",
			vo.RequestVerbSubmit, "/users/me/trade-tags",
			bodyParameter("kind", vo.ToolParameterKindString, "mistake（失誤標籤）或 setup（型態標籤）", true),
			bodyParameter("name", vo.ToolParameterKindString, "標籤名稱", true),
		),
		domains.NewApiToolDomain(
			"trading_rename_trade_tag",
			"標籤改名，貼著它的交易跟著改名。",
			vo.RequestVerbReplace, "/users/me/trade-tags/{id}",
			pathParameter("id", "標籤編號"),
			bodyParameter("name", vo.ToolParameterKindString, "新名稱", true),
		),
		domains.NewApiToolDomain(
			"trading_delete_trade_tag",
			"刪除一個標籤。**還貼在交易上的標籤不能刪**，被拒時說出還有幾筆，建議先從交易上移除或改名。",
			vo.RequestVerbRemove, "/users/me/trade-tags/{id}",
			pathParameter("id", "標籤編號"),
		),
	}
}
