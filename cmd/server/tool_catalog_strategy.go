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
			vo.RequestVerbSubmit, "/strategy-scripts",
			strategyScriptWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_strategy_scripts",
			"列出你用得到的每一支策略腳本：你自己寫的，加上你從市集加入的副本（adopted，看不到算式）。"+
				"每一支都帶著 marketDataKind——它吃 K 線還是合約行情，決定它能拿去哪一種指標計算。",
			vo.RequestVerbRead, "/strategy-scripts",
		),
		domains.NewApiToolDomain(
			"trading_get_strategy_script",
			"讀一支策略腳本的完整內容，含算式本身、它宣告的旋鈕，以及它吃哪一種行情（marketDataKind）。"+
				"\n\n讀得到的是你自己的。從市集加入的副本也是你的，但它的算式是作者的，所以讀回來沒有算式。",
			vo.RequestVerbRead, "/strategy-scripts/{id}",
			pathParameter("id", "策略腳本識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_strategy_script",
			"改一支**你自己的**策略腳本。這是整份改寫：沒帶到的欄位會變成空的，不是保留原值。"+
				"\n\n**唯一的例外是 marketDataKind**：不給就是保留原本的行情種類，照抄原本的也可以；"+
				"要換成另一種會被拒絕（行情種類建立後不得更換），要吃另一種請另建一支。"+
				"\n\n從市集加入的副本改不動——算式是作者的。"+
				contractKCandleScriptNote,
			vo.RequestVerbReplace, "/strategy-scripts/{id}",
			append([]vo.ToolParameterVo{pathParameter("id", "要改哪一支")},
				strategyScriptWriteParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_strategy_script",
			"刪掉一支你自己的策略腳本。",
			vo.RequestVerbRemove, "/strategy-scripts/{id}",
			pathParameter("id", "要刪哪一支"),
		),
		domains.NewApiToolDomain(
			"trading_publish_strategy_script",
			"把一支你自己寫的策略腳本上架到市集，讓別人看得到、加入得了。從市集加入的副本不能再上架。",
			vo.RequestVerbSubmit, "/strategy-scripts/{id}/publication",
			pathParameter("id", "要上架哪一支"),
		),
		domains.NewApiToolDomain(
			"trading_withdraw_strategy_script",
			"把一支策略腳本從市集下架。已經加入的人手上的是自己的副本，不受影響。",
			vo.RequestVerbRemove, "/strategy-scripts/{id}/publication",
			pathParameter("id", "要下架哪一支"),
		),
		domains.NewApiToolDomain(
			"trading_browse_marketplace",
			"瀏覽市集上每一支上架中的策略腳本。列表給的是名字與說明——"+
				"說明是讀者唯一看得到的東西，算式除了作者，誰都看不到，加入之後也一樣。"+
				"每一支都帶著 marketDataKind，看得出它吃 K 線還是合約行情。",
			vo.RequestVerbRead, "/marketplace/strategy-scripts",
		),
		domains.NewApiToolDomain(
			"trading_adopt_strategy_script",
			"加入市集上的一支策略腳本：交易服務**複製一份給你**，出現在你用得到的清單裡（adopted），"+
				"可以拿去計算、回測、組交易策略與機器人。"+
				"\n\n副本是你的，但看不到算式、改不動、不能再上架；之後作者改寫、下架或刪除原本那一支都**不會影響它**。"+
				"你已經有同名的策略腳本（包括再加入同一支）時會被拒絕。不要了就用 trading_delete_strategy_script 刪掉副本。",
			vo.RequestVerbSubmit, "/marketplace/strategy-scripts/{id}/adoption",
			pathParameter("id", "要加入哪一支"),
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
				"label 是你在買賣條件裡怎麼稱呼這一路信號。"+
				"**每個來源的 aggregationInterval 可以各不相同**（例如日線定方向、4h 等回檔、15m 抓進場）："+
				"重演以最細的那一種為步調，較粗的那一路只聽它已經走完的最後一根", true),
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
				"機器人吃的行情必須與交易策略相同，兩種不混用。"+
				ownStrategyScriptsOnlyNote,
			vo.RequestVerbSubmit, "/trading-strategies",
			tradingStrategyWriteParameters()...,
		),
		domains.NewApiToolDomain(
			"trading_list_trading_strategies",
			"列出你的每一份交易策略。",
			vo.RequestVerbRead, "/trading-strategies",
		),
		domains.NewApiToolDomain(
			"trading_get_trading_strategy",
			"讀一份交易策略的完整內容，含信號來源、兩個條件樹、它吃哪一種行情（marketDataKind），"+
				"吃合約行情的另外帶著它的交易模式（tradingMode）。",
			vo.RequestVerbRead, "/trading-strategies/{id}",
			pathParameter("id", "交易策略識別碼"),
		),
		domains.NewApiToolDomain(
			"trading_update_trading_strategy",
			"改一份你自己的交易策略。這是整份改寫：沒帶到的欄位會變成空的。"+
				"\n\n**例外是 marketDataKind 與 tradingMode**：不給就是保留原本的。"+
				"marketDataKind 換成另一種會被拒絕；吃合約行情的那一種可以換交易模式（tradingMode），"+
				"只改名字或條件時不必重帶它。"+
				ownStrategyScriptsOnlyNote,
			vo.RequestVerbReplace, "/trading-strategies/{id}",
			append([]vo.ToolParameterVo{pathParameter("id", "要改哪一份")},
				tradingStrategyWriteParameters()...)...,
		),
		domains.NewApiToolDomain(
			"trading_delete_trading_strategy",
			"刪掉一份你自己的交易策略。掛在它上面的策略機器人會受影響，刪之前先確認。",
			vo.RequestVerbRemove, "/trading-strategies/{id}",
			pathParameter("id", "要刪哪一份"),
		),
	}
}

