package main

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// pathParameter is a value that names *which* thing, written into the address itself.
// Naming a thing is never optional, so these are always required and always text.
func pathParameter(name string, description string) vo.ToolParameterVo {
	return vo.NewToolParameterVo(
		name, vo.ToolParameterKindString, description, true, vo.ToolParameterInPath)
}

// queryParameter is a value that narrows what comes back.
func queryParameter(
	name string,
	kind vo.ToolParameterKind,
	description string,
	isRequired bool,
) vo.ToolParameterVo {
	return vo.NewToolParameterVo(name, kind, description, isRequired, vo.ToolParameterInQuery)
}

// bodyParameter is a value that is part of what is being submitted.
func bodyParameter(
	name string,
	kind vo.ToolParameterKind,
	description string,
	isRequired bool,
) vo.ToolParameterVo {
	return vo.NewToolParameterVo(name, kind, description, isRequired, vo.ToolParameterInBody)
}

// kCandlePriceParameters are the price and volume figures of one K candle.
//
// They are shared by creating one and rewriting one, because they are the same eight
// figures and the trading service applies the same rules to both. Written twice, one
// copy would eventually be missing a figure that the other has.
func kCandlePriceParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("open", vo.ToolParameterKindString, "開盤價（字串形式的精確小數，如 \"100.5\"）", true),
		bodyParameter("high", vo.ToolParameterKindString, "最高價。不得低於最低價", true),
		bodyParameter("low", vo.ToolParameterKindString, "最低價", true),
		bodyParameter("close", vo.ToolParameterKindString, "收盤價", true),
		bodyParameter("volume", vo.ToolParameterKindString, "成交量（標的數量）", true),
		bodyParameter("quoteVolume", vo.ToolParameterKindString,
			"成交額。台股的行情來源沒有這一項，該市場請省略——省略與填 0 意義不同", false),
		bodyParameter("takerBuyBaseVolume", vo.ToolParameterKindString,
			"主動買入量。台股沒有這一項，請省略", false),
		bodyParameter("takerBuyQuoteVolume", vo.ToolParameterKindString,
			"主動買入額。台股沒有這一項，請省略", false),
	}
}

// strategyScriptWriteParameters are what a strategy script is made of. Shared by
// creating one and rewriting one, for the same reason as the figures above.
func strategyScriptWriteParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("name", vo.ToolParameterKindString, "這支策略腳本叫什麼", true),
		bodyParameter("description", vo.ToolParameterKindString,
			"它是做什麼的。上架到市集時，這是讀者唯一看得到的東西", false),
		bodyParameter("script", vo.ToolParameterKindString,
			"一段 Go 算式。形狀固定為 package main / import \"indicator\" / "+
				"func Calculate(data []indicator.KCandle) <依 resultType 而定>。"+
				"只能做純運算，可用 math 與 sort；os、net/http、time、亂數一律取用不到", true),
		bodyParameter("resultType", vo.ToolParameterKindString,
			"指標值種類，五選一：float（預設，回 map[string]float64）、floatList（map[string][]float64）、"+
				"bool、boolList、signal（回 indicator.Signal，值只能是 Buy/Sell/Hold，回測一律用這種）", false),
		bodyParameter("parameters", vo.ToolParameterKindArray,
			"這支算式自己的旋鈕，每個為 {\"name\":…, \"kind\":…, \"defaultValue\":…}。"+
				"kind 只有兩種：lookbackCount（大於零的整數，這條線要回看幾根；系統取所有這種的最大值決定要讀幾根）"+
				"與 number（任何數字，原樣交給算式）。算式以 indicator.LookbackCount(\"名字\") "+
				"與 indicator.Number(\"名字\") 取用", false),
	}
}

