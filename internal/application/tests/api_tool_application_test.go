package application_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func (connector *connector) call(toolName string, accessToken string) dto.ToolResultDto {
	return connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName: toolName, Arguments: noArguments(), AccessToken: accessToken})
}

func TestEveryAbilityTravelsUnderTheConnectorAuthorizationTheCallerBrought(t *testing.T) {
	testCases := []struct {
		name     string
		apiTool  domains.ApiToolDomain
		toolName string
	}{
		{"個人資源的事", listStrategyScripts(), "trading_list_strategy_scripts"},
		{"看公開資料的事", checkHealth(), "trading_health"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			connector := newConnector(t, testCase.apiTool)
			connector.tradingService.EXPECT().
				Send(gomock.Any(), gomock.Any(), jamesAccessToken).
				Return(succeededWith(`[{"id":1}]`), nil)

			resultDto := connector.call(testCase.toolName, jamesAccessToken)

			assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
			assert.Equal(t, `[{"id":1}]`, resultDto.Content)
		})
	}
}

func TestAConnectorAuthorizationTheTradingServiceDoesNotRecognizeAsksToReconnectWithoutRetrying(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), jamesAccessToken).
		Return(notRecognized(), nil).
		Times(1)

	resultDto := connector.call("trading_list_strategy_scripts", jamesAccessToken)

	assert.Equal(t, dto.ToolOutcomeReconnectRequired, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, "/mcp")
	assert.Contains(t, resultDto.Content, "重新連線")
}

func TestARefusalComesBackInTheTradingServicesOwnWords(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(refusedWith("帳號尚未開通，請寄信至 admin@example.com"), nil)

	resultDto := connector.call("trading_list_strategy_scripts", jamesAccessToken)

	assert.Equal(t, dto.ToolOutcomeRefusedByTradingService, resultDto.Outcome)
	assert.Equal(t, "帳號尚未開通，請寄信至 admin@example.com", resultDto.Content)
}

func TestNotReachingTheTradingServiceIsNotToldAsTheCallersMistake(t *testing.T) {
	connector := newConnector(t, checkHealth())
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(vo.TradingServiceResponseVo{}, errors.New("dial tcp: connection refused"))

	resultDto := connector.call("trading_health", jamesAccessToken)

	assert.Equal(t, dto.ToolOutcomeTradingServiceUnreachable, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, domains.ErrTradingServiceUnreachable.Error())
	assert.NotContains(t, resultDto.Content, "重新連線")
}

func TestAnAbilityThisConnectorDoesNotHaveIsSaidSoRatherThanAttempted(t *testing.T) {
	connector := newConnector(t, checkHealth())

	resultDto := connector.call("trading_place_an_order", jamesAccessToken)

	assert.Equal(t, dto.ToolOutcomeUnknownTool, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, "trading_place_an_order")
}

func TestAHalfFilledFormNamesTheMissingBoxAndIsNotSent(t *testing.T) {
	connector := newConnector(t, domains.NewApiToolDomain(
		"trading_sync_k_candle_history", "同步歷史", vo.RequestVerbSubmit, "/k-candles/history",
		vo.NewToolParameterVo("symbol", vo.ToolParameterKindString, "標的", true, vo.ToolParameterInBody),
		vo.NewToolParameterVo("lookbackDays", vo.ToolParameterKindInteger, "天數", true, vo.ToolParameterInBody),
	))

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName:    "trading_sync_k_candle_history",
		Arguments:   map[string]json.RawMessage{"symbol": json.RawMessage(`"BTCUSDT"`)},
		AccessToken: jamesAccessToken,
	})

	assert.Equal(t, dto.ToolOutcomeInvalidArguments, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, "lookbackDays")
}

func TestListingAbilitiesDescribesEveryOneOfThem(t *testing.T) {
	connector := newConnector(t, listStrategyScripts(), checkHealth())

	definitionDtos := connector.apiTools.ListApiTools()

	require.Len(t, definitionDtos, 2)

	describedNames := map[string]bool{}
	for _, definitionDto := range definitionDtos {
		assert.NotEmpty(t, definitionDto.Description, definitionDto.Name)
		describedNames[definitionDto.Name] = true
	}

	assert.Equal(t,
		map[string]bool{"trading_list_strategy_scripts": true, "trading_health": true},
		describedNames)
}
