package main

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// contractKCandleFigureParameters are the figures of one perpetual contract K candle.
//
// They are their own list rather than the spot ones with a few added, because every
// one of them is required here — a contract candle has a single source that reports
// all of them, and one without its mark price, index price or premium index is not a
// contract candle. Shared by creating one and rewriting one, which the trading service
// holds to the same rules.
func contractKCandleFigureParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("open", vo.ToolParameterKindString, "開盤價（字串形式的精確小數）", true),
		bodyParameter("high", vo.ToolParameterKindString, "最高價。不得低於最低價", true),
		bodyParameter("low", vo.ToolParameterKindString, "最低價", true),
		bodyParameter("close", vo.ToolParameterKindString, "收盤價", true),
		bodyParameter("volume", vo.ToolParameterKindString, "成交量", true),
		bodyParameter("quoteVolume", vo.ToolParameterKindString, "成交額。合約每一項都必填", true),
		bodyParameter("takerBuyBaseVolume", vo.ToolParameterKindString, "主動買入量", true),
		bodyParameter("takerBuyQuoteVolume", vo.ToolParameterKindString, "主動買入額", true),
		bodyParameter("tradeCount", vo.ToolParameterKindInteger,
			"成交筆數。零是合法的（那一分鐘真的可能沒有成交），所以不能省略當成零", true),
		bodyParameter("markOpen", vo.ToolParameterKindString,
			"標記價格開盤。標記價格是市場判定槓桿部位撐不撐得住的參考價，四個都必填", true),
		bodyParameter("markHigh", vo.ToolParameterKindString, "標記價格最高。不得低於標記最低", true),
		bodyParameter("markLow", vo.ToolParameterKindString, "標記價格最低", true),
		bodyParameter("markClose", vo.ToolParameterKindString, "標記價格收盤", true),
		bodyParameter("indexOpen", vo.ToolParameterKindString,
			"指數價格開盤（幾家主要交易所現貨價的加權平均）。四個都必填、不得為負", true),
		bodyParameter("indexHigh", vo.ToolParameterKindString, "指數價格最高。不得低於指數最低", true),
		bodyParameter("indexLow", vo.ToolParameterKindString, "指數價格最低", true),
		bodyParameter("indexClose", vo.ToolParameterKindString, "指數價格收盤", true),
		bodyParameter("premiumIndexOpen", vo.ToolParameterKindString,
			"溢價指數開盤：合約比指數價格高或低多少的**比例**。**可以是負的**（合約比現貨便宜），四個都必填", true),
		bodyParameter("premiumIndexHigh", vo.ToolParameterKindString, "溢價指數最高。不得低於溢價指數最低", true),
		bodyParameter("premiumIndexLow", vo.ToolParameterKindString, "溢價指數最低", true),
		bodyParameter("premiumIndexClose", vo.ToolParameterKindString, "溢價指數收盤", true),
	}
}

// contractRangeParameters are how every contract series is read back: one contract,
// one stretch, both ends included.
func contractRangeParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		queryParameter("symbol", vo.ToolParameterKindString, "合約標的，如 BTCUSDT（永續合約的代號）", true),
		queryParameter("startTime", vo.ToolParameterKindString, "起（RFC3339 世界標準時間，含）", true),
		queryParameter("endTime", vo.ToolParameterKindString, "訖（RFC3339 世界標準時間，含）", true),
	}
}

