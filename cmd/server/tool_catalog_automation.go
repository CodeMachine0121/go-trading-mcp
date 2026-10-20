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
		bodyParameter("symbols", vo.ToolParameterKindArray,
			"它盯著哪幾個交易標的（**監聽標的**），字串陣列，例如 [\"BTCUSDT\", \"ETHUSDT\"]：**1 到 10 個、不得重複**"+
				"（大小寫不同視為同一個，例如 BTCUSDT 與 btcusdt 會被拒絕說重複了）。不給、給空的、超過 10 個都會被拒絕。"+
				"同一份交易策略每一輪分別套在每一個監聽標的上，各自判斷、各自送訊息、各自記執行紀錄。"+
				"**合約機器人的每一個標的都必須已經在合約追蹤名單上**（不在的會被點名拒絕，先用 trading_add_to_contract_watchlist 加進去），"+
				"positionPlan 的槓桿也不得超過**任何一個**標的的上限（會點名哪一個最高幾倍）。"+
				"\n\n**修改時給的是修改後的完整清單**：留下來的標的保留它的狀態（上一次送出的信號、機器人持倉…），"+
				"新加的從空白開始，**沒列進去的就是要移除**（它過去的執行紀錄仍保留）。"+
				"**只想改別的設定時也要把原本的標的照抄回去**，先用 trading_get_strategy_bot 讀 watchedSymbols——漏列一個就是把它移除", true),
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
				"\n\n**合約機器人的建議照交易所規則算**：數量＝名目 ÷ 參考價，照數量步進往下取整，"+
				"保證金與名目照取整後的數量重算；止損止盈價照價格跳動單位取最接近的一檔。"+
				"另附**預估強平價**（逐倉、以參考價當進場價、用名目所在那一級的維持保證金；那個標的沒有分級時用最小那一級估算，訊息會說），"+
				"顯示的強平價照價格跳動單位取整；**做多的預估強平價在零以下時**（例如一倍做多）不給價格，"+
				"改寫「這個槓桿下不會被強制平倉」，也不會有強平警告。"+
				"還附**資金費率成本估算**（最近一次結算的費率 × 名目，標明是估算；還沒有結算紀錄就不估）。"+
				"**止損比預估強平價還遠時**，訊息會警告還沒到止損就會先被強制平倉——那組停損等於沒設；"+
				"比的是**未取整**的預估強平價，止損正好落在強平價上算止損先到、不警告，所以顯示的兩個價看起來相同時仍可能警告。"+
				"\n\n**交易所不收的那一筆**：保證金那一行（保證金、槓桿、名目，未取整）照樣寫，"+
				"接著只有一行不收的原因與它的數字——取整後的數量低於最小下單量、名目低於最小名目、"+
				"或名目落在的那一級最高槓桿低於這台機器人的槓桿（說出那一級最高幾倍）；"+
				"不給數量、止損止盈、強平價與資金費率。"+
				"遇到時調整 capital、sizingMode／sizingValue 或 leverage。"+
				"\n\n**合約標的還沒有交易規格時**照舊建議，但不取整、不估強平價，訊息會說明；"+
				"這時強平警告退回粗略判斷（止損距離 × 槓桿達到 100%）。"+
				"**現貨機器人給大於一倍的槓桿會被拒絕**：現貨是拿現金換東西，沒有人借錢給你", false),
		bodyParameter("autoOrderEntrySlippageTolerancePercentage", vo.ToolParameterKindString,
			"**進場滑價容忍度**（字串形式的精確小數，百分點：\"0.1\" 就是 0.1%），**介於 0 到 5（含）**，超出會被拒絕。"+
				"**建立時不給就是 0.1**；**修改時不給就是保留原本的值**，只改別的設定時不要帶它。"+
				"\n\n它決定**合約機器人自動下單開倉**的進場單價格：做多掛在參考價 ×（1 ＋ 容忍度 ÷ 100）、做空掛在參考價 ×（1 − 容忍度 ÷ 100），"+
				"照價格跳動單位往不超過容忍度的方向取整——**成交價絕不會比參考價差超過這麼多**。"+
				"進場單**從那一輪說出結論起最多等一個觸發間隔**：全部成交就掛止損止盈；部分成交就撤掉剩下的、保留並保護已成交的那一份；"+
				"一點都沒成交就撤單、維持空手。"+
				"等待中只要下一輪說出新結論、使用者關掉自動下單或停止機器人，進場單就立刻撤掉，已成交的那一份照樣保留；"+
				"**部分成交的那一份在進場單收尾前還沒有止損**，最長一個觸發間隔。"+
				"**0 代表只肯在參考價成交，快速行情裡可能常常不成交**；設得越寬越容易成交、但可能付出越多滑價。"+
				"**平倉一律立刻以市場價成交，不受它影響**。"+
				"現貨機器人也帶著這一格、也改得動，但現貨機器人目前不下單，所以沒有作用。"+
				"跟其他設定一樣，**只有已停止的機器人改得動**", false),
	}
}

