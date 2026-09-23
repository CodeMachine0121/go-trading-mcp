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
				"func Calculate(data <依 marketDataKind 而定>) <依 resultType 而定>："+
				"吃 K 線的收 []indicator.KCandle，吃合約行情的收 []indicator.ContractKCandle——"+
				"入口收錯形狀就算不動。"+
				"只能做純運算，可用 math 與 sort；os、net/http、time、亂數一律取用不到", true),
		bodyParameter("resultType", vo.ToolParameterKindString,
			"指標值種類，五選一：float（預設，回 map[string]float64）、floatList（map[string][]float64）、"+
				"bool、boolList、signal（回 indicator.Signal，值只能是 Buy/Sell/Hold，回測一律用這種）", false),
		// Leaving it out means two different things on the two abilities that share this
		// list — the spot K candle on a create, the kind already held on a rewrite — and
		// both are said here, in the one box both of them show. The connector forwards
		// only what was filled in, so a blank never leaves here as a kind at all: the
		// trading service's own rule decides, and filling in kCandle on the caller's
		// behalf would get a contract script's rename refused.
		bodyParameter("marketDataKind", vo.ToolParameterKindString,
			"這支算式吃哪一種行情，二選一：kCandle（現貨 K 線）或 contractKCandle（永續合約的合約行情格）。"+
				"**建立時不給就是 kCandle**；**修改時不給就是保留原本的**，照抄原本的也可以。"+
				"**建立後不得更換**——換成另一種會被拒絕，要吃另一種請另建一支", false),
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
// **Which set of rules to trade by is not here, and not anywhere.** There is one, so
// there is nothing to ask. **Do not add it back when a replay of contracts arrives** —
// that is a second ability with its own list, the way contract candles are their own
// line rather than a flag on spot ones. A box here would put a choice back on the two
// replays that cannot honour it.
//
// The same goes for borrowing: this list had a multiplier and a maintenance margin
// until this service stopped lending.
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
			"開倉要付的手續費，佔**押下去的金額**的百分之幾"+
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
	catalog = append(catalog, contractApiTools()...)
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
