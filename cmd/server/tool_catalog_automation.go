package main

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// strategyBotWriteParameters are what a strategy bot is made of. Shared by creating
// one and rewriting one.
//
// The position plan arrives nested rather than as four parallel boxes, the way the two
// condition trees do. Those four figures only mean anything together: flat, an
// assistant could send a stop distance with no capital — a combination nothing
// refuses, and after which nothing happens.
//
// That combination is still not refused: it reads as no plan at all, the trading
// service accepts it, and the cost lands on whoever reads the message. So the
// description is still the only guard over it.
func strategyBotWriteParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("name", vo.ToolParameterKindString, "這台機器人叫什麼", true),
		bodyParameter("symbol", vo.ToolParameterKindString, "它盯著哪一個交易標的", true),
		bodyParameter("tradingStrategyId", vo.ToolParameterKindInteger,
			"它照哪一份交易策略做決定。**只能是吃 K 線的交易策略**——策略機器人目前只跑 K 線，"+
				"指名一份吃合約行情的會被拒絕", true),
		bodyParameter("triggerIntervalMinutes", vo.ToolParameterKindInteger,
			"每幾分鐘跑一輪", true),
		bodyParameter("positionPlan", vo.ToolParameterKindObject,
			"這台機器人每一輪要建議押多少、停在哪裡。**整組可以不給**——"+
				"不給的機器人只報方向，不報數字。形狀："+
				"{\"capital\":\"50000\", \"sizingMode\":\"percentage\", \"sizingValue\":\"10\", "+
				"\"stopLossPercentage\":\"3\", \"takeProfitPercentage\":\"5\"}。"+
				"capital 是這一組的開關：不給它，其餘三樣填了也不算。"+
				"金額一律以字串給精確小數。"+
				"sizingMode 三選一：allIn 全押（不必給 sizingValue）、percentage 押資金的百分之幾、"+
				"fixedAmount 每次押固定金額；不給即 allIn。"+
				"**沒有槓桿這一項**——一台機器人建議得了的，就是重演驗證得了的，"+
				"而重演借不到錢（重演那兩支工具的說明講得完整）。"+
				"stopLossPercentage 與 takeProfitPercentage 是百分點（3 就是 3%），各自可以單獨不給", false),
	}
}

func strategyBotApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_strategy_bot",
			"建立一台策略機器人：讓一份交易策略在一個交易標的上定時自己跑。"+
				"\n\n**建立不等於啟動。** 建好之後要用 trading_start_strategy_bot 才會開始跑。"+
				"\n\n**填了 positionPlan 的機器人會在訊息裡建議止損與止盈價位，"+
				"而回測只在你把那兩個距離填進去時才算它們。**"+
				"所以一套「回測賺 25%」的規則配上一組停損，"+
				"那兩件事**還沒有對過帳**——"+
				"要對帳就用 trading_backtest_trading_strategy 的 stopLossPercentage "+
				"與 takeProfitPercentage 再重演一次。"+
				"在那之前，不要拿那個 25% 替這組停損背書。",
			vo.RequestVerbSubmit, "/strategy-bots", true,
			strategyBotWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_strategy_bots",
			"列出你的每一台策略機器人，含它現在是不是在跑。",
			vo.RequestVerbRead, "/strategy-bots", true,
		),
		domains.NewApiToolDomain(
			"trading_get_strategy_bot",
			"讀一台策略機器人的設定與目前狀態。",
			vo.RequestVerbRead, "/strategy-bots/{id}", true,
			pathParameter("id", "策略機器人識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_strategy_bot",
			"改一台你自己的策略機器人。這是整份改寫：沒帶到的欄位會變成空的。",
			vo.RequestVerbReplace, "/strategy-bots/{id}", true,
			append([]vo.ToolParameterVo{pathParameter("id", "要改哪一台")},
				strategyBotWriteParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_strategy_bot",
			"刪掉一台你自己的策略機器人。",
			vo.RequestVerbRemove, "/strategy-bots/{id}", true,
			pathParameter("id", "要刪哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_start_strategy_bot",
			"啟動一台策略機器人，它開始按自己的間隔一輪一輪跑。",
			vo.RequestVerbSubmit, "/strategy-bots/{id}/power", true,
			pathParameter("id", "要啟動哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_stop_strategy_bot",
			"停掉一台策略機器人。已經跑過的那些輪次紀錄一筆都不刪。",
			vo.RequestVerbRemove, "/strategy-bots/{id}/power", true,
			pathParameter("id", "要停哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_list_strategy_bot_runs",
			"看一台策略機器人跑過哪幾輪，以及每一輪做了什麼決定。"+
				"這是它有沒有在做事的唯一證據——只看它「在跑」不代表它有在做決定。",
			vo.RequestVerbRead, "/strategy-bots/{id}/runs", true,
			pathParameter("id", "要看哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_run_strategy_bot_now",
			"叫一台策略機器人立刻跑一輪，不等它的間隔到。"+
				"\n\n用來在改完交易策略之後馬上看一眼它現在會做什麼決定，而不必等下一輪。",
			vo.RequestVerbSubmit, "/strategy-bots/{id}/runs", true,
			pathParameter("id", "要叫哪一台立刻跑"),
		),
	}
}

func telegramDeliveryApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_get_telegram_delivery",
			"讀你目前的電報通知設定。",
			vo.RequestVerbRead, "/users/me/telegram-delivery", true,
		),
		domains.NewApiToolDomain(
			"trading_save_telegram_delivery",
			"設定用哪一台電報機器人、送到哪個對話。設過了再設一次即覆蓋。"+
				"\n\n設定之後建議用 trading_send_telegram_test_message 送一則測試，"+
				"確認真的收得到——填錯的 chatId 不會有任何錯誤，只是安靜地送不到。",
			vo.RequestVerbReplace, "/users/me/telegram-delivery", true,
			bodyParameter("botToken", vo.ToolParameterKindString, "電報機器人的憑證", true),
			bodyParameter("chatId", vo.ToolParameterKindString, "要送到哪個對話", true),
		),
		domains.NewApiToolDomain(
			"trading_remove_telegram_delivery",
			"移除電報通知設定，之後不再送任何通知。",
			vo.RequestVerbRemove, "/users/me/telegram-delivery", true,
		),
		domains.NewApiToolDomain(
			"trading_send_telegram_test_message",
			"用目前的設定送一則測試訊息，確認真的收得到。",
			vo.RequestVerbSubmit, "/users/me/telegram-delivery/test-message", true,
			bodyParameter("message", vo.ToolParameterKindString, "要送什麼內容", true),
		),
	}
}