// backtestParameters are the account rules a replay trades by. Shared by replaying a
// strategy script and replaying a trading strategy.
//
// The test for belonging here is one question: **does that endpoint actually use it?**
// The two exit distances do — both replays read them, because a trading strategy has
// no opinion about what its owner can sit through — so they are here, and a box added
// once is a box both abilities get.
//
// **The trading mode is deliberately not here.** The two replays no longer want the
// same conditions: replaying a script asks the caller which set of rules to trade by,
// while replaying a trading strategy reads it off the trading strategy itself. So the
// box belongs to the one ability that can answer it, and is added there.
//
// Keeping it shared would leave the trading-strategy replay carrying a box the trading
// service ignores in silence — and an assistant has no way to tell it was ignored. It
// would go on believing it replayed a spot account while reading a report card built
// the other way. Before unifying this again, check whether that endpoint has started
// accepting it.
func backtestParameters() []vo.ToolParameterVo {
	return []vo.ToolParameterVo{
		bodyParameter("symbol", vo.ToolParameterKindString, "要在哪一個交易標的上重演", true),
		bodyParameter("startTime", vo.ToolParameterKindString, "重演的起點（RFC3339 世界標準時間）", true),
		bodyParameter("endTime", vo.ToolParameterKindString, "重演的終點（RFC3339 世界標準時間）", true),
		bodyParameter("initialCapital", vo.ToolParameterKindString, "起始資金（字串形式的精確小數）", true),
		bodyParameter("positionSizingMode", vo.ToolParameterKindString,
			"每次進場押多少的方式。押全部時不必給 positionSizingValue", false),
		bodyParameter("positionSizingValue", vo.ToolParameterKindString,
			"搭配 positionSizingMode 的數字（字串形式的精確小數）", false),
		// The three things an assistant cannot find out before sending: that leaving
		// these out simulates nothing at all, that the distances are measured from
		// the entry fill, and that a candle reaching both levels counts as the stop.
		// The last one is the only one of the three that moves the numbers the
		// *worse* way — a surprise in the good direction gets read as good news,
		// one in the bad direction gets read as a bug somewhere else.
		bodyParameter("stopLossPercentage", vo.ToolParameterKindString,
			"這次重演要不要模擬止損，以及止損價離**進場價**幾個百分點"+
				"（3 就是 3%，字串形式的精確小數）。"+
				"**不給就是完全不模擬止損**——不是套用一個常見的預設值，"+
				"所以不給的那一次成績單講的仍然是「一路抱到訊號叫你出場」。"+
				"負的與超過 100 會被拒絕（正好 100 可以，那讓止損價正好是零）。"+
				"**與機器人 positionPlan 裡那一格同名、不同事**："+
				"那一格是「每一輪要建議什麼」、從最新價量起；"+
				"這一個是「這一次怎麼模擬」、從進場價量起", false),
		bodyParameter("takeProfitPercentage", vo.ToolParameterKindString,
			"這次重演的止盈距離，規則與 stopLossPercentage 一字不差，方向相反。"+
				"兩個可以各自單獨給。"+
				"\n\n**同一根 K 線同時碰到止損與止盈時一律算止損**——"+
				"一根 K 線的高低點說不出哪一個先到，"+
				"而兩種讀法只有這一種永遠不會讓成績單變好看。"+
				"所以兩個都給時，回來的數字會比「先碰到止盈」那種算法差，那是刻意的", false),
		// The two boxes an assistant is most likely to leave out and least able to
		// notice it left out. Everything the exit distances get wrong by guessing
		// shows up as a refusal or a stranger-looking report card; everything these
		// two get wrong by guessing shows up as **a better report card**, and nothing
		// anywhere says so. The bias is also the one that ruins the job the assistant
		// is asked to do most — putting two strategies side by side — so the warning
		// names that job rather than describing the field.
		bodyParameter("entryCostPercentage", vo.ToolParameterKindString,
			"開倉要付的手續費，佔**曝險金額**的百分之幾"+
				"（沒開槓桿時曝險金額就是押注金額；開了 5 倍，同一個費率收的錢就是五倍）"+
				"（0.1 就是 0.1%，字串形式的精確小數）。"+
				"**不給就是完全不計手續費**——不是套用一個常見的費率，"+
				"所以不給的那一次成績單講的是一個交易免費的世界，必然偏樂觀。"+
				"\n\n**偏多少與交易次數成正比**：台股一趟完整進出約 0.47%，"+
				"一年 4 趟只吃掉 1.9%，一年 200 趟吃掉六成本金。"+
				"所以**比較兩支交易頻率差很多的策略時，不填費率等於沒有在比較**——"+
				"交易頻繁的那一支被高估最多，而它的成績單看起來最漂亮，名次可能是反的。"+
				"使用者問「扣掉手續費還賺嗎」時也一樣，填進去再跑一次。"+
				"\n\n負的與超過 100 會被拒絕（正好 100 可以，那讓整筆成交金額都拿去付成本）", false),
		bodyParameter("exitCostPercentage", vo.ToolParameterKindString,
			"平倉要付的手續費與稅，佔**成交金額**的百分之幾，驗證規則與 entryCostPercentage 一字不差。"+
				"\n\n**不給時沿用 entryCostPercentage**，不是不收——"+
				"這與上面那兩個出場距離**不一樣**（那兩格各自獨立、各自留白即不模擬）。"+
				"這兩格是同一件事的兩半，所以幣安那種兩邊一樣的只要填 entryCostPercentage 一格；"+
				"台股買賣不對稱，才需要兩格都填。"+
				"\n\n常見的實際數字：台股手續費 0.1425% 打六折約 **0.0855**，"+
				"賣出再加 0.3% 證交稅，所以出場約 **0.3855**；"+
				"幣安吃單兩邊各 **0.1**。你看不到使用者的券商，這幾個數字要由他確認", false),
		// The one box whose every wrong guess is invisible.
		//
		// Everything else on this list misbehaves in a way that shows: an exit
		// distance guessed wrong produces a stranger report card, a cost rate
		// guessed wrong produces a rosier one. Guessing this one wrong produces a
		// report card for **an account that stopped existing halfway through** —
		// complete, plausible, and counting dozens of trades that never happened.
		//
		// So this description carries six things rather than one, and their order
		// is the design: what happens when it is left out, what happens when it is
		// not, how far the position can fall, the one protection the assistant can
		// actually add, where the damage becomes visible, and last the case that
		// gets refused — because a refusal announces itself and the other five
		// never do.
		bodyParameter("leverage", vo.ToolParameterKindString,
			"這次重演借幾倍的錢：曝險是押下去的錢的幾倍"+
				"（5 就是 5 倍，字串形式的精確小數）。"+
				"\n\n**不給、給 0 或給 1 都是不借錢**——不模擬強制平倉，"+
				"成績單與沒有這一格的時候一字不差。"+
				"\n\n給大於 1 就會模擬三件事：賺賠與手續費都照**放大後的曝險金額**算；"+
				"價格逆著走到撐不住時那一注**被強制平倉、押下去的錢全沒了**；"+
				"而**重演會照樣往下跑**——那個帳戶已經歸零，後面每一筆交易都不曾發生。"+
				"\n\n**撐得住多遠由槓桿決定**：大約 `(100÷槓桿)` 個百分點"+
				"（更準確地說是 `100÷槓桿 − 維持保證金率`）。"+
				"5 倍約 **19.5%**、10 倍約 **9.5%**、20 倍約 **4.5%**——"+
				"替使用者挑倍數之前先算這個數字，再對照那段行情的回檔幅度。"+
				"\n\n**止損比強平近時永遠是止損先出場**（兩者在同一邊，近的先到）。"+
				"所以**開了槓桿就一起給 stopLossPercentage**：那是你唯一擋得住歸零的辦法，"+
				"而不給的話，5 倍配一段跌兩成的行情就是整個帳戶。"+
				"\n\n成績單上的 **liquidationExitCount** 會告訴你這一次歸零過幾次。"+
				"\n\n**介於 0 與 1 之間會整次被拒絕**——0 是「沒填」、1 是「不借錢」，"+
				"但 0.5 兩者都不是。打 0.5 的人多半想押半個部位，那要改的是 positionSizingValue。"+
				"\n\n**借不借得到錢由交易模式決定**：longShort 與 leveragedLong 借得到，"+
				"**現貨（spot）借不到**——現貨是拿現金換東西，沒有人借錢給你，"+
				"所以那個交易模式給大於 1 會整次被拒絕。"+
				"使用者是在合約帳戶上只做多的話，要改的是那份交易策略的交易模式"+
				"（改成 leveragedLong），不是把槓桿拿掉", false),
		bodyParameter("maintenanceMarginRate", vo.ToolParameterKindString,
			"一注帳上剩到多少就被強制出場，佔**曝險金額**的百分之幾"+
				"（0.5 就是 0.5%，字串形式的精確小數）。"+
				"\n\n**不給不是關掉它，是用 0.5%**——這與旁邊每一格的留白都不一樣"+
				"（出場距離留白是不模擬、進場成本率留白是不收費）。"+
				"只要借了錢就一定有人在看著抵押品，這不是一件可以不模擬的事；"+
				"留白只代表你沒有意見，於是用市場上的常見值。"+
				"**沒有借錢時這一格不影響任何結果。**"+
				"\n\n必須小於 `100÷槓桿`（5 倍時小於 20），"+
				"否則那一注在開倉那一棒就已經撐不住，整次重演會被拒絕", false),
	}
}