// contractStrategyBotRunDetailsNote is what the two abilities that read a bot's rounds
// say about the three figures a contract round's record carries beyond the spot ones.
// Without it an assistant reading back a contract bot's history cannot tell a short
// from a long, nor how many times its margin a suggestion carried.
const contractStrategyBotRunDetailsNote = "\n\n**合約機器人的執行紀錄可能多帶三樣**：這個功能上線之後、那一輪有建議部位時，" +
	"suggestedDirection（long 做多／short 做空）、suggestedLeverage（槓桿倍數）與 suggestedNotional（名目）。" +
	"現貨機器人的紀錄沒有這三樣，交易所不收那一筆的那一輪也沒有，**更早以前的紀錄也沒有**——沒有不代表那一輪沒建議。" +
	"合約機器人紀錄上的 suggestedStake 是**保證金**；它與止損價、止盈價在那個合約標的有交易規格時是照交易所規則**取整後**的數字，" +
	"與當時送出的訊息一致；那個合約標的還沒有交易規格時、以及更早以前的紀錄，都是**未取整**的數字"

// contractStrategyBotSkippedRoundNote is what the two abilities that read a bot's rounds
// say about a contract bot's quiet ones. The trading service skips a contract round whose
// contract's newest one-minute candle is more than five minutes old, and books it as hold
// — the same word a round that ran and concluded nothing leaves — so without this an
// assistant reads a bot that has gone blind as a bot with no signal yet.
const contractStrategyBotSkippedRoundNote = "\n\n**合約機器人一連串的 hold 不一定是沒有信號**：某個監聽標的最新一根一分鐘合約 K 線" +
	"比現在早超過 5 分鐘（或一根都沒有）時，**只有那個標的**那一輪會直接跳過、不送訊息，紀錄上一樣是 hold（看那一筆的 symbol），其他標的照常。" +
	"把它讀成「沒有信號」之前，先用 trading_list_contract_k_candles 看那個合約標的最新的 K 線是不是還在進來；" +
	"停了的話多半是它已不在合約追蹤名單上（trading_add_to_contract_watchlist 加回去）"

const strategyBotWatchedSymbolsNote = "\n\n**每個監聽標的的狀態在 watchedSymbols 裡**（依加入的先後順序），每一個帶著：" +
	"symbol 標的、lastSentSignal 上一次送出的信號（沒送過就沒有）、conflicting 衝突記號、haltReason 標的停擺原因（沒停擺就沒有），" +
	"合約機器人另帶 autoOrderPosition（那個標的的機器人持倉：direction 為 long／short，空字串即空手；quantity 數量）。" +
	"機器人本身**不再**有單一的 symbol、lastSentSignal、conflicting、autoOrderPosition——都在各監聽標的上。" +
	"\n\n**兩種停擺分開讀**：機器人上的 haltReason 是**整台停擺**（交易策略或策略腳本找不到、Telegram 出問題，或全部標的都停擺），整台不再跑；" +
	"某個監聽標的上的 haltReason（scriptFailed 策略腳本在它身上算不出來、orderFlowRefused 訂單流超過範圍）是**標的停擺**——" +
	"只有那個標的不跑，**其他標的照常、機器人仍在跑**，不要說成整台停了。只盯一個標的時，它停擺就是整台停擺。" +
	"\n\n**停止再啟動**會清掉每個標的的停擺、上一次送出的信號與衝突記號，機器人持倉不動；這是讓停擺的標的恢復的辦法。" +
	"對已停止的機器人用 trading_run_strategy_bot_now 立刻跑一輪時，每一個標的都跑（含停擺中的），跑出結論的標的就不再停擺"

