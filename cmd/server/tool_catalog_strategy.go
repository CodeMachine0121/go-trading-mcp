package main

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

func strategyScriptApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_strategy_script",
			"建立一支屬於你的策略腳本：一個名字、一段算式、它產出什麼形狀，以及它自己的旋鈕。"+
				"\n\n**要多粗、要幾根、算到什麼時候都不記在策略腳本身上**——那是每一次執行的事。"+
				"所以同一支「二十根均線」可以在一小時的刻度上看一次、再在一分鐘的刻度上看一次，不必存成兩支。"+
				"\n\n**吃哪一種行情（marketDataKind）才是記在它身上的**，而且建立當下就定了、之後不能換。"+
				contractKCandleScriptNote,
			vo.RequestVerbSubmit, "/strategy-scripts", true,
			strategyScriptWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_strategy_scripts",
			"列出你用得到的每一支策略腳本：你自己的，加上你從市集採用的。"+
				"每一支都帶著 marketDataKind——它吃 K 線還是合約行情，決定它能拿去哪一種指標計算。",
			vo.RequestVerbRead, "/strategy-scripts", true,
		),
		domains.NewApiToolDomain(
			"trading_get_strategy_script",
			"讀一支策略腳本的完整內容，含算式本身、它宣告的旋鈕，以及它吃哪一種行情（marketDataKind）。"+
				"\n\n讀得到的是你自己的、你採用過的、以及上架在市集的。別人的且沒上架的會被拒絕。",
			vo.RequestVerbRead, "/strategy-scripts/{id}", true,
			pathParameter("id", "策略腳本識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_strategy_script",
			"改一支**你自己的**策略腳本。這是整份改寫：沒帶到的欄位會變成空的，不是保留原值。"+
				"\n\n**唯一的例外是 marketDataKind**：不給就是保留原本的行情種類，照抄原本的也可以；"+
				"要換成另一種會被拒絕（行情種類建立後不得更換），要吃另一種請另建一支。"+
				"\n\n採用自市集的那些改不動——它們是別人的。"+
				contractKCandleScriptNote,
			vo.RequestVerbReplace, "/strategy-scripts/{id}", true,
			append([]vo.ToolParameterVo{pathParameter("id", "要改哪一支")},
				strategyScriptWriteParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_strategy_script",
			"刪掉一支你自己的策略腳本。",
			vo.RequestVerbRemove, "/strategy-scripts/{id}", true,
			pathParameter("id", "要刪哪一支"),
		),
		domains.NewApiToolDomain(
			"trading_publish_strategy_script",
			"把一支你自己的策略腳本上架到市集，讓別人看得到、採用得了。",
			vo.RequestVerbSubmit, "/strategy-scripts/{id}/publication", true,
			pathParameter("id", "要上架哪一支"),
		),
		domains.NewApiToolDomain(
			"trading_withdraw_strategy_script",
			"把一支策略腳本從市集下架。已經採用它的人不受影響。",
			vo.RequestVerbRemove, "/strategy-scripts/{id}/publication", true,
			pathParameter("id", "要下架哪一支"),
		),
		domains.NewApiToolDomain(
			"trading_browse_marketplace",
			"瀏覽市集上每一支上架中的策略腳本。列表給的是名字與說明——"+
				"說明是讀者唯一看得到的東西，算式本身要採用或讀取才看得到。"+
				"每一支都帶著 marketDataKind，看得出它吃 K 線還是合約行情。",
			vo.RequestVerbRead, "/marketplace/strategy-scripts", true,
		),
		domains.NewApiToolDomain(
			"trading_adopt_strategy_script",
			"採用市集上的一支策略腳本，之後它就出現在你用得到的清單裡，可以拿去計算與回測。"+
				"\n\n採用不是複製：它仍然是作者的，作者改了你就跟著改，而你改不動它。",
			vo.RequestVerbSubmit, "/marketplace/strategy-scripts/{id}/adoption", true,
			pathParameter("id", "要採用哪一支"),
		),
		domains.NewApiToolDomain(
			"trading_abandon_strategy_script",
			"放棄一支採用過的策略腳本，它就從你用得到的清單裡消失。你自己的腳本不受影響。",
			vo.RequestVerbRemove, "/marketplace/strategy-scripts/{id}/adoption", true,
			pathParameter("id", "要放棄哪一支"),
		),
	}
}

