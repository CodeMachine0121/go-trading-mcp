package domains_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/stretchr/testify/assert"
)

func TestEachWayOfFailingAsksTheReaderForSomethingDifferent(t *testing.T) {
	testCases := []struct {
		name            string
		cause           error
		expectedOutcome dto.ToolOutcome
	}{
		{"交易服務不認得這份外掛授權", domains.ErrReconnectRequired, dto.ToolOutcomeReconnectRequired},
		{"交易服務不在", domains.ErrTradingServiceUnreachable, dto.ToolOutcomeTradingServiceUnreachable},
		{"其他沒見過的狀況", errors.New("something else"), dto.ToolOutcomeTradingServiceUnreachable},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resultDto := domains.NewFailureReasonDomain(testCase.cause).ToToolResultDto()

			assert.Equal(t, testCase.expectedOutcome, resultDto.Outcome)
			assert.NotEmpty(t, resultDto.Content)
			assert.False(t, resultDto.Succeeded())
		})
	}
}

func TestAFailureWrappedInContextIsStillRecognisedForWhatItIs(t *testing.T) {
	wrapped := fmt.Errorf("代辦時：%w", domains.ErrReconnectRequired)

	resultDto := domains.NewFailureReasonDomain(wrapped).ToToolResultDto()

	assert.Equal(t, dto.ToolOutcomeReconnectRequired, resultDto.Outcome)
}

func TestNotBeingAbleToReachTheTradingServiceNeverReadsAsReconnecting(t *testing.T) {
	resultDto := domains.NewFailureReasonDomain(errors.New("dial tcp: i/o timeout")).ToToolResultDto()

	assert.NotContains(t, resultDto.Content, "重新連線")
	assert.Contains(t, resultDto.Content, "連不到交易服務")
}

func TestTheReasonIsSaidOnceRatherThanOnceForEveryLayerItPassedThrough(t *testing.T) {
	alreadySaid := fmt.Errorf("%w：dial tcp: connection refused", domains.ErrTradingServiceUnreachable)

	resultDto := domains.NewFailureReasonDomain(alreadySaid).ToToolResultDto()

	assert.Equal(t, dto.ToolOutcomeTradingServiceUnreachable, resultDto.Outcome)
	assert.Equal(t, 1,
		strings.Count(resultDto.Content, domains.ErrTradingServiceUnreachable.Error()),
		"每經過一層就再說一次，讀起來像壞了兩次")
}