// ownStrategyScriptsOnlyNote is said on both trading strategy writes, since a signal source naming someone else's
// script is refused on either.
const ownStrategyScriptsOnlyNote = "\n\n**信號來源只能指名你自己的策略腳本**（含從市集加入的副本）。" +
	"指名別人的——即使在市集上——會被拒絕，要先用 trading_adopt_strategy_script 加入，再指名那份副本。"

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
	"TopTraderPositionLongShare、TopTraderPositionShortShare、TopTraderPositionLongShortRatio（大戶多空持倉比）；" +
	"Footprint（**訂單流足跡**，indicator.Footprint：PriceStep 價位寬度、Levels 一串 indicator.PriceLevel——" +
	"每一個有 Price、BuyVolume（買方主動吃單）、SellVolume（賣方主動吃單）、TradeCount，價位由低到高、只列有成交的——" +
	"Delta 買賣差、PointOfControl 成交最多的價位，平手取低）。" +
	"\n\n**訂單流足跡要先宣告一個 kind 為 orderFlowPriceStep 的參數才有**，沒宣告時整份是空的。" +
	"每一筆成交的價格往下取到寬度的整數倍就是它的價位；只算這一格收盤以前的成交。" +
	"訂單流**不存進資料庫**：過完的日子讀交易所的每日歷史檔，今天與檔案還沒出現的昨天用即時查詢；" +
	"交易服務只在記憶中暫留最近整理過的幾天，所以同一段第二次通常很快，但重啟、或請求落在另一台分身上時會重拿，第一次會比較慢。" +
	orderFlowReachNote +
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
// replayWindowNote says each replayed run reads only its declared look-back, as a bot round does, so a script needing more history must declare it.
const replayWindowNote = "\n\n**重演裡每一格只看得到它自己的回看根數**：每一格執行一次算式，拿到的是到那一格為止、最多宣告的最大回看根數那麼多格，" +
	"與機器人每一輪讀的一樣；沒宣告回看根數的才看得到暖機起點以來的全部。所以算式裡均線、ATR 要用多少根，回看根數就要宣告多大。"

// multiIntervalReplayNote explains a replay over sources on different intervals, which only replays of trading strategies meet.
const multiIntervalReplayNote = "\n\n**信號來源可以各用不同的彙總刻度**：重演以最細的那一種為步調逐格往前走，" +
	"回覆與回測輪次的 interval 就是它。較粗的來源在**自己的刻度上**執行算式，每一格只聽它**已經走完的最後一根**——" +
	"還在走的那一根看不到；一根都還沒走完時算持平。暖機照每個來源自己的刻度換算（日線回看 80 根就先讀 80 天），" +
	"換算後超過一次重演的上限會被拒絕，請減少那個來源的回看根數或縮短期間。"

