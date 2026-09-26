package persistence_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
)

var savedAt = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

func aLiveVerdict() domains.ConnectorAuthorizationDomain {
	return domains.NewConnectorAuthorizationDomain(vo.ConnectorAuthorizationInspectionVo{
		IsActive: true, Subject: "42", ExpiresAt: savedAt.Add(15 * time.Minute)})
}

func TestAJudgementIsFoundOnlyUntilItsTimeRunsOut(t *testing.T) {
	testCases := []struct {
		name          string
		askedAt       time.Time
		expectedFound bool
	}{
		{"三十秒後", savedAt.Add(30 * time.Second), true},
		{"剛好一分鐘", savedAt.Add(time.Minute), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := persistence.NewConnectorAuthorizationVerdictRepository()
			repository.Save("token-a", aLiveVerdict(), savedAt)

			found, isFound := repository.Find("token-a", testCase.askedAt)

			assert.Equal(t, testCase.expectedFound, isFound)
			if testCase.expectedFound {
				assert.Equal(t, "42", found.ToDto().Subject)
			}
		})
	}
}

func TestAJudgementBelongsOnlyToTheTokenItWasMadeAbout(t *testing.T) {
	repository := persistence.NewConnectorAuthorizationVerdictRepository()
	repository.Save("token-a", aLiveVerdict(), savedAt)

	_, isFound := repository.Find("token-b", savedAt)

	assert.False(t, isFound)
}

func TestSavingAgainAfterOthersRanOutKeepsTheFreshOne(t *testing.T) {
	repository := persistence.NewConnectorAuthorizationVerdictRepository()
	repository.Save("token-a", aLiveVerdict(), savedAt)
	repository.Save("token-b", aLiveVerdict(), savedAt.Add(2*time.Minute))

	_, isStaleFound := repository.Find("token-a", savedAt.Add(2*time.Minute))
	_, isFreshFound := repository.Find("token-b", savedAt.Add(2*time.Minute))

	assert.False(t, isStaleFound)
	assert.True(t, isFreshFound)
}
