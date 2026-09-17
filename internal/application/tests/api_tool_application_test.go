package application_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func (connector *connector) signInOn(t *testing.T, sessionKey string, grant vo.SessionGrantVo) {
	t.Helper()

	connector.tradingService.EXPECT().SignIn(gomock.Any(), gomock.Any()).Return(grant, nil)
	_, signInError := connector.authentication.SignIn(
		context.Background(), sessionKey, jamesAccount)
	require.NoError(t, signInError)
}

func (connector *connector) call(toolName string, sessionKey string) dto.ToolResultDto {
	return connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName: toolName, Arguments: noArguments(), SessionKey: sessionKey})
}

func TestAnAbilityThatNeedsIdentityTravelsUnderThisConnectionsProof(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("james-token"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "james-token").
		Return(succeededWith(`[{"id":1}]`), nil)

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
	assert.Equal(t, `[{"id":1}]`, resultDto.Content)
}

func TestAnAbilityThatNeedsIdentityIsNotEvenAttemptedWithoutOne(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeSignInRequired, resultDto.Outcome)
	assert.Equal(t, "請先登入", resultDto.Content)
}

func TestAnAbilityThatNeedsNoIdentityTravelsWithoutOne(t *testing.T) {
	connector := newConnector(t, checkHealth())
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "").
		Return(succeededWith(`{"status":"Healthy"}`), nil)

	resultDto := connector.call("trading_health", aConnection)

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
}

func TestAnExpiredSigningInIsRenewedWithoutTheUserNoticing(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, anExpiredGrant("stale"))
	connector.tradingService.EXPECT().
		RenewSession(gomock.Any(), "stale-refresh").
		Return(aLiveGrant("fresh"), nil)
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "fresh").
		Return(succeededWith("[]"), nil)

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
}

func TestASigningInPastSavingAsksForANewOneRatherThanRetrying(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aStaleGrant("gone"))

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeSignInExpired, resultDto.Outcome)
	assert.Equal(t, "登入已失效，請重新登入", resultDto.Content)
}

func TestARefusedRenewalGivesUpTheIdentityRatherThanKeepingAUselessOne(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, anExpiredGrant("stale"))
	connector.tradingService.EXPECT().
		RenewSession(gomock.Any(), "stale-refresh").
		Return(vo.SessionGrantVo{Outcome: vo.TradingServiceRefused, Content: "續用憑證已作廢"}, nil)

	first := connector.call("trading_list_strategy_scripts", aConnection)
	second := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeSignInExpired, first.Outcome)
	assert.Equal(t, dto.ToolOutcomeSignInRequired, second.Outcome,
		"作廢的那一份不該留著被拿去換第二次")
}

func TestBeingUnableToReachTheTradingServiceWhileRenewingIsNotToldAsAnExpiry(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, anExpiredGrant("stale"))
	connector.tradingService.EXPECT().
		RenewSession(gomock.Any(), "stale-refresh").
		Return(vo.SessionGrantVo{}, errors.New("connection refused"))

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeTradingServiceUnreachable, resultDto.Outcome)
	assert.NotContains(t, resultDto.Content, "請重新登入",
		"叫人白打一次密碼，是把「網路不通」誤診成「你過期了」的代價")
}

func TestTwoAsksHittingTheExpiryTogetherSpendTheRenewalOnlyOnce(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, anExpiredGrant("stale"))
	connector.tradingService.EXPECT().
		RenewSession(gomock.Any(), "stale-refresh").
		Return(aLiveGrant("fresh"), nil).
		Times(1)
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "fresh").
		Return(succeededWith("[]"), nil).
		Times(2)

	resultDtos := make([]dto.ToolResultDto, 2)
	var bothDone sync.WaitGroup

	bothDone.Add(2)
	for index := range resultDtos {
		go func() {
			defer bothDone.Done()
			resultDtos[index] = connector.call("trading_list_strategy_scripts", aConnection)
		}()
	}
	bothDone.Wait()

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDtos[0].Outcome)
	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDtos[1].Outcome)
}