const shortTermReplayNote = "\n\n**研發短線策略時，fillTiming 用 nextOpen**：收盤成交（close，不給即是）是讓說出信號的那一格在它自己的收盤價成交，而收盤那一刻其實已經過去了——短線賺的常常就是那一點點，所以收盤成交的成績單必然偏樂觀。下一格開盤成交時，最後一格說出的信號不會成交。" +
	"\n\n**樣本外：給 validationStartTime，把這段期間切成調參段（之前）與驗證段（之後）。**回來會多 inSample 與 validation 兩份結果（形狀與整段相同），兩段各自從初始資金、空手開始，算式在驗證段看得到之前的歷史。" +
	"**只拿 inSample 調參數，只拿 validation 判斷這支策略好不好**；看過 validation 之後又回頭改參數，那一段就不再是樣本外，請換一段更晚、還沒看過的期間當驗證段。" +
	"驗證起點必須落在期間之內、兩段都要至少一格，否則會被拒絕。" +
	"\n\n**成績單多五格**：profitFactor（賺的那幾筆淨損益合計 ÷ 虧的那幾筆合計，一筆都沒虧時是 null——不是無限大）、expectancy（每一筆平均淨賺多少）、averageHoldingSeconds（平均持倉秒數）、maximumConsecutiveLossCount（最多連虧幾筆，打平會打斷）、costToGrossProfitRatio（手續費吃掉扣成本前價差的比例，0.25 就是吃掉四分之一；價差本身沒賺時是 null；合約的資金費用不算在內）。結束時還開著的那一注不算進這五格。" +
	"**短線策略要看 expectancy 與 costToGrossProfitRatio**：勝率高而 expectancy 接近零、或成本吃掉大半價差，都是交易次數多、沒有真的優勢。" +
	"\n\n**拿到成績單後，筆數、勝率請讀成績單，不要數交易明細**：交到你手上的成績單已經精簡過（見 trading_get_backtest_run）。"

// backgroundReplayNote is shared by the four replays so none of them tells a different story about what comes back.
const backgroundReplayNote = "\n\n**它不等算完**——要求合法就立刻回一筆回測輪次（id、status），請拿 id 用 trading_get_backtest_run 看進度與成績單；短的幾秒，四年份的 1 小時刻度可能要幾分鐘。**把輪次編號說給使用者**，並說明可以隨時回來看。" +
	"\n\nstatus 是 queued 時代表前面還有在跑的，**會自動開始，不要重送**；被拒絕「你同時最多 N 筆」時，等其中一筆結束或用 trading_cancel_backtest_run 取消一筆再送。" +
	"\n\n回測有背景時間上限（預設三十分鐘）：超過的輪次會是 failed、原因寫「超過時間上限」，請縮短期間或改用粗一點的刻度再送一筆。要求本身不合法（區間顛倒、腳本不是你的、旋鈕沒宣告過……）時照舊當場被拒絕、不留輪次。"

func backtestRunApiTools(replayWaitLimit time.Duration) []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_get_backtest_run",
			"看一筆回測輪次走到哪：status 五選一——queued（排隊中，前面還有在跑的，會自動開始）、running（進行中，progressPercent 說走到幾成）、"+
				"succeeded（完成，result 是成績單）、failed（failureReason 說為什麼）、cancelled（已取消）。回覆也帶要跑什麼：kind、symbol、interval、startTime、endTime、subjectName。"+
				"\n\n**可以順便等一下**：給 waitSeconds（最多 60），輪次在那之前結束就立刻回，沒結束就回當時的狀態與進度——不必每隔幾秒戳一次。"+
				"\n\n**failureReason 照原話說**：「超過時間上限」是回測太長、「被重新啟動打斷」是交易服務換了一台，兩者都不是腳本壞了，不要猜成腳本有錯；腳本真的出錯時寫的是那一句錯誤。"+
				"\n\n**成績單（result）交到你手上之前會先精簡**：資金曲線超過兩百點時平均取兩百點（頭尾都在），並多一個 equityCurvePointTotalCount 說原本幾點；"+
				"交易明細超過一百筆時只列最近一百筆，並多一個 closedTradeTotalCount 說總共幾筆；inSample 與 validation 各自照做。**成績單的每一個數字都沒有動**——筆數、勝率請讀成績單，不要數交易明細。"+
				"\n\n不是自己的輪次回 404。",
			vo.RequestVerbRead, "/backtest-runs/{id}",
			pathParameter("id", "四件重演能力回的那個輪次 id"),
			queryParameter("waitSeconds", vo.ToolParameterKindInteger,
				"最多等幾秒再回（0 到 60）。輪次在那之前結束就提前回；不給即不等", false),
		).Waiting(replayWaitLimit).CondensingReplayResults(),
		domains.NewApiToolDomain(
			"trading_list_backtest_runs",
			"列出使用者自己的回測輪次，新到舊，回覆帶 totalCount。**回覆時說出總數。**"+
				"每一筆帶 status、要跑什麼（kind、symbol、interval、startTime、endTime、subjectName）、submittedAt，以及完成後才有的 totalReturnRate、maximumDrawdown、tradeCount；失敗的帶 failureReason。"+
				"\n\n使用者問「昨天那幾次回測」「上次跑的結果」時用這一件找，**不要憑記憶重跑**。要看完整成績單再拿 id 用 trading_get_backtest_run。"+
				"\n\n可以只看某一種 status、只看某一支策略腳本（strategyScriptId）或某一份交易策略（tradingStrategyId）的；limit 不給即預設 20、最多 200。",
			vo.RequestVerbRead, "/backtest-runs",
			queryParameter("status", vo.ToolParameterKindString,
				"只看這一種狀態：queued／running／succeeded／failed／cancelled。不給即全部", false),
			queryParameter("strategyScriptId", vo.ToolParameterKindInteger, "只看重演這一支策略腳本的輪次", false),
			queryParameter("tradingStrategyId", vo.ToolParameterKindInteger, "只看重演這一份交易策略的輪次", false),
			queryParameter("limit", vo.ToolParameterKindInteger, "最多回幾筆，預設 20、最多 200", false),
		),
		domains.NewApiToolDomain(
			"trading_cancel_backtest_run",
			"取消一筆還沒跑完的回測輪次：排隊中的直接變成 cancelled、永遠不會開始；進行中的盡快停下、也變成 cancelled，沒有成績單。"+
				"\n\n已經結束（succeeded／failed／cancelled）的不能取消，會被拒絕並說它已經結束；不是自己的回 404。",
			vo.RequestVerbSubmit, "/backtest-runs/{id}/cancel",
			pathParameter("id", "要取消哪一筆回測輪次"),
		),
	}
}

func backtestApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_backtest_strategy_script",
			"拿**一支策略腳本**重演一段已經發生過的行情。"+
				"\n\n**要嘛指名一支既有的策略腳本（strategyScriptId），要嘛自己帶一段算式（script），兩者只能挑一個。**"+
				"\n\n重演一律以 signal 這種指標值種類執行——它讀的就是每根一個買賣信號，"+
				"所以這裡沒有 resultType 可填、也不需要填。"+
				"\n\n輪次完成後（trading_get_backtest_run）拿到的是成績單、交易明細與資金曲線。"+
				"**成績單裡的交易筆數一定要看**——一張幾乎沒有交易的漂亮成績單會被讀成「很穩」，"+
				"而真相是這份策略根本沒有在做決定。"+
				"\n\n模擬了出場價位時，**stopLossExitCount 也一定要看**："+
				"十次出場八次是被停損掃出去的策略，與十次都靡訊號出場的，"+
				"報酬率可以一模一樣——而前者是停損在支撑它，"+
				"後者是還沒遇到那個掃光它的盤。每一筆交易自己也帶著 exitReason。"+
				costedReportCardNote+
				spotOnlyReplayNote+
				shortTermReplayNote+
				replayWindowNote+
				backgroundReplayNote,
			vo.RequestVerbSubmit, "/backtests",
			append(append([]vo.ToolParameterVo{
				bodyParameter("strategyScriptId", vo.ToolParameterKindInteger,
					"要重演哪一支既有的策略腳本。與 script 擇一", false),
			}, backtestParameters()...),
				bodyParameter("aggregationInterval", vo.ToolParameterKindString,
					"彙總刻度，六選一：1m／5m／15m／1h／4h／1d", false),
				bodyParameter("script", vo.ToolParameterKindString,
					"一段還沒存起來的算式。與 strategyScriptId 擇一", false),
				bodyParameter("parameters", vo.ToolParameterKindArray,
					"自帶算式時它宣告的旋鈕。"+parameterKindsNote+parameterValuesCheckedNote, false),
				bodyParameter("parameterValues", vo.ToolParameterKindArray,
					"這一次要把旋鈕調成多少，每個為 {\"name\":…, \"value\":…}。只用於這次重演，不寫回腳本", false),
			)...,
		),
		domains.NewApiToolDomain(
			"trading_backtest_trading_strategy",
			"拿**一份交易策略**重演一段已經發生過的行情。"+
				"\n\n信號來源與買賣兩個條件都取自那份交易策略本身，這裡不必也不能再說一次。"+
				"\n\n這一支與 trading_backtest_strategy_script 的差別："+
				"那一支重演的是單獨一支算式產出的信號，這一支重演的是幾支信號組合出來的決定。"+
				costedReportCardNote+
				spotOnlyReplayNote+
				multiIntervalReplayNote+
				shortTermReplayNote+
				replayWindowNote+
				backgroundReplayNote,
			vo.RequestVerbSubmit, "/trading-strategies/{id}/backtests",
			append([]vo.ToolParameterVo{pathParameter("id", "要重演哪一份交易策略")},
				backtestParameters()...)...,
		),
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