// tradingStrategyWriteParameters are what a trading strategy is made of. Shared by
// creating one and rewriting one.
//
// The two conditions arrive nested exactly as a person builds them, because that is
// the shape they are thought in. A flat list with parent references would mean the
// caller flattens a tree that the trading service immediately rebuilds — two chances
// to disagree about one condition.
func tradingStrategyWriteParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("name", vo.ToolParameterKindString, "這份交易策略叫什麼", true),
		bodyParameter("signalSources", vo.ToolParameterKindArray,
			"這份策略要跑哪幾支策略腳本，每個為 "+
				"{\"label\":\"短均線\", \"strategyScriptId\":1, \"aggregationInterval\":\"1h\", "+
				"\"parameterValues\":[{\"name\":\"期數\",\"value\":20}]}。"+
				"label 是你在買賣條件裡怎麼稱呼這一路信號", true),
		bodyParameter("buyCondition", vo.ToolParameterKindObject,
			"什麼時候買。巢狀結構，兩種節點擇一："+
				"比較節點 {\"sourceLabel\":\"短均線\", \"signal\":\"buy\"}，"+
				"或群組節點 {\"operator\":\"and\", \"conditions\":[…]}（operator 為 and／or）", true),
		bodyParameter("sellCondition", vo.ToolParameterKindObject,
			"什麼時候賣。形狀與 buyCondition 完全相同", true),
		// Both left blank are forwarded as nothing at all, for the reason the strategy
		// script's own kind is: a blank means the spot K candle on a create and "keep
		// what is there" on a rewrite, and only the trading service knows which.
		bodyParameter("marketDataKind", vo.ToolParameterKindString,
			"這份交易策略的信號吃哪一種行情，二選一：kCandle（現貨 K 線）或 contractKCandle（永續合約的合約行情格）。"+
				"**建立時不給就是 kCandle**；**修改時不給就是保留原本的**。**建立後不得更換**。"+
				"**每一個信號來源指名的策略腳本都要吃同一種行情**，混進另一種會整份被拒絕，並說出是哪一個來源", false),
		bodyParameter("tradingMode", vo.ToolParameterKindString,
			"**只有吃合約行情的交易策略才有**：它的買賣照哪一種規則讀，三選一——"+
				"longShort（多空反手：賣出時持多倉就平掉並同一棒反手開空）、"+
				"longOnly（只做多：賣出只平多倉）、shortOnly（只做空：買入只平空倉）。"+
				"不給就是 longShort；修改時可以換。**吃 K 線的交易策略沒有交易模式，給了會被拒絕**", false),
	}
}

func tradingStrategyApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_trading_strategy",
			"建立一份交易策略：把幾支策略腳本當成信號來源，再用買賣條件把它們的信號組合成決定。"+
				"\n\n**它吃哪一種行情（marketDataKind）建立當下就定了**：吃 K 線的拿去 trading_backtest_trading_strategy 重演、"+
				"可以掛上策略機器人；吃合約行情的拿去 trading_backtest_contract_trading_strategy 重演，"+
				"並且記著自己的交易模式（tradingMode），可以掛上**合約機器人**（marketDataKind 為 contractKCandle 的策略機器人）。"+
				"機器人吃的行情必須與交易策略相同，兩種不混用。",
			vo.RequestVerbSubmit, "/trading-strategies", true,
			tradingStrategyWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_trading_strategies",
			"列出你的每一份交易策略。",
			vo.RequestVerbRead, "/trading-strategies", true,
		),
		domains.NewApiToolDomain(
			"trading_get_trading_strategy",
			"讀一份交易策略的完整內容，含信號來源、兩個條件樹、它吃哪一種行情（marketDataKind），"+
				"吃合約行情的另外帶著它的交易模式（tradingMode）。",
			vo.RequestVerbRead, "/trading-strategies/{id}", true,
			pathParameter("id", "交易策略識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_trading_strategy",
			"改一份你自己的交易策略。這是整份改寫：沒帶到的欄位會變成空的。"+
				"\n\n**例外是 marketDataKind 與 tradingMode**：不給就是保留原本的。"+
				"marketDataKind 換成另一種會被拒絕；吃合約行情的那一種可以換交易模式（tradingMode），"+
				"只改名字或條件時不必重帶它。",
			vo.RequestVerbReplace, "/trading-strategies/{id}", true,
			append([]vo.ToolParameterVo{pathParameter("id", "要改哪一份")},
				tradingStrategyWriteParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_trading_strategy",
			"刪掉一份你自己的交易策略。掛在它上面的策略機器人會受影響，刪之前先確認。",
			vo.RequestVerbRemove, "/trading-strategies/{id}", true,
			pathParameter("id", "要刪哪一份"),
		),
	}
}