const strategyBotPerSymbolExposureNote = "\n\n**部位規劃是整台機器人的設定，每個監聽標的各自照整份算**：" +
	"盯 N 個標的時，同時最多可能有 N 份部位——**總曝險最多是部位規劃的 N 倍**，沒有同時持倉上限。建立或加標的時要讓使用者知道。" +
	"**合約機器人開著自動下單時，每一個監聽標的都會真的在使用者的幣安帳戶上各自開倉平倉**——加一個標的就是多一個會真的下單的標的"

const strategyBotAutoOrderNote = "\n\n**每一台機器人都帶著 autoOrderEnabled（自動下單：true 開、false 關）**，現貨與合約機器人都有。" +
	autoOrderReallyTradesNote + "關著的機器人每一輪只判信號、只送 Telegram 通知。" +
	"回覆裡沒有這一格時，不要猜它是開還是關。" +
	"\n\n**每一台機器人也帶著 autoOrderEntrySlippageTolerancePercentage（進場滑價容忍度，百分點）**：" +
	"自動下單開倉時，進場單的價格最多比參考價差這麼多（做多最多貴這麼多、做空最多便宜這麼多），一個觸發間隔內沒成交就撤單"

const strategyBotReplyNote = strategyBotWatchedSymbolsNote + strategyBotAutoOrderNote

// contractAutoOrderResultNote is what the two abilities that lead to a bot's rounds say about
// a round's auto order, so that an assistant reads an entry order that filled nothing as
// cancelled rather than as a failure, and reads a positive slippage as the owner's cost.
const contractAutoOrderResultNote = "\n\n**合約機器人那一輪排了自動下單時，紀錄多帶一份 autoOrder（下單結果）**；沒排的那一輪沒有。" +
	"status 是整體結果：pending 等待下單（還在送或進場單還在等成交）、filled 全部完成、" +
	"partiallyDone 部分完成（平倉已成、新倉沒開成）、unfilled **沒有成交**（進場單送出了卻一點都沒成交、已撤單，沒有開倉也沒有平倉——不是失敗）、" +
	"notPlaced 沒有下單、abandoned 送出前就放棄（太晚或使用者煞車）；reason 是交易服務說的原因。" +
	"另有 action（動作）、平倉的 closedQuantity／closeAveragePrice、開倉的 openedDirection／openedQuantity／openAveragePrice、" +
	"stopLossPrice／takeProfitPrice，以及 protectionMissing（true 表示倉位開著卻沒掛上應有的止損或止盈，要提醒使用者）。" +
	"\n\n**開倉的那一輪另帶進場單的細節**（數字都是字串形式的精確小數）：referencePrice 參考價、entryLimitPrice 進場限價、" +
	"entrySlippageTolerancePercentage 當時用的容忍度、requestedQuantity 要求數量、entryExecutedQuantity 已成交數量、" +
	"entrySlippagePercentage 相對參考價的滑價百分比——**正數表示對使用者不利**（做多買貴了、做空賣便宜了），負數表示比參考價更好；" +
	"有成交才有滑價。entryStatus 是進場單狀態：working 等待成交（最多等一個觸發間隔）、filled 全部成交、" +
	"partiallyFilled 部分成交（其餘已撤單，例如「部分成交 0.001／0.002」）、unfilled 沒有成交（已撤單）。" +
	"沒有參考價可算的那一輪沒有這些欄位；**這個功能上線前的紀錄也沒有**——沒有不代表那一輪沒下單"

func strategyBotApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_strategy_bot",
			"建立一台策略機器人：讓一份交易策略在 1 到 10 個交易標的（監聽標的）上定時自己跑。"+
				"\n\n**機器人分兩種，建立當下就定了**：現貨機器人（marketDataKind 不給或 kCandle）引用吃 K 線的交易策略、盯現貨標的；"+
				"合約機器人（marketDataKind 為 contractKCandle）引用吃合約行情的交易策略、盯合約追蹤名單上的合約標的，"+
				"每一輪照那份交易策略的交易模式說出做多、做空、平多或平空，並可以在 positionPlan 設槓桿。"+
				"合約機器人的止損止盈要對帳就用 trading_backtest_contract_trading_strategy 重演。"+
				"\n\n**建立不等於啟動。** 建好之後要用 trading_start_strategy_bot 才會開始跑。"+
				"\n\n**填了 positionPlan 的機器人會在訊息裡建議止損與止盈價位，"+
				"而回測只在你把那兩個距離填進去時才算它們。**"+
				"所以一套「回測賺 25%」的規則配上一組停損，"+
				"那兩件事**還沒有對過帳**——"+
				"要對帳就用 trading_backtest_trading_strategy（合約機器人用 trading_backtest_contract_trading_strategy）"+
				"的 stopLossPercentage 與 takeProfitPercentage 再重演一次。"+
				"在那之前，不要拿那個 25% 替這組停損背書。"+
				"\n\n**新建的機器人自動下單一律是關的**，這裡沒有自動下單這一格。"+
				strategyBotPerSymbolExposureNote+strategyBotReplyNote+
				binanceAutoOrderWebOnlyNote+orderFlowReachNote,
			vo.RequestVerbSubmit, "/strategy-bots",
			strategyBotWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_strategy_bots",
			"列出你的每一台策略機器人，含它現在是不是在跑。"+
				"每一台都帶著 marketDataKind，看得出它是現貨機器人（kCandle）還是合約機器人（contractKCandle）。"+
				strategyBotReplyNote,
			vo.RequestVerbRead, "/strategy-bots",
			queryParameter("marketDataKind", vo.ToolParameterKindString,
				"只列其中一種：kCandle（現貨機器人）或 contractKCandle（合約機器人）。不給就全部列出；其他值會被拒絕", false),
		),
		domains.NewApiToolDomain(
			"trading_get_strategy_bot",
			"讀一台策略機器人的設定與目前狀態，含它吃哪一種行情（marketDataKind）；"+
				"合約機器人的 positionPlan 另帶著它的槓桿倍數（leverage）。"+
				strategyBotReplyNote,
			vo.RequestVerbRead, "/strategy-bots/{id}",
			pathParameter("id", "策略機器人識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_strategy_bot",
			"改一台你自己的策略機器人。這是整份改寫：沒帶到的欄位會變成空的。"+
				"**只有已停止的機器人改得動**：在跑的會被拒絕，先用 trading_stop_strategy_bot 停掉、改完再用 trading_start_strategy_bot 啟動。"+
				"\n\n**例外有三樣，不給就是保留原本的**：marketDataKind（只改名字或間隔時不必重帶它，換成另一種會被拒絕）、"+
				"autoOrderEntrySlippageTolerancePercentage（進場滑價容忍度，不給不會變成 0，也不會回到 0.1）、"+
				"以及自動下單（這裡沒有那一格，改完維持原本的開或關）。"+
				"合約機器人的 positionPlan 是整組改寫，沒帶 leverage 就回到一倍。"+
				"\n\n**symbols 是修改後的完整清單**：只改名字、間隔或部位規劃時也要把原本的標的照抄回去，漏列一個就是把它移除。"+
				"**移除一個標的被拒時整次修改什麼都不改**：reason 為 watchedSymbolHoldsPosition 是那個標的還有機器人持倉，要先讓它空手才能移除；"+
				"reason 為 watchedSymbolAutoOrderOutstanding 是那個標的還有沒做完的自動下單，要等它做完才能移除。"+
				"遇到時照原話告訴使用者，**不要重送同一份**；要改別的設定就把那個標的留在清單裡再送。"+
				strategyBotPerSymbolExposureNote+strategyBotReplyNote+
				binanceAutoOrderWebOnlyNote+orderFlowReachNote,
			vo.RequestVerbReplace, "/strategy-bots/{id}",
			append([]vo.ToolParameterVo{pathParameter("id", "要改哪一台")},
				strategyBotWriteParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_strategy_bot",
			"刪掉一台你自己的策略機器人。"+
				"\n\n**這台機器人還有一張自動下單的進場單掛在幣安上時，刪除會被拒絕**，拒絕的 reason 是 autoOrderEntryOutstanding——"+
				"刪了就沒有人撤那張單，它之後成交會是一個沒有止損的倉位。"+
				"遇到時告訴使用者：請他**自己到網頁上這台機器人的詳細頁關掉自動下單**，等進場單撤掉（最多一個觸發間隔）後再刪；"+
				"這個外掛不能替他關，不要重送同一個刪除。"+
				binanceAutoOrderWebOnlyNote,
			vo.RequestVerbRemove, "/strategy-bots/{id}",
			pathParameter("id", "要刪哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_start_strategy_bot",
			"啟動一台策略機器人，它開始按自己的間隔一輪一輪跑，每一輪把交易策略套在每一個監聽標的上。"+
				strategyBotReplyNote,
			vo.RequestVerbSubmit, "/strategy-bots/{id}/power",
			pathParameter("id", "要啟動哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_stop_strategy_bot",
			"停掉一台策略機器人。已經跑過的那些輪次紀錄一筆都不刪。停掉不會關掉它的自動下單。"+
				strategyBotReplyNote,
			vo.RequestVerbRemove, "/strategy-bots/{id}/power",
			pathParameter("id", "要停哪一台"),
		),
		domains.NewApiToolDomain(
			"trading_list_strategy_bot_runs",
			"看一台策略機器人跑過哪幾輪，以及每一輪做了什麼決定。"+
				"這是它有沒有在做事的唯一證據——只看它「在跑」不代表它有在做決定。"+
				"\n\n**每一筆紀錄帶著 symbol，說明它是哪一個監聽標的的那一輪**；每一輪每個參與的標的各記一筆。"+
				"輪次編號（runNumber）在同一台機器人的各標的之間不重複，所以記到交易日誌的「機器人＋第幾輪」照舊指得到唯一一筆。"+
				"每個標的各記得最近 50 輪；被移除的標的過去的紀錄仍在。使用者只問某一個標的時給 symbol 篩選。"+
				contractStrategyBotRunDetailsNote+contractAutoOrderResultNote+contractStrategyBotSkippedRoundNote,
			vo.RequestVerbRead, "/strategy-bots/{id}/runs",
			pathParameter("id", "要看哪一台"),
			queryParameter("symbol", vo.ToolParameterKindString,
				"只看這一個監聽標的的紀錄（大小寫不拘），由新到舊。不給就是全部標的；格式不對會被拒絕", false),
		),
		domains.NewApiToolDomain(
			"trading_run_strategy_bot_now",
			"叫一台策略機器人立刻跑一輪，不等它的間隔到。"+
				"\n\n用來在改完交易策略之後馬上看一眼它現在會做什麼決定，而不必等下一輪。"+
				"回應是這台機器人跑完之後的樣子，不是這一輪的紀錄——跑完用 trading_list_strategy_bot_runs 讀這一輪的紀錄。"+
				contractStrategyBotSkippedRoundNote+contractAutoOrderResultNote+strategyBotReplyNote,
			vo.RequestVerbSubmit, "/strategy-bots/{id}/runs",
			pathParameter("id", "要叫哪一台立刻跑"),
		),
	}
}

func telegramDeliveryApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_get_telegram_delivery",
			"讀你目前的電報通知設定。",
			vo.RequestVerbRead, "/users/me/telegram-delivery",
		),
		domains.NewApiToolDomain(
			"trading_save_telegram_delivery",
			"設定用哪一台電報機器人、送到哪個對話。設過了再設一次即覆蓋。"+
				"\n\n設定之後建議用 trading_send_telegram_test_message 送一則測試，"+
				"確認真的收得到——填錯的 chatId 不會有任何錯誤，只是安靜地送不到。",
			vo.RequestVerbReplace, "/users/me/telegram-delivery",
			bodyParameter("botToken", vo.ToolParameterKindString, "電報機器人的憑證", true),
			bodyParameter("chatId", vo.ToolParameterKindString, "要送到哪個對話", true),
		),
		domains.NewApiToolDomain(
			"trading_remove_telegram_delivery",
			"移除電報通知設定，之後不再送任何通知。",
			vo.RequestVerbRemove, "/users/me/telegram-delivery",
		),
		domains.NewApiToolDomain(
			"trading_send_telegram_test_message",
			"用目前的設定送一則測試訊息，確認真的收得到。",
			vo.RequestVerbSubmit, "/users/me/telegram-delivery/test-message",
			bodyParameter("message", vo.ToolParameterKindString, "要送什麼內容", true),
		),
	}
}
