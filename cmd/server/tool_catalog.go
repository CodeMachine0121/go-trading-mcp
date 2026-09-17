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
		bodyParameter("tradingMode", vo.ToolParameterKindString,
			"這次重演照哪一套規則交易。省略即一直留在市場裡", false),
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