// contractKCandleScriptNote is how to write a strategy script that eats the perpetual
// contract, said once for every ability that has the assistant write or run one.
//
// One copy for the reason the replay notes have one: an assistant reading two
// wordings of one shape writes a script that fits only one of them. It names every
// figure by the name the script reads it under, because a figure the assistant cannot
// name is a figure it guesses — and a wrong guess is not refused, it simply does not
// compile, which reads as "the script is broken" about a script that is not.
//
// The last paragraph is the half that cannot be learnt from the boxes: where a contract
// script can go — a contract bot included — and that it is never rewritten into a spot
// one to fit a spot bot.
const contractKCandleScriptNote = "\n\n**吃合約行情（marketDataKind 為 contractKCandle）的算式**，入口是 " +
	"func Calculate(data []indicator.ContractKCandle) <依 resultType 而定>——照現貨的寫法收 []indicator.KCandle 會算不動。" +
	"每一格是一個走完的刻度區間，**現貨 K 線有的每一項這裡都有、而且同名**" +
	"（Symbol、OpenTimeUnixSeconds、Open、High、Low、Close、Volume、QuoteVolume、TakerBuyBaseVolume、TakerBuyQuoteVolume），" +
	"所以讀收盤價、成交量的那幾行不必改。另外多了：" +
	"TradeCount（成交筆數，int64）；" +
	"Mark、Index、PremiumIndex（標記價格、指數價格、溢價指數，各是一個 indicator.PriceLine，有 Open／High／Low／Close，溢價指數可以是負的）；" +
	"FundingRate（收盤時現行的資金費率，正的是做多付給做空）與 FundingSettledInBar（這一格內有沒有真的結算）；" +
	"OpenInterest、OpenInterestValue（持倉量、持倉價值）、" +
	"AccountLongShare、AccountShortShare、AccountLongShortRatio（多空人數比）、" +
	"TopTraderPositionLongShare、TopTraderPositionShortShare、TopTraderPositionLongShortRatio（大戶多空持倉比）。" +
	"\n\n**資金費率每一格都延續上一次結算的費率，只有 FundingSettledInBar 為真的那一格才是真的收付**——" +
	"把每一格的 FundingRate 加總，算出來的是一個從來沒有人付過的數字。" +
	"\n\n**沒有值一律是零，分不出「沒錄到」與「真的是零」**：舊資料沒有指數價格與溢價指數（那一格整組為零）、" +
	"第一次結算之前沒有費率、持倉統計**只有錄到或同步過的那段才有值**而且要夠新（沒有就整組為零；" +
	"更早的用 trading_sync_contract_k_candle_history 補）。算式要自己判斷，例如持倉量為零多半是沒錄到。" +
	"\n\n**吃合約行情的策略腳本用在合約那一邊**：trading_calculate_contract_indicator 算指標、" +
	"trading_backtest_contract_strategy_script 在合約帳戶上重演、當吃合約行情的交易策略的信號來源。" +
	"現貨的指標計算、現貨重演、吃 K 線的交易策略都會拒絕它。" +
	"**要讓它常駐盯盤，就把用它的合約交易策略掛上一台合約機器人**（trading_create_strategy_bot 的 marketDataKind 給 contractKCandle）——" +
	"不要替使用者改寫成一支吃 K 線的、掛上現貨機器人去湊，那盯的是另一種商品。"

