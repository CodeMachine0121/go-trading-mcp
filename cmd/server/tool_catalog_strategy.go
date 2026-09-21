package main

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

func strategyScriptApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_strategy_script",
			"建立一支屬於你的策略腳本：一個名字、一段算式、它產出什麼形狀，以及它自己的旋鈕。"+
				"\n\n**要多粗、要幾根、算到什麼時候都不記在策略腳本身上**——那是每一次執行的事。"+
				"所以同一支「二十根均線」可以在一小時的刻度上看一次、再在一分鐘的刻度上看一次，不必存成兩支。",
			vo.RequestVerbSubmit, "/strategy-scripts", true,
			strategyScriptWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_strategy_scripts",
			"列出你用得到的每一支策略腳本：你自己的，加上你從市集採用的。",
			vo.RequestVerbRead, "/strategy-scripts", true,
		),
		domains.NewApiToolDomain(
			"trading_get_strategy_script",
			"讀一支策略腳本的完整內容，含算式本身與它宣告的旋鈕。"+
				"\n\n讀得到的是你自己的、你採用過的、以及上架在市集的。別人的且沒上架的會被拒絕。",
			vo.RequestVerbRead, "/strategy-scripts/{id}", true,
			pathParameter("id", "策略腳本識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_strategy_script",
			"改一支**你自己的**策略腳本。這是整份改寫：沒帶到的欄位會變成空的，不是保留原值。"+
				"\n\n採用自市集的那些改不動——它們是別人的。",
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
				"說明是讀者唯一看得到的東西，算式本身要採用或讀取才看得到。",
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
		// Beside the name, because the mode and the name are what this set of rules
		// *is*, while the sources and the two conditions are what it is made of.
		//
		// The description has to map what the person said onto one of the three
		// spellings. An assistant picking this cannot see their broker and cannot see
		// whether the market allows shorting — the sentence they typed is its only
		// clue, so the sentence has to be in here.
		//
		// Three spellings because the box answers two questions at once, and the pair
		// somebody is most likely to describe without naming — long only, on borrowed
		// money — is the one an assistant reaching for the obvious "cannot short"
		// answer gets wrong. Getting it wrong is not cosmetic: that person's bot then
		// cannot be saved with the leverage they are actually running.
		//
		// The venue is the trap. Two of the three run on a perpetual contract account,
		// so "I trade Binance perps" narrows nothing — and the half of the question it
		// leaves open is the half the default answers wrongly. Guessing long-short
		// reverses every sell into a short the person never asked for, and the report
		// card that comes back is entirely plausible, so nothing tells them. That is
		// why this box says to ask rather than to infer.
		bodyParameter("tradingMode", vo.ToolParameterKindString,
			"這份規則是寫給哪一種帳戶的。它答兩個各自獨立的問題——做不做得了空、借不借得到錢——"+
				"而三個取值就是那兩個問題的三種合法組合（做得了空卻借不到錢不存在："+
				"放空本來就要先借到東西才賣得出去）。"+
				"spot 做不了空、也借不到錢（賣出＝平掉回現金，空手時賣出不動作）；"+
				"longShort 做得了空、也借得到錢（賣出＝平掉多倉並反手做空）；"+
				"leveragedLong 做不了空、但借得到錢——進出場與 spot 一字不差，"+
				"差別只有它開得了槓桿。省略即 longShort。"+
				"使用者說他的帳戶不能放空（台股現貨、ETF、多數券商帳戶）時給 spot；"+
				"**說他在合約帳戶上只做多、要上一點槓桿（例如幣安永續開多）時給 leveragedLong**——"+
				"這時給 spot 是錯的，他會連重演都跑不了，機器人也存不進大於 1 倍的槓桿。"+
				"\n\n**場所不決定模式，做不做空才決定。** 他說「我在幣安永續」「我開合約」"+
				"只講了場所——longShort 與 leveragedLong 都跑在那裡，那句話一個都沒排除掉。"+
				"**沒問出他放不放空之前不要猜。** 兩個方向猜錯的代價差很多："+
				"該給 leveragedLong 卻給了 spot，他會被拒絕，至少他知道；"+
				"**該給 leveragedLong 卻給了 longShort（也就是不給，因為那是預設），"+
				"他每一次賣出都會被反手做空**——成績單照樣跑得出來、看起來完全合理，"+
				"而那是一份與他實際會做的事相反的東西，他不會發現。"+
				"不確定就問一句：「賣出的時候，要幫你反手做空，還是把錢收回來等下一個買點？」"+
				"\n\n它決定了重演這份規則時用哪一套算法、這份規則開不開得了槓桿，"+
				"也決定了機器人訊息寫「買入／出場」還是「做多／做空」", false),
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
	}
}

func tradingStrategyApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_create_trading_strategy",
			"建立一份交易策略：把幾支策略腳本當成信號來源，再用買賣條件把它們的信號組合成決定。",
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
			"讀一份交易策略的完整內容，含信號來源與兩個條件樹。",
			vo.RequestVerbRead, "/trading-strategies/{id}", true,
			pathParameter("id", "交易策略識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_trading_strategy",
			"改一份你自己的交易策略。這是整份改寫：沒帶到的欄位會變成空的。",
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

// costedReportCardNote is how to read a report card that had the fees taken out of
// it, said once for both replays.
//
// Both of them need it and it has to be the same sentence in both: it corrects a
// reading the assistant already has. Two copies would be two places to improve the
// wording, and the day only one of them improved, one replay would be telling the
// assistant that profit is gross while the other said it is net — and the numbers
// themselves look identical either way.
const costedReportCardNote = "\n\n**填了交易成本時，成績單多一個 totalTransactionCost**——這次總共付掉多少。有了它才答得出「這支策略是抓價差不行，還是被手續費吃掉」，而同一個報酬率本來講得出這兩個完全不同的故事。**每一筆交易的 profit 已經是扣掉成本後的淨額，勝率也是照淨額算的**——價差賺得到、卻賺不過手續費的那一趟不算贏，別把它讀成賺錢的交易。"

// liquidationReportCardNote is what a borrowed replay adds to the report card, said
// once for both replays.
//
// One copy for the reason the costed note has one: the two replays must never tell
// an assistant two different things about the same report card, and two wordings
// are two chances for only one of them to get improved.
//
// It says why rather than what. "There is a count of liquidations" is something an
// assistant can see in the response; that a respectable return rate can belong to an
// account which was emptied three times on the way is not.
const liquidationReportCardNote = "\n\n**開了槓桿時，成績單多一個 liquidationExitCount**——這次有幾注是被強制平倉打掉的（沒開槓桿時恆為零）。**這一格一定要看**：同一個報酬率講得出兩個完全不同的故事——一個是停損一路擋著、從來沒有真的危險過，另一個是這個帳戶歸零過三次而報酬率是靠剩下那幾筆湊回來的。少了這一格，兩者在成績單上長得一模一樣。每一筆交易的 exitReason 也會寫著 liquidation，看得出是哪幾筆。"

func backtestApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_backtest_strategy_script",
			"拿**一支策略腳本**重演一段已經發生過的行情。"+
				"\n\n**要嘛指名一支既有的策略腳本（strategyScriptId），要嘛自己帶一段算式（script），兩者只能挑一個。**"+
				"\n\n重演一律以 signal 這種指標值種類執行——它讀的就是每根一個買賣信號，"+
				"所以這裡沒有 resultType 可填、也不需要填。"+
				"\n\n回來的是成績單加交易明細，沒有資金曲線（每一點都從交易明細推得回來）。"+
				"**成績單裡的交易筆數一定要看**——一張幾乎沒有交易的漂亮成績單會被讀成「很穩」，"+
				"而真相是這份策略根本沒有在做決定。"+
				"\n\n模擬了出場價位時，**stopLossExitCount 也一定要看**："+
				"十次出場八次是被停損掃出去的策略，與十次都靡訊號出場的，"+
				"報酬率可以一模一樣——而前者是停損在支撑它，"+
				"後者是還沒遇到那個掃光它的盤。每一筆交易自己也帶著 exitReason。"+
				costedReportCardNote+
				liquidationReportCardNote,
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
				// This replay's own box, not one every replay shares. There is no
				// trading strategy here to ask, so the caller is the only one who can
				// say it — which is also why the spellings are written out here. On
				// the other path the mode is read off a stored strategy and the
				// assistant never types it; on this one, a box that names none of
				// them leaves it guessing at a string.
				bodyParameter("tradingMode", vo.ToolParameterKindString,
					"這次重演照哪一套規則交易，三選一。"+
						"longShort 做得了空、也借得到錢（賣出＝平掉多倉並反手做空）；"+
						"spot 做不了空、也借不到錢（賣出＝平掉回現金，空手時賣出不動作）；"+
						"leveragedLong 做不了空、但借得到錢——進出場與 spot 一字不差，"+
						"差別只有它開得了槓桿（合約帳戶只做多就是這一種）。"+
						"**場所不決定模式**：longShort 與 leveragedLong 都跑在永續合約上，"+
						"差的是賣出要不要反手做空。沒問出他放不放空就用預設，"+
						"等於替他把每一次賣出換成一個空倉，而成績單看起來完全合理。"+
						"省略即 longShort，也就是一直留在市場裡。"+
						"認不得的值會整次被拒絕，不會默默用預設的那一個", false),
			)...,
		),
		domains.NewApiToolDomain(
			"trading_backtest_trading_strategy",
			"拿**一份交易策略**重演一段已經發生過的行情。"+
				"\n\n信號來源、買賣兩個條件與**交易模式**都取自那份交易策略本身，"+
				"這裡不必也不能再說一次。"+
				"\n\n**使用者說他的帳戶不能放空時，要改的不是這一次重演**——"+
				"用 trading_update_trading_strategy 把那一份的 tradingMode 改成 spot，之後每一次重演都跟著對。"+
				"\n\n**他是在合約帳戶上只做多的話，要改成 leveragedLong 而不是 spot**："+
				"兩者的進出場一字不差，但 spot 借不到錢，他上線在用的槓桿在這裡會被整份拒絕，"+
				"於是他驗證不了自己真正在做的事。"+
				"\n\n這一支與 trading_backtest_strategy_script 的差別："+
				"那一支重演的是單獨一支算式產出的信號，這一支重演的是幾支信號組合出來的決定；"+
				"而那一支沒有交易策略可問，所以交易模式由你當次指定。"+
				costedReportCardNote+
				liquidationReportCardNote,
			vo.RequestVerbSubmit, "/trading-strategies/{id}/backtests", true,
			append([]vo.ToolParameterVo{pathParameter("id", "要重演哪一份交易策略")},
				backtestParameters()...)...,
		),
	}
}
