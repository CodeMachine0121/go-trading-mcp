package main

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// strategyBotWriteParameters are what a strategy bot is made of. Shared by creating
// one and rewriting one.
func strategyBotWriteParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("name", vo.ToolParameterKindString, "這台機器人叫什麼", true),
		bodyParameter("symbol", vo.ToolParameterKindString, "它盯著哪一個交易標的", true),
		bodyParameter("tradingStrategyId", vo.ToolParameterKindInteger,
			"它照哪一份交易策略做決定", true),
		bodyParameter("triggerIntervalMinutes", vo.ToolParameterKindInteger,
			"每幾分鐘跑一輪", true),
	}
}

func strategyBotApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_strategy_bot",
			"建立一台策略機器人：讓一份交易策略在一個交易標的上定時自己跑。"+
				"\n\n**建立不等於啟動。** 建好之後要用 trading_start_strategy_bot 才會開始跑。",
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

func assistantConversationApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_ask_assistant",
			"問交易服務內建的行情助手一句話。"+
				"\n\n**它回 202，不回答案。** 它收下問題、回覆對話識別碼與這次回答的識別碼，"+
				"答案在連線之外寫完——一次回答可能來回四十趟，那可能是好幾分鐘。"+
				"請隔一會兒用 trading_get_assistant_conversation 讀那段對話，看最後一則的 status："+
				"running（還在寫）、answered（寫完了）、failed（壞了，並留下一句原因）。"+
				"\n\n**同一段對話一次只跑一則**，前一則還在寫時再送會回 409。"+
				"不指名對話即開一段新的。今日額度用盡回 429 並說明何時重置。"+
				"\n\n注意：這個助手與你（正在讀這段文字的助理）是兩回事。"+
				"它跑在交易服務裡、看得到交易服務的資料，但不知道這段對話。",
			vo.RequestVerbSubmit, "/chat", true,
			bodyParameter("question", vo.ToolParameterKindString, "要問什麼。空白即拒絕", true),
			bodyParameter("conversationId", vo.ToolParameterKindInteger,
				"要接著問哪一段對話。省略即開一段新的", false),
		),
		domains.NewApiToolDomain(
			"trading_list_assistant_conversations",
			"列出與行情助手的每一段對話，最近有動靜的排前面。",
			vo.RequestVerbRead, "/chat/conversations", true,
		),
		domains.NewApiToolDomain(
			"trading_get_assistant_conversation",
			"讀一段對話的每一則訊息，依時間由早到晚。"+
				"\n\n這是拿 trading_ask_assistant 那個問題的答案的方式。"+
				"看最後一則的 status：running 就等一下再讀一次。",
			vo.RequestVerbRead, "/chat/conversations/{id}", true,
			pathParameter("id", "對話識別碼"),
		),
	}
}