// costedReportCardNote is how to read a report card that had the fees taken out of
// it, said once for both replays.
//
// Both of them need it and it has to be the same sentence in both: it corrects a
// reading the assistant already has. Two copies would be two places to improve the
// wording, and the day only one of them improved, one replay would be telling the
// assistant that profit is gross while the other said it is net — and the numbers
// themselves look identical either way.
const costedReportCardNote = "\n\n**填了交易成本時，成績單多一個 totalTransactionCost**——這次總共付掉多少。有了它才答得出「這支策略是抓價差不行，還是被手續費吃掉」，而同一個報酬率本來講得出這兩個完全不同的故事。**每一筆交易的 profit 已經是扣掉成本後的淨額，勝率也是照淨額算的**——價差賺得到、卻賺不過手續費的那一趟不算贏，別把它讀成賺錢的交易。"

// spotOnlyReplayNote is what the two spot replays actually do, said once for both.
//
// One copy for the reason the costed note has one: the two replays must never tell an
// assistant two different things about the same replay, and two wordings are two
// chances for only one of them to get improved.
//
// It says what these two do **and** where the rest went, because the second half is
// the one an assistant cannot find out by reading the boxes. A missing box reads as
// "not supported yet, try another way" — so somebody asking to short would get a spot
// strategy bent into shape rather than the contract replay. The last sentence is there
// to stop exactly that.
const spotOnlyReplayNote = "\n\n**這兩件只重演現貨。** 買入時空手就開倉、已經有倉位就當作沒聽到；賣出就平倉把錢收回來、之後空手等下一個買點；空手時聽到賣出什麼都不做。\n\n**這兩件沒有交易模式可以指定，也開不了槓桿**——借錢、做空與強制平倉在合約帳戶上，所以這裡也不會有強制平倉這種出場。吃合約行情的策略腳本或交易策略拿來這裡會被拒絕。\n\n**使用者提到要放空、要開槓桿、或說他在合約帳戶上操作時，改用合約重演**（trading_backtest_contract_strategy_script、trading_backtest_contract_trading_strategy）——不要在這裡換一組設定去湊：湊出來的成績單是照他沒做的操作算的，而他不會發現。"

// shortTermReplayNote is how to replay for short-term work and how to read what comes
// back, said once for all four replays.
//
// One copy for the reason the other notes have one. It leads with the two habits that
// decide whether a short-term report card means anything — filling at the next open,
// and judging only by a part the tuning never saw — because an assistant left to
// itself does neither, and nothing in the numbers says so.
const shortTermReplayNote = "\n\n**研發短線策略時，fillTiming 用 nextOpen**：收盤成交（close，不給即是）是讓說出信號的那一格在它自己的收盤價成交，而收盤那一刻其實已經過去了——短線賺的常常就是那一點點，所以收盤成交的成績單必然偏樂觀。下一格開盤成交時，最後一格說出的信號不會成交。" +
	"\n\n**樣本外：給 validationStartTime，把這段期間切成調參段（之前）與驗證段（之後）。**回來會多 inSample 與 validation 兩份結果（形狀與整段相同），兩段各自從初始資金、空手開始，算式在驗證段看得到之前的歷史。" +
	"**只拿 inSample 調參數，只拿 validation 判斷這支策略好不好**；看過 validation 之後又回頭改參數，那一段就不再是樣本外，請換一段更晚、還沒看過的期間當驗證段。" +
	"驗證起點必須落在期間之內、兩段都要至少一格，否則會被拒絕。" +
	"\n\n**成績單多五格**：profitFactor（賺的那幾筆淨損益合計 ÷ 虧的那幾筆合計，一筆都沒虧時是 null——不是無限大）、expectancy（每一筆平均淨賺多少）、averageHoldingSeconds（平均持倉秒數）、maximumConsecutiveLossCount（最多連虧幾筆，打平會打斷）、costToGrossProfitRatio（手續費吃掉扣成本前價差的比例，0.25 就是吃掉四分之一；價差本身沒賺時是 null；合約的資金費用不算在內）。結束時還開著的那一注不算進這五格。" +
	"**短線策略要看 expectancy 與 costToGrossProfitRatio**：勝率高而 expectancy 接近零、或成本吃掉大半價差，都是交易次數多、沒有真的優勢。" +
	"\n\n**一次重演有整體的時間上限**：跑太久會整次被拒絕（不會給半張成績單），請縮短期間或改用粗一點的刻度再試。" +
	"\n\n**長重演交到你手上之前會先精簡**：資金曲線超過兩百點時平均取兩百點（頭尾都在），並多一個 equityCurvePointTotalCount 說原本幾點；交易明細超過一百筆時只列最近一百筆，並多一個 closedTradeTotalCount 說總共幾筆；inSample 與 validation 各自照做。**成績單的每一個數字都沒有動**——筆數、勝率請讀成績單，不要數交易明細。"

