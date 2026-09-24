package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Peeking at a contract asks the contract line's live address for that contract, needs
// nobody signed in, and stays on the line for the first update exactly as long as the
// spot peek does.
func TestPeekingAtAContractAsksItsOwnLiveLineForTheFirstUpdate(t *testing.T) {
	request, buildError := apiToolNamed(t, "trading_peek_live_contract_k_candle").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{"symbol": json.RawMessage(`"BTCUSDT"`)}))

	require.NoError(t, buildError)
	assert.Equal(t, vo.RequestVerbRead, request.Verb)
	assert.Equal(t, "/contract-k-candles/live", request.Path)
	assert.Equal(t, "BTCUSDT", request.Query["symbol"])
	assert.False(t, request.CarriesIdentity, "看一眼即時更新不需要登入，與現貨相同")
	assert.Equal(t, 10*time.Second, request.LiveUpdateWaitLimit)

	spotRequest, spotBuildError := apiToolNamed(t, "trading_peek_live_k_candle").BuildRequest(
		domains.NewToolArgumentsDomain(map[string]json.RawMessage{"symbol": json.RawMessage(`"BTCUSDT"`)}))
	require.NoError(t, spotBuildError)
	assert.Equal(t, "/k-candles/live", spotRequest.Path, "現貨的看一眼照舊走現貨那一條")
}

// A contract peek names the contract it looks at; without one there is nothing to ask.
func TestPeekingAtAContractNeedsTheContractNamed(t *testing.T) {
	symbol, isDeclared := boxNamed(abilityNamed(t, "trading_peek_live_contract_k_candle"), "symbol")

	require.True(t, isDeclared)
	assert.True(t, symbol.IsRequired)
}

// What the assistant reads before it peeks: which market this is, what the figures
// are, what each status asks of it, and what gets it refused.
func TestPeekingAtAContractSaysWhatItShowsAndWhatGetsItRefused(t *testing.T) {
	description := abilityNamed(t, "trading_peek_live_contract_k_candle").Description

	for _, phrase := range []string{
		// Its own market, apart from spot.
		"**永續合約**", "與 trading_peek_live_k_candle 是兩件事", "這一件只看合約那一邊",
		// Last price, not mark.
		"**最新價**", "**不含標記價格**", "trading_list_contract_k_candles",
		// The three statuses.
		"forming（這一根還在走，數字還會變，系統不存它",
		"closed（這一根走完了，但**不是由即時更新存下**——每分鐘那一輪會在一分鐘內把完整的那一根存進來）",
		"stalled（即時更新斷了，系統自己在重連，**等一下就好**）",
		"**不會**出現 unavailable 或 marketClosed",
		// Watched contracts only.
		"**只看得到合約追蹤名單上的合約標的**", "trading_add_to_contract_watchlist", "系統不認得的代號回找不到",
		// A quiet wait is an answer.
		"等滿沒有收到東西是正常結果，不是錯誤",
	} {
		assert.Contains(t, description, phrase)
	}
}

// The spot peek is word for word what it was: its five statuses and its watchlist rule.
func TestPeekingAtSpotIsUnchangedByTheContractTwin(t *testing.T) {
	description := abilityNamed(t, "trading_peek_live_k_candle").Description

	assert.Contains(t, description, "unavailable（這一檔分不到這個市場的即時名額，**不會自己好**，要改觀察清單）")
	assert.Contains(t, description, "marketClosed（這個市場現在休市，**什麼都別做**）")
	assert.NotContains(t, description, "合約")
}
