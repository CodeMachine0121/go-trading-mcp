package main

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

func kCandleApiTools(liveUpdateWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_k_candle",
			"新增一根 K 線。一根固定涵蓋一分鐘，起始時間必須落在一分鐘刻度上、且不得指向未來。"+
				"同一個交易標的同一個起始時間再新增一次即覆蓋，不會產生第二根。"+
				"\n\n一般不需要用這一支——K 線由系統自己抓。它是給手動補資料與測試用的。",
			vo.RequestVerbSubmit, "/k-candles", true,
			append([]vo.ToolParameterVo{
				bodyParameter("symbol", vo.ToolParameterKindString, "交易標的，如 BTCUSDT 或 2330", true),
				bodyParameter("openTime", vo.ToolParameterKindString,
					"這一分鐘從何時開始（RFC3339 世界標準時間，如 2026-08-29T09:00:00Z）", true),
			}, kCandlePriceParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_list_k_candles",
			"查一段區間的原始 K 線（一根一分鐘），起訖兩端都包含，依起始時間由早到晚。"+
				"\n\n單次筆數有上限，超過即拒絕；要看長區間請改用 trading_get_k_candle_series，"+
				"它會把區間彙總成比較粗的刻度。區間內沒有資料是正常結果，回空陣列而不是錯誤。",
			vo.RequestVerbRead, "/k-candles", false,
			queryParameter("symbol", vo.ToolParameterKindString, "交易標的", true),
			queryParameter("startTime", vo.ToolParameterKindString,
				"起（RFC3339 世界標準時間，含）", true),
			queryParameter("endTime", vo.ToolParameterKindString, "訖（RFC3339 世界標準時間，含）", true),
		),
		domains.NewApiToolDomain(
			"trading_get_k_candle_series",
			"查一段區間的彙總 K 線序列：同一個刻度區間裡的 K 線合併成一根"+
				"（開盤取最早、收盤取最晚、最高取最高、最低取最低、成交數字加總）。沒有資料的刻度區間不產出那一根。"+
				"\n\n刻度邊界一律自世界標準時間當日零點起切分，查詢區間的起訖不必對齊。"+
				"\n\n**interval 與 displayableCandleCount 兩者只能給一個。**"+
				"兩個都給即整次拒絕（兩者矛盾時沒有正確的取捨）；兩個都不給時由系統挑一種刻度，"+
				"回應一律說出實際用的是哪一種，請照抄不要自行推算。",
			vo.RequestVerbRead, "/k-candles/series", false,
			queryParameter("symbol", vo.ToolParameterKindString, "交易標的", true),
			queryParameter("startTime", vo.ToolParameterKindString, "起（RFC3339 世界標準時間，含）", true),
			queryParameter("endTime", vo.ToolParameterKindString, "訖（RFC3339 世界標準時間，含）", true),
			queryParameter("interval", vo.ToolParameterKindString,
				"彙總刻度，六選一：1m／5m／15m／1h／4h／1d。與 displayableCandleCount 互斥", false),
			queryParameter("displayableCandleCount", vo.ToolParameterKindInteger,
				"你這邊一次擺得下幾根，由系統據此挑一種刻度。必須大於零。與 interval 互斥", false),
		),
		domains.NewApiToolDomain(
			"trading_get_k_candle",
			"讀一根指定的 K 線。指名的那一根不存在時回 404。",
			vo.RequestVerbRead, "/k-candles/{symbol}/{openTime}", false,
			pathParameter("symbol", "交易標的"),
			pathParameter("openTime", "起始時間（RFC3339 世界標準時間）"),
		),
		domains.NewApiToolDomain(
			"trading_update_k_candle",
			"改一根既有 K 線的價量數字。**要改哪一根由 symbol 與 openTime 決定**——"+
				"內文不要再帶一次交易標的或起始時間，帶了且與這兩個不同會被拒絕。",
			vo.RequestVerbReplace, "/k-candles/{symbol}/{openTime}", true,
			append([]vo.ToolParameterVo{
				pathParameter("symbol", "要改哪一個交易標的的 K 線"),
				pathParameter("openTime", "要改哪一根（RFC3339 世界標準時間）"),
			}, kCandlePriceParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_k_candle",
			"刪掉一根指定的 K 線。成功沒有內容可回。",
			vo.RequestVerbRemove, "/k-candles/{symbol}/{openTime}", true,
			pathParameter("symbol", "交易標的"),
			pathParameter("openTime", "起始時間（RFC3339 世界標準時間）"),
		),
		domains.NewApiToolDomain(
			"trading_backfill_k_candles",
			"手動補齊一個交易標的的歷史，從它最後一根的下一根補到現在。"+
				"\n\n**補多久由系統的設定決定，不是你說了算**——所以這一支沒有回溯天數可填。"+
				"要指定回溯多久請改用 trading_sync_k_candle_history。"+
				"\n\n加進觀察清單時會自動補一次，所以正常情況下不必按這一支。沒登錄過的代號回 404。",
			vo.RequestVerbSubmit, "/k-candles/backfill", true,
			bodyParameter("symbol", vo.ToolParameterKindString, "要補哪一個交易標的", true),
		),
		domains.NewApiToolDomain(
			"trading_sync_k_candle_history",
			"同步一段指定長度的歷史。**它不等抓完**——立刻回一筆輪次識別碼，"+
				"請拿那個識別碼用 trading_get_k_candle_history_sync 看進度。四年的一分鐘 K 線要跑十幾分鐘。"+
				"\n\n它向來源問**整段**（所以中間的破洞補得到），但**只寫進沒有的那幾根**，已經有的原封不動。"+
				"所以報告裡的「存了幾根」講的是這一次新增了幾根——整段本來就齊全時它是 0，而那是實話。"+
				"\n\n回溯天數必須在 1 到系統上限之間，超過會被擋下來並告訴你上限是多少。"+
				"沒登錄過的代號回 404。同一個標的同時只跑一趟，再按一次回 409。",
			vo.RequestVerbSubmit, "/k-candles/history", true,
			bodyParameter("symbol", vo.ToolParameterKindString, "要同步哪一個交易標的", true),
			bodyParameter("lookbackDays", vo.ToolParameterKindInteger,
				"往回抓幾天。沒有預設值，不給即拒絕", true),
		),
		domains.NewApiToolDomain(
			"trading_get_k_candle_history_sync",
			"看一趟歷史同步走到哪：status（running／succeeded／failed）、completedChunks / totalChunks、"+
				"storedCount、skippedCount。"+
				"\n\n**fetchFailureReason 與 failureReason 是兩件事**："+
				"前者是行情來源不答話（這趟仍算 succeeded，那是查到的事），後者才是交易服務自己壞掉。",
			vo.RequestVerbRead, "/k-candles/history/{id}", true,
			pathParameter("id", "trading_sync_k_candle_history 回的那個輪次識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_peek_live_k_candle",
			"看一眼某個交易標的現在的樣子。**只看一眼**：收到第一則即時更新就回，最多等幾秒。"+
				"\n\n每則更新的 status 是五者之一，而後三者要人做的事完全相反："+
				"forming（這一根還在走，數字還會變，系統不存它、指標計算也不用它）、"+
				"closed（這一根走完了，同一刻已存入）、"+
				"stalled（即時更新斷了，系統自己在重連，**等一下就好**）、"+
				"unavailable（這一檔分不到這個市場的即時名額，**不會自己好**，要改觀察清單）、"+
				"marketClosed（這個市場現在休市，**什麼都別做**）。"+
				"\n\n等滿沒有收到東西是正常結果，不是錯誤。要連續看請重複呼叫。",
			vo.RequestVerbRead, "/k-candles/live", false,
			queryParameter("symbol", vo.ToolParameterKindString, "要看哪一個交易標的", true),
		).Watching(liveUpdateWaitLimit),
	}
}

func tradingSymbolApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_list_trading_symbols",
			"列出系統認得的每一個交易標的：已登錄的，加上實際有 K 線的，去重、依名稱由小到大。"+
				"\n\n**它不等於觀察清單**——觀察清單是這張表裡標記為 isWatched 的那個子集。"+
				"每一檔帶著六件事：symbol、displayName（行情來源怎麼稱呼它；加密貨幣留空）、market、isWatched、"+
				"isWithinTradingSession（現在開著嗎）、hasTradingSession（這個市場會不會收盤）、"+
				"hasLiveUpdates（現在有沒有即時名額）。"+
				"\n\n後三件只有交易服務答得出來，請不要自行推算——休市日不是算得出來的。",
			vo.RequestVerbRead, "/trading-symbols", false,
		),
		domains.NewApiToolDomain(
			"trading_add_to_watchlist",
			"把一個交易標的加進觀察清單，之後系統每分鐘一輪自己把 K 線抓回來。"+
				"\n\n加之前會先向該市場確認代號存在並記下它給的名稱，**加完立刻補齊那一檔的歷史**"+
				"（補失敗不會讓加入失敗）。"+
				"\n\n**market 一定要給，系統不從代號長相猜**：台股的 ETF、權證有英數混合的代號，"+
				"加密貨幣的代號格式根本沒有規則。",
			vo.RequestVerbSubmit, "/watchlist", true,
			bodyParameter("symbol", vo.ToolParameterKindString, "交易標的代號", true),
			bodyParameter("market", vo.ToolParameterKindString,
				"所屬市場：taiwanStock（台股，台北時間 09:00–13:30、週一至週五）"+
					"或 crypto（加密貨幣，全天候）", true),
		),
		domains.NewApiToolDomain(
			"trading_remove_from_watchlist",
			"把一個交易標的移出觀察清單。**只停止自動抓取**——已經抓回來的 K 線一根都不刪，"+
				"這個標的仍然被系統認得、仍然選得到。",
			vo.RequestVerbRemove, "/watchlist/{symbol}", true,
			pathParameter("symbol", "要停止追蹤哪一個交易標的"),
		),
	}
}