// contractApiTools are everything about **perpetual contracts**, which the trading
// service keeps as a line of their own beside spot: their own candles, their own
// watchlist, their own history syncs — the same symbol names a different instrument
// on each venue, and neither side ever reads or overwrites the other.
//
// Every name carries "contract" so that an assistant choosing between a spot ability
// and its contract twin cannot mistake one for the other by name alone.
func contractApiTools(liveUpdateWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_contract_k_candle",
			"新增一根**永續合約** K 線。與現貨 K 線是兩種東西：同一個代號同一個起始時間兩邊各存一根、互不覆蓋。"+
				"\n\n一根合約 K 線**每一項都必填**——價量、成交筆數、標記價格、指數價格、溢價指數，缺任何一項即拒絕。"+
				"同代號同起始時間再新增一次即覆蓋。一般不需要用這一支，合約 K 線由系統自己抓。",
			vo.RequestVerbSubmit, "/contract-k-candles", false,
			append([]vo.ToolParameterVo{
				bodyParameter("symbol", vo.ToolParameterKindString, "合約標的，如 BTCUSDT", true),
				bodyParameter("openTime", vo.ToolParameterKindString,
					"這一分鐘從何時開始（RFC3339 世界標準時間，落在一分鐘刻度上、不得指向未來）", true),
			}, contractKCandleFigureParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_list_contract_k_candles",
			"查一段區間的**永續合約** K 線（一根一分鐘），起訖兩端都包含，依起始時間由早到晚。"+
				"只回合約的，不會回到現貨那根。"+
				"\n\n每一根除了價量，還帶標記價格、指數價格、溢價指數各自的開高低收與成交筆數。"+
				"**指數價格與溢價指數是 null 的那幾根，是這兩項出現之前存下的舊資料**，不是零；"+
				"用 trading_sync_contract_k_candle_history 同步那一段就會補上。"+
				"單次筆數有上限，超過即拒絕；要看長區間請改用 trading_get_contract_k_candle_series。"+
				"結束早於開始也會被拒絕。",
			vo.RequestVerbRead, "/contract-k-candles", false,
			contractRangeParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_get_contract_k_candle_series",
			"查一段區間的**永續合約**彙總 K 線序列：同一個刻度區間裡的合約 K 線合併成一根，"+
				"說法與規則與 trading_get_k_candle_series（現貨）一字不差。沒有資料的刻度區間不產出那一根。"+
				"\n\n合併方式：價量照現貨（開取最早、收取最晚、高取最高、低取最低、成交數字加總），成交筆數加總，"+
				"標記價格、指數價格、溢價指數三條線各自一樣合併。"+
				"**一格裡只要有一根是舊資料、缺指數價格或溢價指數，那一格的那條線就是 null**——不是零，也不是用剩下幾根湊的。"+
				"\n\n**interval 與 displayableCandleCount 兩者只能給一個**，兩個都給即整次拒絕；兩個都不給時由系統挑一種刻度，"+
				"回應一律說出實際用的是哪一種，請照抄不要自行推算。區間依刻度切出的格數超過單次上限會被拒絕，"+
				"可縮小區間或改用更長的刻度。",
			vo.RequestVerbRead, "/contract-k-candles/series", false,
			append(contractRangeParameters(),
				queryParameter("interval", vo.ToolParameterKindString,
					"彙總刻度，六選一：1m／5m／15m／1h／4h／1d。與 displayableCandleCount 互斥", false),
				queryParameter("displayableCandleCount", vo.ToolParameterKindInteger,
					"你這邊一次擺得下幾根，由系統據此挑一種刻度。必須大於零。與 interval 互斥", false),
			)...,
		),
		domains.NewApiToolDomain(
			"trading_get_contract_k_candle",
			"讀一根指定的永續合約 K 線。不存在時回 404。",
			vo.RequestVerbRead, "/contract-k-candles/{symbol}/{openTime}", false,
			pathParameter("symbol", "合約標的"),
			pathParameter("openTime", "起始時間（RFC3339 世界標準時間）"),
		),
		domains.NewApiToolDomain(
			"trading_update_contract_k_candle",
			"改一根既有永續合約 K 線的數字。**要改哪一根由 symbol 與 openTime 決定**——"+
				"內文不要再帶交易標的或起始時間。所有數字一樣必填。",
			vo.RequestVerbReplace, "/contract-k-candles/{symbol}/{openTime}", false,
			append([]vo.ToolParameterVo{
				pathParameter("symbol", "要改哪一個合約標的的 K 線"),
				pathParameter("openTime", "要改哪一根（RFC3339 世界標準時間）"),
			}, contractKCandleFigureParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_contract_k_candle",
			"刪掉一根指定的永續合約 K 線。現貨同代號同時間那根完全不受影響。成功沒有內容可回；"+
				"指名的那一根不存在時回 404。",
			vo.RequestVerbRemove, "/contract-k-candles/{symbol}/{openTime}", false,
			pathParameter("symbol", "合約標的"),
			pathParameter("openTime", "起始時間（RFC3339 世界標準時間）"),
		),
		domains.NewApiToolDomain(
			"trading_backfill_contract_k_candles",
			"手動補齊一個合約標的的 K 線，從它最後一根補到現在。**補多久由系統的設定決定**，沒有回溯天數可填。"+
				"\n\n它只補 K 線——資金費率與持倉統計不需要手動補，系統每一輪都從上一筆接著問。"+
				"沒登錄過的代號回 404。",
			vo.RequestVerbSubmit, "/contract-k-candles/backfill", false,
			bodyParameter("symbol", vo.ToolParameterKindString, "要補哪一個合約標的", true),
		),
		domains.NewApiToolDomain(
			"trading_sync_contract_k_candle_history",
			"同步一段指定長度的永續合約歷史。**它不等抓完**——立刻回一筆輪次（識別碼與現貨那串互不相干），"+
				"請用 trading_get_contract_k_candle_history_sync 看進度。"+
				"\n\n它問整段、只寫沒有的；**已經齊全的那幾天連問都不問**。"+
				"指數價格與溢價指數出現之前存下的舊 K 線不算齊全，同步會**只把那兩組補上**，其他數字不動。"+
				"\n\n永續合約的歷史比現貨短：起點比合約上市還早時，來源有幾根就存幾根，不算失敗。"+
				"沒登錄過的代號回 404；同一個標的同時只跑一趟，再按一次回 409。",
			vo.RequestVerbSubmit, "/contract-k-candles/history", false,
			bodyParameter("symbol", vo.ToolParameterKindString, "要同步哪一個合約標的", true),
			bodyParameter("lookbackDays", vo.ToolParameterKindInteger,
				"往回抓幾天。沒有預設值；超過上限會被拒絕並說出上限", true),
		),
		domains.NewApiToolDomain(
			"trading_get_contract_k_candle_history_sync",
			"看一趟永續合約歷史同步走到哪。**識別碼是合約自己那一串**，拿現貨的識別碼來問會問到別的東西或查無此趟。"+
				"\n\nfetchFailureReason（來源不答話，這趟仍算 succeeded）與 failureReason（交易服務自己壞掉）是兩件事。",
			vo.RequestVerbRead, "/contract-k-candles/history/{id}", false,
			pathParameter("id", "trading_sync_contract_k_candle_history 回的那個輪次識別碼"),
		),
		// The contract twin of trading_peek_live_k_candle, and a separate ability rather
		// than a box on that one: the trading service follows contracts live on a route
		// and a stream of their own, and one ability answering for both would let an
		// assistant quote a spot price to a question about the perpetual contract.
		domains.NewApiToolDomain(
			"trading_peek_live_contract_k_candle",
			"看一眼某個**永續合約**標的現在的樣子。**只看一眼**：收到第一則即時更新就回，最多等幾秒。"+
				"與 trading_peek_live_k_candle 是兩件事——同一個代號的現貨與合約是兩個市場，這一件只看合約那一邊。"+
				"\n\n內容是**最新價**那一根一分鐘 K 線的開高低收與成交量，**不含標記價格**"+
				"（標記價格、指數價格隨每分鐘那一輪存下的合約 K 線出現，用 trading_list_contract_k_candles 讀）。"+
				"\n\n每則更新的 status 是三者之一："+
				"forming（這一根還在走，數字還會變，系統不存它、指標計算也不用它）、"+
				"closed（這一根走完了，但**不是由即時更新存下**——每分鐘那一輪會在一分鐘內把完整的那一根存進來）、"+
				"stalled（即時更新斷了，系統自己在重連，**等一下就好**）。"+
				"合約永不休市、也沒有即時名額上限，所以**不會**出現 unavailable 或 marketClosed。"+
				"\n\n**只看得到合約追蹤名單上的合約標的**：不在名單上會被拒絕，先用 trading_add_to_contract_watchlist 加進去；"+
				"系統不認得的代號回找不到。"+
				"\n\n等滿沒有收到東西是正常結果，不是錯誤。要連續看請重複呼叫。",
			vo.RequestVerbRead, "/contract-k-candles/live", false,
			queryParameter("symbol", vo.ToolParameterKindString, "要看哪一個合約標的", true),
		).Watching(liveUpdateWaitLimit),
		domains.NewApiToolDomain(
			"trading_list_contract_trading_symbols",
			"列出系統認得的每一個**永續合約**標的：已登錄的，加上實際有合約 K 線的，去重、依名稱排序。"+
				"它與 trading_list_trading_symbols（現貨）是兩份清單，同一個代號可以兩邊都在。"+
				"\n\n每一檔帶 isWatched 與 **tradingSpecification**（交易規格：價格跳動單位、數量步進、最小下單量、"+
				"最小名目、維持保證金率、強平手續費率、資金費率結算間隔、確認時間）。"+
				"tradingSpecification 為 null 代表還沒記下，不是規格為零。"+
				"**這裡的維持保證金率只是最小那一級**；部位越大比例越高，完整的每一級請用 "+
				"trading_get_contract_maintenance_margin_tiers。",
			vo.RequestVerbRead, "/contract-trading-symbols", false,
		),
		domains.NewApiToolDomain(
			"trading_add_to_contract_watchlist",
			"把一個**永續合約**標的加進合約追蹤名單（與現貨觀察清單各自獨立）。"+
				"\n\n加之前會向合約來源確認它存在、還在交易、而且是永續的——已停止交易的、有交割日的、"+
				"來源認不得的都用同一句「找不到這個代號」拒絕。來源當下不可用時回 503，稍後再試。"+
				"\n\n**加完當場補齊四樣東西**：合約 K 線、從上市第一天起的完整資金費率、最近三十天的持倉統計、"+
				"交易規格（有設定帳戶金鑰時還有完整的維持保證金分級）。**所以這一支要等二十秒左右**，那是正常的。"+
				"任何一樣補失敗都不會讓加入失敗。",
			vo.RequestVerbSubmit, "/contract-watchlist", false,
			bodyParameter("symbol", vo.ToolParameterKindString,
				"合約代號。注意合約與現貨的代號不一定對應：現貨 SHIBUSDT 在合約叫 1000SHIBUSDT，價格差一千倍", true),
		),
		domains.NewApiToolDomain(
			"trading_remove_from_contract_watchlist",
			"把一個合約標的移出合約追蹤名單。**只停止追蹤**——已存下的合約 K 線、資金費率結算、持倉統計一筆都不刪，"+
				"交易規格照舊每天刷新；現貨那邊完全不受影響。"+
				"\n\n**但盯這個合約標的的合約機器人會就此失明**：新的合約 K 線不再進來，它每一輪都會跳過——"+
				"不送訊息、紀錄上是 hold——直到把它加回來為止。移除之前先用 trading_list_strategy_bots "+
				"（marketDataKind 給 contractKCandle）看有沒有合約機器人在盯它，有的話先跟使用者確認。",
			vo.RequestVerbRemove, "/contract-watchlist/{symbol}", false,
			pathParameter("symbol", "要停止追蹤哪一個合約標的"),
		),
		domains.NewApiToolDomain(
			"trading_list_contract_funding_rate_settlements",
			"查一個合約標的在一段時間內的**資金費率結算**，依結算時間由早到晚。"+
				"\n\n每一筆是一次結算：settlementTime、fundingRate、markPrice。"+
				"**費率為正時做多的人付給做空的人，為負時反過來**，金額按部位名目計——這是合約持倉最重要的一項成本，"+
				"持倉跨過結算時間點才收付。多數合約每八小時結算一次（有的四小時、一小時，看交易規格的 fundingIntervalHours）。"+
				"\n\n結算時間照來源原樣記下，可能帶一毫秒的尾數，**不要自行取整**。"+
				"markPrice 為 null 的是來源早年沒記下的結算，不是零。區間內沒有結算回空陣列。"+
				"\n\n**會被拒絕的情況**：結束早於開始；區間裡的結算超過單次筆數上限（請縮小區間，分段查）。",
			vo.RequestVerbRead, "/contract-funding-rate-settlements", false,
			contractRangeParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_contract_position_statistics",
			"查一個合約標的在一段時間內的**持倉統計**（交易所每五分鐘一筆），依統計時間由早到晚。"+
				"\n\n每一筆有：openInterest／openInterestValue（還沒平倉的合約總量與它值多少錢）、"+
				"account 開頭的三個（全體帳戶裡做多、做空各佔多少與比值）、"+
				"topTraderPosition 開頭的三個（持倉最大那批帳戶的**部位**裡多空各佔多少與比值）。"+
				"價漲而持倉量漲是新資金進場，價漲而持倉量跌是空方平倉。"+
				"\n\n**來源只留最近三十天**，所以系統手上的歷史是從加入追蹤名單前三十天開始錄的；"+
				"更早的查不到不是壞掉，是從來沒有人錄。"+
				"\n\n**會被拒絕的情況**：結束早於開始；區間裡的持倉統計超過單次筆數上限"+
				"（五分鐘一筆，一千筆約三天半，請縮小區間分段查）。",
			vo.RequestVerbRead, "/contract-position-statistics", false,
			contractRangeParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_get_contract_maintenance_margin_tiers",
			"查一個合約標的**完整的維持保證金分級**，由第一級到最後一級。"+
				"\n\n每一級：notionalFloor～notionalCap（這一級涵蓋的部位名目）、maintenanceMarginRate、"+
				"maintenanceAmount（速算額）、maximumLeverage、confirmedAt。"+
				"一筆部位的維持保證金＝名目 × 它所在那一級的維持保證金率 − 那一級的速算額。"+
				"\n\n**這份資料要交易服務設定了幣安帳戶金鑰才會有**；沒設定時回空陣列，那不是錯誤，"+
				"而是只能用交易規格裡最小那一級的維持保證金率。"+
				"還沒抓過分級的合約標的一樣回空陣列；代號留白會被拒絕。",
			vo.RequestVerbRead, "/contract-maintenance-margin-tiers", false,
			queryParameter("symbol", vo.ToolParameterKindString, "合約標的", true),
		),
	}
}