func TestOneConnectionNeverActsAsAnother(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("james-token"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "james-token").
		Return(succeededWith("[]"), nil)

	signedIn := connector.call("trading_list_strategy_scripts", aConnection)
	stranger := connector.call("trading_list_strategy_scripts", anotherConnection)

	assert.Equal(t, dto.ToolOutcomeSucceeded, signedIn.Outcome)
	assert.Equal(t, dto.ToolOutcomeSignInRequired, stranger.Outcome)
}

func TestEachConnectionTravelsUnderItsOwnProof(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("james-token"))
	connector.signInOn(t, anotherConnection, aLiveGrant("somebody-token"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "james-token").Return(succeededWith("[1]"), nil)
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "somebody-token").Return(succeededWith("[2]"), nil)

	assert.Equal(t, "[1]", connector.call("trading_list_strategy_scripts", aConnection).Content)
	assert.Equal(t, "[2]", connector.call("trading_list_strategy_scripts", anotherConnection).Content)
}

func TestAProofTheCallerBroughtAlongIsUsedAsGiven(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "brought-along").
		Return(succeededWith("[]"), nil)

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName:            "trading_list_strategy_scripts",
		Arguments:           noArguments(),
		SessionKey:          aConnection,
		SuppliedAccessToken: "brought-along",
	})

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
}

func TestAProofTheCallerBroughtAlongWinsOverTheOneBeingHeld(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("held"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "brought-along").
		Return(succeededWith("[]"), nil)

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName:            "trading_list_strategy_scripts",
		Arguments:           noArguments(),
		SessionKey:          aConnection,
		SuppliedAccessToken: "brought-along",
	})

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
}

func TestAProofTheCallerBroughtAlongIsNeverRenewed(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("held"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "brought-along").
		Return(notRecognized(), nil).
		Times(1)

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName:            "trading_list_strategy_scripts",
		Arguments:           noArguments(),
		SessionKey:          aConnection,
		SuppliedAccessToken: "brought-along",
	})

	assert.Equal(t, dto.ToolOutcomeSignInExpired, resultDto.Outcome,
		"續用的那一半不在手上，再送一次只會送出同一份被拒絕的憑證")
}

func TestAProofTheTradingServiceRejectsIsRenewedAndTheAskMadeOnceMore(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("revoked-elsewhere"))
	gomock.InOrder(
		connector.tradingService.EXPECT().
			Send(gomock.Any(), gomock.Any(), "revoked-elsewhere").Return(notRecognized(), nil),
		connector.tradingService.EXPECT().
			RenewSession(gomock.Any(), "revoked-elsewhere-refresh").Return(aLiveGrant("fresh"), nil),
		connector.tradingService.EXPECT().
			Send(gomock.Any(), gomock.Any(), "fresh").Return(succeededWith("[]"), nil),
	)

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
}

func TestARejectionThatSurvivesARenewalIsNotAskedAThirdTime(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("no-good"))
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(notRecognized(), nil).Times(2)
	connector.tradingService.EXPECT().
		RenewSession(gomock.Any(), "no-good-refresh").Return(aLiveGrant("fresh"), nil).Times(1)

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeSignInExpired, resultDto.Outcome)
}

func TestARefusalComesBackInTheTradingServicesOwnWords(t *testing.T) {
	connector := newConnector(t, checkHealth())
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(refusedWith("回溯天數必須在 1 到 3650 之間"), nil)

	resultDto := connector.call("trading_health", aConnection)

	assert.Equal(t, dto.ToolOutcomeRefusedByTradingService, resultDto.Outcome)
	assert.Equal(t, "回溯天數必須在 1 到 3650 之間", resultDto.Content,
		"一字不改——助理要靠這句話知道該把天數改成多少")
}

func TestNotReachingTheTradingServiceIsNotToldAsTheCallersMistake(t *testing.T) {
	connector := newConnector(t, checkHealth())
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(vo.TradingServiceResponseVo{}, errors.New("dial tcp: connection refused"))

	resultDto := connector.call("trading_health", aConnection)

	assert.Equal(t, dto.ToolOutcomeTradingServiceUnreachable, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, domains.ErrTradingServiceUnreachable.Error())
}

