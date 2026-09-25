package main

import (
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangingMarketDataTravelsUnderTheSignedInIdentityWhileReadingItDoesNot(t *testing.T) {
	testCases := []struct {
		toolName        string
		carriesIdentity bool
	}{
		{toolName: "trading_create_k_candle", carriesIdentity: true},
		{toolName: "trading_update_k_candle", carriesIdentity: true},
		{toolName: "trading_delete_k_candle", carriesIdentity: true},
		{toolName: "trading_backfill_k_candles", carriesIdentity: true},
		{toolName: "trading_sync_k_candle_history", carriesIdentity: true},
		{toolName: "trading_get_k_candle_history_sync", carriesIdentity: true},
		{toolName: "trading_add_to_watchlist", carriesIdentity: true},
		{toolName: "trading_remove_from_watchlist", carriesIdentity: true},
		{toolName: "trading_create_contract_k_candle", carriesIdentity: true},
		{toolName: "trading_update_contract_k_candle", carriesIdentity: true},
		{toolName: "trading_delete_contract_k_candle", carriesIdentity: true},
		{toolName: "trading_backfill_contract_k_candles", carriesIdentity: true},
		{toolName: "trading_sync_contract_k_candle_history", carriesIdentity: true},
		{toolName: "trading_get_contract_k_candle_history_sync", carriesIdentity: true},
		{toolName: "trading_add_to_contract_watchlist", carriesIdentity: true},
		{toolName: "trading_remove_from_contract_watchlist", carriesIdentity: true},
		{toolName: "trading_list_k_candles", carriesIdentity: false},
		{toolName: "trading_get_k_candle_series", carriesIdentity: false},
		{toolName: "trading_get_k_candle", carriesIdentity: false},
		{toolName: "trading_peek_live_k_candle", carriesIdentity: false},
		{toolName: "trading_list_trading_symbols", carriesIdentity: false},
		{toolName: "trading_list_contract_k_candles", carriesIdentity: false},
		{toolName: "trading_get_contract_k_candle_series", carriesIdentity: false},
		{toolName: "trading_get_contract_k_candle", carriesIdentity: false},
		{toolName: "trading_peek_live_contract_k_candle", carriesIdentity: false},
		{toolName: "trading_list_contract_trading_symbols", carriesIdentity: false},
		{toolName: "trading_list_contract_funding_rate_settlements", carriesIdentity: false},
		{toolName: "trading_list_contract_position_statistics", carriesIdentity: false},
		{toolName: "trading_get_contract_maintenance_margin_tiers", carriesIdentity: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.toolName, func(t *testing.T) {
			request, buildError := apiToolNamed(t, testCase.toolName).BuildRequest(
				domains.NewToolArgumentsDomain(everyBoxFilledIn()))

			require.NoError(t, buildError)
			assert.Equal(t, testCase.carriesIdentity, request.CarriesIdentity)
		})
	}
}