func backtestApiTools(replayWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_backtest_strategy_script",
			"拿**一支策略腳本**重演一段已經發生過的行情。"+
				"\n\n**要嘛指名一支既有的策略腳本（strategyScriptId），要嘛自己帶一段算式（script），兩者只能挑一個。**"+
				"\n\n重演一律以 signal 這種指標值種類執行——它讀的就是每根一個買賣信號，"+
				"所以這裡沒有 resultType 可填、也不需要填。"+
				"\n\n回來的是成績單、交易明細與資金曲線。"+
				"**成績單裡的交易筆數一定要看**——一張幾乎沒有交易的漂亮成績單會被讀成「很穩」，"+
				"而真相是這份策略根本沒有在做決定。"+
				"\n\n模擬了出場價位時，**stopLossExitCount 也一定要看**："+
				"十次出場八次是被停損掃出去的策略，與十次都靡訊號出場的，"+
				"報酬率可以一模一樣——而前者是停損在支撑它，"+
				"後者是還沒遇到那個掃光它的盤。每一筆交易自己也帶著 exitReason。"+
				costedReportCardNote+
				spotOnlyReplayNote+
				shortTermReplayNote,
			vo.RequestVerbSubmit, "/backtests", true,
			append(append([]vo.ToolParameterVo{
				bodyParameter("strategyScriptId", vo.ToolParameterKindInteger,
					"要重演哪一支既有的策略腳本。與 script 擇一", false),
			}, backtestParameters()...),
				bodyParameter("aggregationInterval", vo.ToolParameterKindString,
					"彙總刻度，六選一：1m／5m／15m／1h／4h／1d", false),
				bodyParameter("script", vo.ToolParameterKindString,
					"一段還沒存起來的算式。與 strategyScriptId 擇一", false),
				bodyParameter("parameters", vo.ToolParameterKindArray,
					"自帶算式時它宣告的旋鈕", false),
				bodyParameter("parameterValues", vo.ToolParameterKindArray,
					"這一次要把旋鈕調成多少，每個為 {\"name\":…, \"value\":…}。只用於這次重演，不寫回腳本", false),
			)...,
		).Waiting(replayWaitLimit).CondensingReplayResults(),
		domains.NewApiToolDomain(
			"trading_backtest_trading_strategy",
			"拿**一份交易策略**重演一段已經發生過的行情。"+
				"\n\n信號來源與買賣兩個條件都取自那份交易策略本身，這裡不必也不能再說一次。"+
				"\n\n這一支與 trading_backtest_strategy_script 的差別："+
				"那一支重演的是單獨一支算式產出的信號，這一支重演的是幾支信號組合出來的決定。"+
				costedReportCardNote+
				spotOnlyReplayNote+
				shortTermReplayNote,
			vo.RequestVerbSubmit, "/trading-strategies/{id}/backtests", true,
			append([]vo.ToolParameterVo{pathParameter("id", "要重演哪一份交易策略")},
				backtestParameters()...)...,
		).Waiting(replayWaitLimit).CondensingReplayResults(),
	}
}