func contractBacktestApiTools() []domains.ApiToolDomain {
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
				shortTermReplayNote+
				replayWindowNote+
				backgroundReplayNote+
				orderFlowReachNote,
			vo.RequestVerbSubmit, "/contract-backtests",
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
					"自帶算式時它宣告的旋鈕。"+parameterKindsNote+parameterValuesCheckedNote, false),
				bodyParameter("parameterValues", vo.ToolParameterKindArray,
					"這一次要把旋鈕調成多少，每個為 {\"name\":…, \"value\":…}。只用於這次重演，不寫回腳本", false),
			)...,
		),
		domains.NewApiToolDomain(
			"trading_backtest_contract_trading_strategy",
			"拿**一份吃合約行情的交易策略**在**合約帳戶**上重演一段已經發生過的永續合約行情。"+
				"\n\n信號來源、買賣兩個條件、**交易模式**都取自那份交易策略本身，這裡不必也不能再說一次——"+
				"這一件**沒有 tradingMode 那一格**，要換交易模式請用 trading_update_trading_strategy 改那份交易策略。"+
				"吃 K 線的交易策略會被拒絕，那一份要用 trading_backtest_trading_strategy。"+
				"\n\n成績單多一個 conflictedCandleCount：買賣條件同時成立、當作持平的格數。"+
				costedReportCardNote+
				contractAccountReplayNote+
				multiIntervalReplayNote+
				shortTermReplayNote+
				replayWindowNote+
				backgroundReplayNote+
				orderFlowReachNote,
			vo.RequestVerbSubmit, "/trading-strategies/{id}/contract-backtests",
			append([]vo.ToolParameterVo{pathParameter("id", "要重演哪一份交易策略")},
				contractBacktestParameters()...)...,
		),
	}
}

// orderFlowReachNote is how much order flow one run may ask for and what to change when it
// asks for more. It travels with every ability where the limit bites — the contract script
// note, both contract replays and creating a bot — because a limit only said where scripts
// are written is a limit the assistant meets for the first time as a refusal.
const orderFlowReachNote = "\n\n**宣告了 orderFlowPriceStep 的算式，一次至多拿 31 個世界標準時間日曆日的訂單流**：" +
	"從它讀的第一格（**回看根數與重演的暖機都算在內**）的開頭，數到最後一格收盤；跨過午夜就多算一天。" +
	"所以四小時刻度回看 200 根就已經超過，跟要看的區間多短無關。" +
	"\n\n**重演的暖機**：重演一支策略腳本時，它宣告了回看根數就以最大的那個當暖機；**一個都沒宣告時先讀預設的 200 根**——" +
	"四小時或一天刻度因此一定超過，宣告一個小的 lookbackCount（例如算式真正需要的根數）就會取代那 200 根。" +
	"重演一份交易策略時，暖機是**每個信號來源的回看根數換成重演步調（最細的那個刻度）的根數後取最大**——" +
	"日線回看 80 根、步調是 15m 時就是 7,680 根；刻度都相同時就是最大的回看根數。只有每一支都沒宣告時才是預設的 200 根——" +
	"所以縮短訂單流那一支的回看不夠，換算後回看最長的那一支也要一起看。來源刻度不同時，讀取起點還會往前對齊到最粗那一種的格子起點，" +
	"訂單流的天數也從那裡算。重演在**送出時**就照整段判，超過即拒絕。" +
	"\n\n**機器人沒有暖機，而且每個信號來源各算各的**：合約機器人每一輪讓每個來源各讀自己宣告的最大回看根數那麼多格（沒宣告就只讀一格），" +
	"所以用了訂單流的那一支要的是「它自己的回看根數 × 彙總刻度」那麼長；被停下時要改的是那一支，不是別的來源。" +
	"另外一格至多 1,000 個價位、一次至多 2,000,000 個價位。" +
	"\n\n**超過時的出路**：縮短區間、減少回看根數（重演時沒宣告的就宣告一個小的）、換細一點的彙總刻度；價位太多就放寬 orderFlowPriceStep。" +
	"機器人一輪來不及備齊訂單流會跳過那一輪、下一輪接著補；**要的超過上述任何一個上限，機器人會被停下**並說出原因。"
