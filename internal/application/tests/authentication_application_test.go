package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var jamesAccount = dto.SignInDto{Email: "james@example.com", Password: "correct horse"}

func TestSigningInTellsYouWhoYouAreAndNeverHandsBackAProof(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().
		SignIn(gomock.Any(), jamesAccount).
		Return(aLiveGrant("james-token"), nil)

	outcomeDto, signInError := connector.authentication.SignIn(
		context.Background(), aConnection, jamesAccount)

	require.NoError(t, signInError)
	assert.Equal(t, dto.ToolOutcomeSucceeded, outcomeDto.Outcome)
	assert.Equal(t, "james@example.com", outcomeDto.Session.Email)
	assert.Equal(t, now.Add(15*60*1_000_000_000), outcomeDto.Session.AccessTokenExpiresAt)
	assert.NotContains(t, outcomeDto.Content, "james-token")
}

func TestSigningInWithTheWrongPasswordRepeatsTheTradingServicesOwnWords(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().
		SignIn(gomock.Any(), gomock.Any()).
		Return(vo.SessionGrantVo{
			Outcome: vo.TradingServiceRefused, Content: "電子郵件或密碼不正確"}, nil)

	outcomeDto, signInError := connector.authentication.SignIn(
		context.Background(), aConnection, jamesAccount)

	require.NoError(t, signInError)
	assert.Equal(t, dto.ToolOutcomeRefusedByTradingService, outcomeDto.Outcome)
	assert.Equal(t, "電子郵件或密碼不正確", outcomeDto.Content)
}

func TestSigningInWhenTheTradingServiceCannotBeReachedSaysSo(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().
		SignIn(gomock.Any(), gomock.Any()).
		Return(vo.SessionGrantVo{}, errors.New("connection refused"))

	_, signInError := connector.authentication.SignIn(
		context.Background(), aConnection, jamesAccount)

	require.Error(t, signInError)
}

func TestARefusedSigningInLeavesTheConnectionWithNoIdentityAtAll(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().
		SignIn(gomock.Any(), gomock.Any()).
		Return(vo.SessionGrantVo{Outcome: vo.TradingServiceRefused, Content: "電子郵件或密碼不正確"}, nil)

	_, _ = connector.authentication.SignIn(context.Background(), aConnection, jamesAccount)

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName: "trading_list_strategy_scripts", Arguments: noArguments(), SessionKey: aConnection})

	assert.Equal(t, dto.ToolOutcomeSignInRequired, resultDto.Outcome)
}

func TestSigningInAgainOnTheSameConnectionBecomesTheNewPerson(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	gomock.InOrder(
		connector.tradingService.EXPECT().SignIn(gomock.Any(), gomock.Any()).
			Return(aLiveGrant("first"), nil),
		connector.tradingService.EXPECT().SignIn(gomock.Any(), gomock.Any()).
			Return(aLiveGrant("second"), nil),
	)
	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "second").
		Return(succeededWith("[]"), nil)

	_, _ = connector.authentication.SignIn(context.Background(), aConnection, jamesAccount)
	_, _ = connector.authentication.SignIn(context.Background(), aConnection,
		dto.SignInDto{Email: "somebody@example.com", Password: "another"})

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName: "trading_list_strategy_scripts", Arguments: noArguments(), SessionKey: aConnection})

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
}

func TestSigningOutVoidsTheChainAndLeavesTheConnectionAsNobody(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().SignIn(gomock.Any(), gomock.Any()).
		Return(aLiveGrant("james-token"), nil)
	connector.tradingService.EXPECT().
		RevokeSession(gomock.Any(), "james-token-refresh").
		Return(succeededWith(""), nil)

	_, _ = connector.authentication.SignIn(context.Background(), aConnection, jamesAccount)
	signOutError := connector.authentication.SignOut(context.Background(), aConnection)

	require.NoError(t, signOutError)

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName: "trading_list_strategy_scripts", Arguments: noArguments(), SessionKey: aConnection})

	assert.Equal(t, dto.ToolOutcomeSignInRequired, resultDto.Outcome)
}

func TestSigningOutWhenNobodyWasSignedInIsStillSuccess(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())

	signOutError := connector.authentication.SignOut(context.Background(), aConnection)

	require.NoError(t, signOutError,
		"目的已經達成了——這個連線本來就不代表任何人")
}

func TestRenewingOnDemandSwapsInTheFreshProof(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())
	connector.tradingService.EXPECT().SignIn(gomock.Any(), gomock.Any()).
		Return(anExpiredGrant("stale"), nil)
	connector.tradingService.EXPECT().RenewSession(gomock.Any(), "stale-refresh").
		Return(aLiveGrant("fresh"), nil)
	connector.tradingService.EXPECT().Send(gomock.Any(), gomock.Any(), "fresh").
		Return(succeededWith("[]"), nil)

	_, _ = connector.authentication.SignIn(context.Background(), aConnection, jamesAccount)
	renewalError := connector.authentication.RenewSession(context.Background(), aConnection)

	require.NoError(t, renewalError)

	resultDto := connector.apiTools.CallApiTool(context.Background(), dto.ToolCallDto{
		ToolName: "trading_list_strategy_scripts", Arguments: noArguments(), SessionKey: aConnection})

	assert.Equal(t, dto.ToolOutcomeSucceeded, resultDto.Outcome)
}

func TestRenewingWithoutHavingSignedInAsksYouToSignIn(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())

	renewalError := connector.authentication.RenewSession(context.Background(), aConnection)

	assert.ErrorIs(t, renewalError, domainsErrSignInRequired())
}

// TestSigningInLetsGoOfIdentitiesNobodyCanUseAnyMore guards the one place this
// connector's memory could grow without bound.
//
// A connection can disappear without ever signing out — a laptop shuts, a client
// crashes. Nothing then arrives to clear its drawer, so without this every
// connection that ever signed in leaves a live pair of proofs behind for as long as
// the process runs.
func TestSigningInLetsGoOfIdentitiesNobodyCanUseAnyMore(t *testing.T) {
	connector := newConnector(t, listStrategyScripts())

	// 甲登入之後就人間蒸發了，而且它那一份連續用都過期了——再也做不了任何事。
	connector.signInOn(t, aConnection, aStaleGrant("abandoned"))

	// 乙登入。這是這個外掛唯一會長大的時刻，所以也是該放手的時刻。
	connector.signInOn(t, anotherConnection, aLiveGrant("current"))

	connector.tradingService.EXPECT().
		Send(gomock.Any(), gomock.Any(), "current").
		Return(succeededWith("[]"), nil)

	abandoned := connector.call("trading_list_strategy_scripts", aConnection)
	current := connector.call("trading_list_strategy_scripts", anotherConnection)

	assert.Equal(t, dto.ToolOutcomeSignInRequired, abandoned.Outcome,
		"那一份已經被放掉了——不是「失效」，是不在了")
	assert.Equal(t, dto.ToolOutcomeSucceeded, current.Outcome,
		"還用得到的那一份一根寒毛都不能動")
}