// contractBacktestParameters are the account rules a contract replay trades by, shared
// by replaying a contract strategy script and replaying a contract trading strategy.
//
// It is the spot list and two more boxes, not a copy of it: everything a spot replay is
// told a contract replay is told in the same words, because the trading service reads
// them by the same rules. The trading mode is not here — only the script replay has one
// to give, since a contract trading strategy says its own.
func contractBacktestParameters() []vo.ToolParameterVo {
	return append(backtestParameters(),
		bodyParameter("leverage", vo.ToolParameterKindString,
			"槓桿倍數（字串形式的精確小數）：名目 ＝ 押下去的保證金 × 這個倍數。"+
				"**不給或 0 就是一倍**——合約帳戶上的一倍仍然是一筆合約部位，做空一倍一樣會在價格翻倍時被強制平倉。"+
				"小於一會被拒絕；超過這個標的分級允許的最高槓桿也會被拒絕（會說出上限）。"+
				"開倉當下名目所在那一級不允許這麼高的槓桿時，那一次開倉被擋下、記在 blockedOpeningCount", false),
		bodyParameter("slippagePercentage", vo.ToolParameterKindString,
			"每一次成交往不利方向偏幾個百分點（0.05 就是 0.05%）：買進成交得貴一點、賣出成交得便宜一點，"+
				"信號進出、止損、止盈都算，強制平倉不算。**不給就是不計**。負的與超過 100 會被拒絕", false),
	)
}

// contractAccountReplayNote is how a contract account is replayed and how its report
// card reads, said once for both contract replays.
//
// One copy for the reason the spot notes have one. It is long because every sentence
// is something the assistant would otherwise read wrong off the numbers: a liquidated
// trade looks like a large ordinary loss, a funding total looks like a fee, and a
// strategy the venue kept refusing looks like a cautious one.
const contractAccountReplayNote = "\n\n**這是在逐倉合約帳戶上重演**：每一注押下去的是保證金，承擔的是名目（保證金 × 槓桿），賺賠照數量 × 價差算，**一注最多賠光它自己的保證金**，可用資金不會變成負的。" +
	"\n\n**交易成本在這裡照名目收，不是照押下去的保證金**：entryCostPercentage 那一格說的「押下去的金額」在合約帳戶上指的是名目，五倍槓桿時一趟手續費是保證金的五倍那麼多；押全部時會自己留出照名目算的進場成本。" +
	"\n\n**強制平倉看標記價格，不是最新價**；維持保證金照開倉當下名目所在的那一級分級算，**沒有分級時退回交易規格最小那一級**（大部位的強平價會被算得太遠）。分級只有今天那一組，重演過去也用它——成績單的 maintenanceMarginBasis 會說出是哪一種、何時確認的。" +
	"\n\n**資金費率一律計入，不能關**：帶著倉位走過的每一次結算都照數量 × 標記價格 × 費率收付（正的費率做多付、做空收），直接進出那一注的保證金，**所以付了費率強平價會往進場價靠近**。在某一格收盤才開的倉不付那一格內的結算。" +
	"\n\n一格裡的順序是：先收付資金費率，再看止損與強平（**離進場價近的先到**），再看止盈，最後才照這一格的信號在收盤成交。同一格同時碰到兩邊一律算不利的那一側。數量照交易規格的數量步進往下取整，止損止盈價對齊價格跳動單位；低於最小下單量或最小名目的開倉被擋下。" +
	"\n\n**讀成績單時一定要看**：liquidationExitCount（被強平幾筆，每一筆 exitReason 為 liquidation、profit 是整筆保證金加進場成本的損失）、totalFundingFee（淨付出的資金費用，負的是淨收入）、longTradeCount／longWinRate 與 shortTradeCount／shortWinRate（多空分開，沒有那一邊的勝率是 null）、blockedOpeningCount（被交易規則擋下的開倉——一張幾乎沒有交易的成績單可能是一直被擋，不是很穩）。每一筆交易帶著 direction、leverage、quantity、margin、fundingFee。" +
	"\n\n**會被拒絕的情況**：這個合約標的還沒有交易規格（要先加入合約追蹤名單）；槓桿小於一或超過上限；滑點為負或超過 100；另外指定 maintenanceMarginRate（它由分級決定）；湊不出兩格。" +
	"\n\n**多空反手一旦進場就一直在場內**，只有止損、止盈或強平能讓它回到空手——使用者想要「平掉但不反手」時，請他改用 longOnly 或 shortOnly。"

