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
//
// A bot is a spot bot or a contract bot, told apart by one box rather than by a second
// set of abilities: the trading service keeps both kinds behind the same door, and a
// second set of abilities would have this connector pinning a kind on the assistant's
// behalf — a rule the trading service does not have. It is the same box a strategy
// script and a trading strategy carry, and it is forwarded only when filled in: a
// blank means a spot bot on a create and "keep what it is" on a rewrite, and filling
// in kCandle here would get a contract bot's rename refused.
func strategyBotWriteParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("name", vo.ToolParameterKindString, "這台機器人叫什麼", true),
		bodyParameter("marketDataKind", vo.ToolParameterKindString,
			"這台機器人吃哪一種行情，二選一：kCandle（**現貨機器人**，盯現貨 K 線）或 "+
				"contractKCandle（**合約機器人**，盯永續合約的合約行情格）。"+
				"**建立時不給就是 kCandle**；**修改時不給就是保留原本的**，照抄原本的也可以。"+
				"**建立後不得更換**——換成另一種會被拒絕，要另一種請另建一台", false),
		bodyParameter("symbol", vo.ToolParameterKindString,
			"它盯著哪一個交易標的。**合約機器人盯的合約標的必須已經在合約追蹤名單上**"+
				"（不在會被拒絕，先用 trading_add_to_contract_watchlist 加進去）；現貨機器人照舊", true),
		bodyParameter("tradingStrategyId", vo.ToolParameterKindInteger,
			"它照哪一份交易策略做決定。**交易策略吃的行情必須與這台機器人相同**："+
				"現貨機器人只能引用吃 K 線的交易策略，合約機器人只能引用吃合約行情的交易策略，"+
				"對不上會整台被拒絕", true),
		bodyParameter("triggerIntervalMinutes", vo.ToolParameterKindInteger,
			"每幾分鐘跑一輪", true),
		bodyParameter("positionPlan", vo.ToolParameterKindObject,
			"這台機器人每一輪要建議押多少、停在哪裡。**整組可以不給**——"+
				"不給的機器人只報方向，不報數字。形狀："+
				"{\"capital\":\"50000\", \"sizingMode\":\"percentage\", \"sizingValue\":\"10\", "+
				"\"stopLossPercentage\":\"3\", \"takeProfitPercentage\":\"5\"}。"+
				"capital 是這一組的開關：不給它，其餘幾樣填了也不算。"+
				"金額一律以字串給精確小數。"+
				"sizingMode 三選一：allIn 全押（不必給 sizingValue）、percentage 押資金的百分之幾、"+
				"fixedAmount 每次押固定金額；不給即 allIn。"+
				"stopLossPercentage 與 takeProfitPercentage 是百分點（3 就是 3%），各自可以單獨不給。"+
				"\n\n**槓桿倍數只有合約機器人收**，合約機器人才在這一組裡另加一項，例如 \"leverage\":\"5\"（字串形式的精確小數）：不給或零即一倍，小於一會被拒絕，"+
				"超過那個合約標的最高那一級允許的槓桿會被拒絕並說出上限。"+
				"合約機器人建議的是**保證金**（照 sizingMode 從 capital 算出來的那一筆），名目＝保證金 × 槓桿，"+
				"止損止盈的虧賺照名目算；做空時止損在上、止盈在下。"+
				"止損距離乘上槓桿達到 100% 時，訊息會警告還沒到止損就會先被強制平倉——那組停損等於沒設。"+
				"**現貨機器人給大於一倍的槓桿會被拒絕**：現貨是拿現金換東西，沒有人借錢給你", false),
	}
}

// contractStrategyBotSkippedRoundNote is what the two abilities that read a bot's rounds
// say about a contract bot's quiet ones. The trading service skips a contract round whose
// contract's newest one-minute candle is more than five minutes old, and books it as hold
// — the same word a round that ran and concluded nothing leaves — so without this an
// assistant reads a bot that has gone blind as a bot with no signal yet.
const contractStrategyBotSkippedRoundNote = "\n\n**合約機器人一連串的 hold 不一定是沒有信號**：它的合約標的最新一根一分鐘合約 K 線" +
	"比現在早超過 5 分鐘（或一根都沒有）時，那一輪會直接跳過、不送訊息，紀錄上一樣是 hold。" +
	"把它讀成「沒有信號」之前，先用 trading_list_contract_k_candles 看那個合約標的最新的 K 線是不是還在進來；" +
	"停了的話多半是它已不在合約追蹤名單上（trading_add_to_contract_watchlist 加回去）"

func strategyBotApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_strategy_bot",
			"建立一台策略機器人：讓一份交易策略在一個交易標的上定時自己跑。"+
				"\n\n**機器人分兩種，建立當下就定了**：現貨機器人（marketDataKind 不給或 kCandle）引用吃 K 線的交易策略、盯現貨標的；"+
				"合約機器人（marketDataKind 為 contractKCandle）引用吃合約行情的交易策略、盯合約追蹤名單上的合約標的，"+
				"每一輪照那份交易策略的交易模式說出做多、做空、平多或平空，並可以在 positionPlan 設槓桿。"+
				"合約機器人的止損止盈要對帳就用 trading_backtest_contract_trading_strategy 重演。"+
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
			"列出你的每一台策略機器人，含它現在是不是在跑。"+
				"每一台都帶著 marketDataKind，看得出它是現貨機器人（kCandle）還是合約機器人（contractKCandle）。",
			vo.RequestVerbRead, "/strategy-bots", true,
			queryParameter("marketDataKind", vo.ToolParameterKindString,
				"只列其中一種：kCandle（現貨機器人）或 contractKCandle（合約機器人）。不給就全部列出；其他值會被拒絕", false),
		),
		domains.NewApiToolDomain(
			"trading_get_strategy_bot",
			"讀一台策略機器人的設定與目前狀態，含它吃哪一種行情（marketDataKind）；"+
				"合約機器人的 positionPlan 另帶著它的槓桿倍數（leverage）。",
			vo.RequestVerbRead, "/strategy-bots/{id}", true,
			pathParameter("id", "策略機器人識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_strategy_bot",
			"改一台你自己的策略機器人。這是整份改寫：沒帶到的欄位會變成空的。"+
				"\n\n**例外是 marketDataKind**：不給就是保留原本的種類，只改名字或間隔時不必重帶它；"+
				"換成另一種會被拒絕。合約機器人的 positionPlan 是整組改寫，沒帶 leverage 就回到一倍。",
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
				"這是它有沒有在做事的唯一證據——只看它「在跑」不代表它有在做決定。"+contractStrategyBotSkippedRoundNote,
			vo.RequestVerbRead, "/strategy-bots/{id}/runs", true,
			pathParameter("id", "要看哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_run_strategy_bot_now",
			"叫一台策略機器人立刻跑一輪，不等它的間隔到。"+
				"\n\n用來在改完交易策略之後馬上看一眼它現在會做什麼決定，而不必等下一輪。"+contractStrategyBotSkippedRoundNote,
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
