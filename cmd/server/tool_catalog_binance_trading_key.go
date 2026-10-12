package main

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

const binanceAutoOrderWebOnlyNote = "\n\n**存入、更換、移除幣安交易金鑰，以及打開或關掉自動下單，這個外掛都做不到**——" +
	"只能由使用者自己在網頁上做：金鑰在設定頁，自動下單在那台機器人的詳細頁。" +
	"使用者要求時，說明這件事要他自己到網頁上做，不要送出任何改動。" +
	"**絕不要請使用者把 API Key 或 Secret Key 貼進對話**；他主動貼了，也不要複述、不要轉送到任何能力裡，請他改到網頁設定頁自己填"

// autoOrderReallyTradesNote is the one place that says what an auto order does, because the instructions and every bot ability once all repeated a sentence that went stale on the same day.
const autoOrderReallyTradesNote = "提到自動下單時要說清楚：**合約機器人開著自動下單時，會真的在使用者的幣安帳戶上開倉與平倉**——" +
	"開倉掛一張有上限的限價進場單（價格不超過參考價 ± 那台機器人的進場滑價容忍度）、平倉立刻以市場價成交，並掛上止損止盈；" +
	"現貨機器人開著目前仍不下單。"

// Deliberately only the status view: every other trading key route and the auto-order switch accept web sign-in only.
func binanceTradingKeyApiTools() []domains.ApiToolDomain {
	return []domains.ApiToolDomain{
		domains.NewApiToolDomain(
			"trading_get_binance_trading_key_status",
			"查看你有沒有設定**幣安交易金鑰**，以及它的**可交易市場**。"+
				"回覆三樣：configured（有沒有設定）、tradableMarkets（spot 現貨、contract 合約，或兩者）、configuredAt（設定時刻，世界標準時間）。"+
				"\n\n**回覆不含任何一段金鑰內容**，連 API Key 的結尾也沒有——要認出是哪一組，請使用者到網頁設定頁看。"+
				"\n\n**configured 為 false 時只說「尚未設定」**，並指引使用者到網頁的設定頁自己設定；"+
				"這時 tradableMarkets 是空的，configuredAt 是一個沒有意義的零時刻，不要把它說成日期。"+
				"\n\n可交易市場是存入當下幣安說的；使用者之後在幣安那邊改了權限，要在網頁上重存一次才會更新。"+
				binanceAutoOrderWebOnlyNote,
			vo.RequestVerbRead, "/users/me/binance-trading-key/status",
		),
	}
}