// apiToolCatalog is everything this connector can do.
//
// **One thing the trading service offers is deliberately not here: its own chat
// assistant.** Relaying it would let one AI spend another AI's budget — an assistant
// calling an assistant, with a token bill attached and nobody between them deciding
// it was worth it. The person can still use that assistant directly; what is removed
// is a model's ability to reach for it unprompted.
//
// **This list is the feature.** Every ability shares one path through the connector,
// so what makes fifty of them fifty different things is only what is written here.
// Adding the fifty-first is adding an entry — no handler, no branch, nothing in any
// other file.
//
// Two rules hold for every entry. First, the description is written for the assistant
// and not for a person browsing docs: it says what the ability does *and what will get
// it refused*, because an assistant that cannot see the second one discovers it by
// being refused. Second, nothing here repeats a rule the trading service enforces —
// the descriptions explain, they do not validate.
func apiToolCatalog(liveUpdateWaitLimit time.Duration) []domains.ApiToolDomain {
	catalog := []domains.ApiToolDomain{}
	catalog = append(catalog, systemApiTools()...)
	catalog = append(catalog, accountApiTools()...)
	catalog = append(catalog, kCandleApiTools(liveUpdateWaitLimit)...)
	catalog = append(catalog, tradingSymbolApiTools()...)
	catalog = append(catalog, indicatorApiTools()...)
	catalog = append(catalog, strategyScriptApiTools()...)
	catalog = append(catalog, tradingStrategyApiTools()...)
	catalog = append(catalog, backtestApiTools()...)
	catalog = append(catalog, strategyBotApiTools()...)
	catalog = append(catalog, telegramDeliveryApiTools()...)

	return catalog
}

func systemApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_health",
			"確認交易服務活著。它只檢查行程還在，不檢查任何業務功能——資料庫壞掉時它照樣回答活著。",
			vo.RequestVerbRead, "/health", false,
		),
	}
}

func accountApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_register_user",
			"用電子郵件與密碼建立一位使用者。不需要先登入（系統一位使用者都沒有時，關起來就沒有人建得出第一位）。"+
				"\n\n密碼有一條上限 72 位元組的規則（中文字一個算三個），超過是拒絕而不是截短。"+
				"回覆永遠不含密碼或由它算出來的任何東西。"+
				"\n\n建立完成之後，請用 trading_sign_in 登入。",
			vo.RequestVerbSubmit, "/users", false,
			bodyParameter("email", vo.ToolParameterKindString, "當帳號用的電子郵件", true),
			bodyParameter("password", vo.ToolParameterKindString, "密碼，上限 72 位元組", true),
		),
		domains.NewApiToolDomain(
			"trading_get_current_user",
			"問交易服務「我是誰」，回覆目前這份身分的使用者識別碼與電子郵件。"+
				"用來確認外掛保管的身分是不是你以為的那一個。",
			vo.RequestVerbRead, "/users/me", true,
		),
		domains.NewApiToolDomain(
			"trading_change_password",
			"更換密碼。舊密碼對不上即拒絕；新密碼一樣受 72 位元組的上限規則。"+
				"\n\n改完之後既有的登入仍然有效——這一支不會把你登出。",
			vo.RequestVerbSubmit, "/users/me/password", true,
			bodyParameter("currentPassword", vo.ToolParameterKindString, "目前的密碼", true),
			bodyParameter("newPassword", vo.ToolParameterKindString, "要改成的新密碼", true),
		),
	}
}