func TestAnAbilityThisConnectorDoesNotHaveIsSaidSoRatherThanAttempted(t *testing.T) {
	connector := newConnector(t, checkHealth())

	resultDto := connector.call("trading_place_an_order", aConnection)

	assert.Equal(t, dto.ToolOutcomeUnknownTool, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, "trading_place_an_order")
}

func TestAHalfFilledFormNamesTheMissingBoxAndIsNotSent(t *testing.T) {
	connector := newConnector(t, domains.NewApiToolDomain(
		"trading_sync_k_candle_history", "同步歷史", vo.RequestVerbSubmit, "/k-candles/history", false,
		vo.NewToolParameterVo("symbol", vo.ToolParameterKindString, "標的", true, vo.ToolParameterInBody),
		vo.NewToolParameterVo("lookbackDays", vo.ToolParameterKindInteger, "天數", true, vo.ToolParameterInBody),
	))

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName:   "trading_sync_k_candle_history",
		Arguments:  map[string]json.RawMessage{"symbol": json.RawMessage(`"BTCUSDT"`)},
		SessionKey: aConnection,
	})

	assert.Equal(t, dto.ToolOutcomeInvalidArguments, resultDto.Outcome)
	assert.Contains(t, resultDto.Content, "lookbackDays")
}

func TestListingAbilitiesDescribesEveryOneOfThem(t *testing.T) {
	connector := newConnector(t, listStrategyScripts(), checkHealth())

	definitionDtos := connector.apiTools.ListApiTools()

	require.Len(t, definitionDtos, 2)

	requiresSignInPerNames := map[string]bool{}
	for _, definitionDto := range definitionDtos {
		assert.NotEmpty(t, definitionDto.Description, definitionDto.Name)
		requiresSignInPerNames[definitionDto.Name] = definitionDto.RequiresSignIn
	}

	assert.Equal(t,
		map[string]bool{"trading_list_strategy_scripts": true, "trading_health": false},
		requiresSignInPerNames)
}

func TestLosingTheTradingServiceBetweenTheRenewalAndTheRetryIsSaidAsSuch(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("revoked-elsewhere"))
	gomock.InOrder(
		connector.tradingService.EXPECT().
			Send(gomock.Any(), gomock.Any(), "revoked-elsewhere").Return(notRecognized(), nil),
		connector.tradingService.EXPECT().
			RenewSession(gomock.Any(), "revoked-elsewhere-refresh").Return(aLiveGrant("fresh"), nil),
		connector.tradingService.EXPECT().
			Send(gomock.Any(), gomock.Any(), "fresh").
			Return(vo.TradingServiceResponseVo{}, errors.New("connection reset")),
	)

	resultDto := connector.call("trading_list_strategy_scripts", aConnection)

	assert.Equal(t, dto.ToolOutcomeTradingServiceUnreachable, resultDto.Outcome)
}

func TestTwoRejectedAsksSpendTheRenewalOnceAndShareWhatItBought(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.signInOn(t, aConnection, aLiveGrant("revoked-elsewhere"))

	// 兩件事同時帶著同一份憑證出發，兩件都被交易服務退回來。它們會一起撞上續用那道門。
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "revoked-elsewhere").
		Return(notRecognized(), nil).
		Times(2)

	// 先進門的那一件花掉續用；後進門的那一件看到手上已經換過了，就用換來的那一份。
	// 兩件都去換的話，交易服務會把整條換發鏈作廢，把真正的使用者登出。
	connector.tradingService.EXPECT().
		RenewSession(gomock.Any(), "revoked-elsewhere-refresh").
		Return(aLiveGrant("fresh"), nil).
		Times(1)

	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "fresh").
		Return(succeededWith("[]"), nil).
		Times(2)

	resultDtos := make([]dto.ToolResultDto, 2)
	var bothDone sync.WaitGroup

	bothDone.Add(2)
	for index := range resultDtos {
		go func() {
			defer bothDone.Done()
			resultDtos[index] = connector.call("trading_list_strategy_scripts", aConnection)
		}()
	}
	bothDone.Wait()

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDtos[0].Outcome)
	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDtos[1].Outcome)
}