// indicatorCalculationParameters are what an indicator calculation is asked with, on
// either market — everything but the symbol. Shared, because the trading service asks
// both calculations for the same body: a box added once is a box both get, and a
// wording improved once reads the same on both.
//
// The symbol is left to each ability: the same code names two instruments on the two
// venues, so each one says which kind of symbol it takes.
func indicatorCalculationParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("strategyScriptId", vo.ToolParameterKindInteger,
			"要跑哪一支既有的策略腳本。與 script 擇一", false),
		bodyParameter("startTime", vo.ToolParameterKindString,
			"觀察區間的起點（RFC3339 世界標準時間）。這是唯一沒有預設值的欄位", true),
		bodyParameter("endTime", vo.ToolParameterKindString,
			"觀察區間的終點，也就是計算截止時間。省略即現在；指向未來視同現在，不拒絕", false),
		bodyParameter("aggregationInterval", vo.ToolParameterKindString,
			"彙總刻度，六選一：1m／5m／15m／1h／4h／1d。省略即 1m", false),
		bodyParameter("script", vo.ToolParameterKindString,
			"一段還沒存起來的算式。與 strategyScriptId 擇一", false),
		bodyParameter("resultType", vo.ToolParameterKindString,
			"自帶算式時它產出什麼形狀：float／floatList／bool／boolList／signal。省略即 float。"+
				"指名既有策略腳本時這一格會被忽略", false),
		bodyParameter("parameters", vo.ToolParameterKindArray,
			"自帶算式時它宣告的旋鈕。指名既有策略腳本時會被忽略", false),
		bodyParameter("parameterValues", vo.ToolParameterKindArray,
			"這一次要把旋鈕調成多少，每個為 {\"name\":…, \"value\":…}。"+
				"沒給的用宣告的預設值；給了一個沒宣告過的名字則整次拒絕", false),
	}
}

func indicatorApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_calculate_indicator",
			"用一段算式算一次指標。"+
				"\n\n**要嘛指名一支既有的策略腳本（strategyScriptId），要嘛自己帶一段算式（script），"+
				"兩者只能挑一個。** 指名既有的那一支時，它已經宣告過 resultType 與 parameters，"+
				"再送一次會被忽略。"+
				"\n\n**這一支只算現貨 K 線。** 指名一支吃合約行情（marketDataKind 為 contractKCandle）的策略腳本會被拒絕"+
				"——那一支要用 trading_calculate_contract_indicator 算。"+
				"\n\n你說的是**觀察區間**（startTime 到 endTime），要拿幾根 K 線是系統算出來的"+
				"（要看幾格 ＋ 最大回看根數 − 1），沒有地方可以填、也不需要填。"+
				"\n\n幾條會讓你困惑的規則：只採用**走完**的刻度區間（還在走的那一格裝了一半，"+
				"值會自己變）；沒有資料的刻度區間不產出那一根，也不補洞；"+
				"觀察區間與該市場的交易時段完全沒有重疊時整次拒絕（台股的深夜與週末），"+
				"這與「湊不出最少可算根數」是兩回事——前者換什麼刻度都一樣。"+
				"\n\n回應多帶 interval（這次實際用的刻度）、usedCandleCount 與 openTimes"+
				"（餵給算式的每一根從哪裡開始，由早到晚），"+
				"所以要把一條線畫回圖上不必自己反推是哪幾根。"+
				"\n\n算式跑不動（讀不懂、執行失敗、越權、逾時）回 422。",
			vo.RequestVerbSubmit, "/indicator-calculations", true,
			append([]vo.ToolParameterVo{
				bodyParameter("symbol", vo.ToolParameterKindString, "要算哪一個交易標的", true),
			}, indicatorCalculationParameters()...)...,
		),
		// Sits beside the spot calculation because it runs somebody's strategy script, not market data.
		domains.NewApiToolDomain(
			"trading_calculate_contract_indicator",
			"在一個**永續合約**標的上用一段算式算一次指標。填的東西與 trading_calculate_indicator 一模一樣，"+
				"規則也一字不差（觀察區間、彙總刻度、只採用走完的刻度區間、旋鈕、指標值種類、回應的形狀）；"+
				"差別只在算式收到的是**合約行情格**而不是 K 線，而且只問合約那一條線。"+
				"合約沒有交易時段，所以不會有「觀察區間沒有交易」這種拒絕。"+
				"\n\n**要嘛指名一支吃合約行情的策略腳本（strategyScriptId），要嘛自己帶一段算式（script），兩者只能挑一個。**"+
				"\n\n**會被拒絕的情況**：指名的那一支**吃的是 K 線**（那一支要用 trading_calculate_indicator 算）；"+
				"那一支不存在、不是你的也沒上架（404，三種情形同一個答案）；"+
				"走完的格子湊不出最少可算根數時整次拒絕，並說出**可用根數**與**最少可算根數**"+
				"（從來沒存過合約 K 線的代號，可用根數就是零——先用 trading_backfill_contract_k_candles 補）；"+
				"parameterValues 給了一個沒宣告過的名字。"+
				"算式跑不動（讀不懂、入口照現貨收 K 線、執行失敗、越權、逾時）回 422。"+
				contractKCandleScriptNote,
			vo.RequestVerbSubmit, "/contract-indicator-calculations", true,
			append([]vo.ToolParameterVo{
				bodyParameter("symbol", vo.ToolParameterKindString,
					"要算哪一個合約標的，如 BTCUSDT（永續合約的代號，與現貨代號不一定對應）", true),
			}, indicatorCalculationParameters()...)...,
		),
	}
}
