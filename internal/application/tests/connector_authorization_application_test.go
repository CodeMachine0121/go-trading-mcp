package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface/mocks"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const protectedResourceUrl = "https://trading-mcp.example.com/mcp"

var firstCallAt = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

type gatekeeper struct {
	connectorAuthorization *application.ConnectorAuthorizationApplication
	tradingService         *mocks.MockITradingServiceProxy
	now                    *time.Time
}

func newGatekeeper(t *testing.T) *gatekeeper {
	t.Helper()

	controller := gomock.NewController(t)
	tradingServiceProxy := mocks.NewMockITradingServiceProxy(controller)
	now := firstCallAt
	clock := mocks.NewMockIClock(controller)
	clock.EXPECT().Now().DoAndReturn(func() time.Time { return now }).AnyTimes()

	return &gatekeeper{
		connectorAuthorization: application.NewConnectorAuthorizationApplication(
			service.NewConnectorAuthorizationService(
				tradingServiceProxy,
				persistence.NewConnectorAuthorizationVerdictRepository(),
				clock,
				protectedResourceUrl,
			)),
		tradingService: tradingServiceProxy,
		now:            &now,
	}
}

func (gatekeeper *gatekeeper) verify() error {
	_, verificationError := gatekeeper.connectorAuthorization.VerifyConnectorAuthorization(
		context.Background(), jamesAccessToken)

	return verificationError
}

func (gatekeeper *gatekeeper) later(elapsed time.Duration) {
	*gatekeeper.now = firstCallAt.Add(elapsed)
}

func judgedLive(audience string, expiresAt time.Time) vo.ConnectorAuthorizationInspectionVo {
	return vo.ConnectorAuthorizationInspectionVo{
		IsActive: true, Subject: "42", Audience: audience, ExpiresAt: expiresAt}
}

func TestAConnectorAuthorizationIsLetInOnlyWhenTheTradingServiceSaysItCountsHere(t *testing.T) {
	testCases := []struct {
		name          string
		inspection    vo.ConnectorAuthorizationInspectionVo
		expectedError error
	}{
		{"有效且發給這個外掛", judgedLive(protectedResourceUrl, firstCallAt.Add(15*time.Minute)), nil},
		{"只差結尾斜線", judgedLive(protectedResourceUrl+"/", firstCallAt.Add(15*time.Minute)), nil},
		{"發給別的服務", judgedLive("https://elsewhere.example.com/mcp", firstCallAt.Add(15*time.Minute)),
			domains.ErrConnectorAuthorizationRejected},
		{"交易服務判定失效", vo.ConnectorAuthorizationInspectionVo{IsActive: false},
			domains.ErrConnectorAuthorizationRejected},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gatekeeper := newGatekeeper(t)
			gatekeeper.tradingService.EXPECT().
				InspectConnectorAuthorization(gomock.Any(), jamesAccessToken).
				Return(testCase.inspection, nil)

			verificationError := gatekeeper.verify()

			if testCase.expectedError == nil {
				assert.NoError(t, verificationError)
			} else {
				assert.ErrorIs(t, verificationError, testCase.expectedError)
			}
		})
	}
}

func TestALetInAuthorizationSaysWhoseItIs(t *testing.T) {
	gatekeeper := newGatekeeper(t)
	gatekeeper.tradingService.EXPECT().
		InspectConnectorAuthorization(gomock.Any(), jamesAccessToken).
		Return(judgedLive(protectedResourceUrl, firstCallAt.Add(15*time.Minute)), nil)

	authorizationDto, verificationError := gatekeeper.connectorAuthorization.VerifyConnectorAuthorization(
		context.Background(), jamesAccessToken)

	require.NoError(t, verificationError)
	assert.Equal(t, "42", authorizationDto.Subject)
	assert.Equal(t, firstCallAt.Add(15*time.Minute), authorizationDto.ExpiresAt)
}

func TestATradingServiceThatCannotBeReachedIsNotToldAsARejectedAuthorization(t *testing.T) {
	gatekeeper := newGatekeeper(t)
	gatekeeper.tradingService.EXPECT().
		InspectConnectorAuthorization(gomock.Any(), jamesAccessToken).
		Return(vo.ConnectorAuthorizationInspectionVo{}, domains.ErrTradingServiceUnreachable)

	verificationError := gatekeeper.verify()

	assert.ErrorIs(t, verificationError, domains.ErrTradingServiceUnreachable)
	assert.NotErrorIs(t, verificationError, domains.ErrConnectorAuthorizationRejected)
}

func TestTheSameAuthorizationIsAskedAboutAgainOnlyOnceItsJudgementRunsOut(t *testing.T) {
	testCases := []struct {
		name               string
		expiresAt          time.Time
		secondCallAfter    time.Duration
		expectedAskedTimes int
	}{
		{"三十秒後再來", firstCallAt.Add(15 * time.Minute), 30 * time.Second, 1},
		{"滿一分鐘再來", firstCallAt.Add(15 * time.Minute), time.Minute, 2},
		{"二十秒後到期、二十五秒後再來", firstCallAt.Add(20 * time.Second), 25 * time.Second, 2},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gatekeeper := newGatekeeper(t)
			gatekeeper.tradingService.EXPECT().
				InspectConnectorAuthorization(gomock.Any(), jamesAccessToken).
				Return(judgedLive(protectedResourceUrl, testCase.expiresAt), nil).
				Times(testCase.expectedAskedTimes)

			require.NoError(t, gatekeeper.verify())
			gatekeeper.later(testCase.secondCallAfter)
			_ = gatekeeper.verify()
		})
	}
}

func TestNotBeingAbleToAskIsNeverRemembered(t *testing.T) {
	gatekeeper := newGatekeeper(t)
	gomock.InOrder(
		gatekeeper.tradingService.EXPECT().
			InspectConnectorAuthorization(gomock.Any(), jamesAccessToken).
			Return(vo.ConnectorAuthorizationInspectionVo{}, domains.ErrTradingServiceUnreachable),
		gatekeeper.tradingService.EXPECT().
			InspectConnectorAuthorization(gomock.Any(), jamesAccessToken).
			Return(judgedLive(protectedResourceUrl, firstCallAt.Add(15*time.Minute)), nil),
	)

	assert.Error(t, gatekeeper.verify())
	gatekeeper.later(time.Second)
	assert.NoError(t, gatekeeper.verify())
}