func contractBacktestApiTools(replayWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_backtest_contract_strategy_script",
			"拿**一支吃合約行情的策略腳本**在**合約帳戶**上重演一段已經發生過的永續合約行情：可多可空、借得到錢、會被強制平倉、要付資金費率。"+
				"\n\n**要嘛指名一支既有的策略腳本（strategyScriptId），要嘛自己帶一段算式（script），兩者只能挑一個。**"+
				"指名的那一支必須吃合約行情（marketDataKind 為 contractKCandle），吃 K 線的會被拒絕，那一支要用 trading_backtest_strategy_script。"+
				"重演一律以 signal 執行，算式收 []indicator.ContractKCandle、回 indicator.Signal。"+
				"\n\n**交易模式（tradingMode）由這一次說**，三選一：longShort（預設，多空反手）、longOnly（只做多）、shortOnly（只做空）；"+
				"**沒有 spot 這一種**——現貨的事用 trading_backtest_strategy_script。"+
				costedReportCardNote+
				contractAccountReplayNote+
				shortTermReplayNote,
			vo.RequestVerbSubmit, "/contract-backtests", true,
			append(append([]vo.ToolParameterVo{
				bodyParameter("strategyScriptId", vo.ToolParameterKindInteger,
					"要重演哪一支既有的、吃合約行情的策略腳本。與 script 擇一", false),
			}, contractBacktestParameters()...),
				bodyParameter("tradingMode", vo.ToolParameterKindString,
					"這一次照哪一種規則交易：longShort（不給即是；買入開多或把空倉反手成多、賣出開空或把多倉反手成空）、"+
						"longOnly（只做多，賣出只平倉）、shortOnly（只做空，買入只平倉）。其他值（含 spot）會被拒絕", false),
				bodyParameter("aggregationInterval", vo.ToolParameterKindString,
					"彙總刻度，六選一：1m／5m／15m／1h／4h／1d", false),
				bodyParameter("script", vo.ToolParameterKindString,
					"一段還沒存起來、吃合約行情的算式。與 strategyScriptId 擇一", false),
				bodyParameter("parameters", vo.ToolParameterKindArray,
					"自帶算式時它宣告的旋鈕", false),
				bodyParameter("parameterValues", vo.ToolParameterKindArray,
					"這一次要把旋鈕調成多少，每個為 {\"name\":…, \"value\":…}。只用於這次重演，不寫回腳本", false),
			)...,
		).Waiting(replayWaitLimit).CondensingReplayResults(),
		domains.NewApiToolDomain(
			"trading_backtest_contract_trading_strategy",
			"拿**一份吃合約行情的交易策略**在**合約帳戶**上重演一段已經發生過的永續合約行情。"+
				"\n\n信號來源、買賣兩個條件、**交易模式**都取自那份交易策略本身，這裡不必也不能再說一次——"+
				"這一件**沒有 tradingMode 那一格**，要換交易模式請用 trading_update_trading_strategy 改那份交易策略。"+
				"吃 K 線的交易策略會被拒絕，那一份要用 trading_backtest_trading_strategy。"+
				"\n\n成績單多一個 conflictedCandleCount：買賣條件同時成立、當作持平的格數。"+
				costedReportCardNote+
				contractAccountReplayNote+
				shortTermReplayNote,
			vo.RequestVerbSubmit, "/trading-strategies/{id}/contract-backtests", true,
			append([]vo.ToolParameterVo{pathParameter("id", "要重演哪一份交易策略")},
				contractBacktestParameters()...)...,
		).Waiting(replayWaitLimit).CondensingReplayResults(),
	}
}
