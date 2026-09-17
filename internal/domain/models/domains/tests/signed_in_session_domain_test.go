package domains_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
)

var noon = time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)

func sessionExpiringAt(accessExpiry time.Time, refreshExpiry time.Time) domains.SignedInSessionDomain {
	return domains.NewSignedInSessionDomain("james@example.com", vo.NewTokenPairVo(
		"access-token", accessExpiry, "refresh-token", refreshExpiry))
}

func TestASigningInIsUsableUntilItIsCloseEnoughToExpiryToArriveLate(t *testing.T) {
	testCases := []struct {
		name         string
		accessExpiry time.Time
		expectUsable bool
	}{
		{"還有很久", noon.Add(15 * time.Minute), true},
		{"還有一分鐘，趕得上", noon.Add(1 * time.Minute), true},
		{"只剩十秒，路上就過期了", noon.Add(10 * time.Second), false},
		{"剛好到期", noon, false},
		{"已經過期", noon.Add(-1 * time.Minute), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			signedInSession := sessionExpiringAt(testCase.accessExpiry, noon.Add(30*24*time.Hour))

			assert.Equal(t, testCase.expectUsable, signedInSession.IsAccessTokenUsable(noon))
		})
	}
}

func TestThereIsAWayBackUntilTheRenewalItselfHasRunOut(t *testing.T) {
	testCases := []struct {
		name            string
		refreshExpiry   time.Time
		expectRenewable bool
	}{
		{"續用還有三十天", noon.Add(30 * 24 * time.Hour), true},
		{"續用還有一秒", noon.Add(1 * time.Second), true},
		{"續用剛好到期", noon, false},
		{"續用已經過期", noon.Add(-1 * time.Second), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			signedInSession := sessionExpiringAt(noon.Add(-time.Hour), testCase.refreshExpiry)

			assert.Equal(t, testCase.expectRenewable, signedInSession.IsRenewable(noon))
		})
	}
}

func TestRenewingReplacesBothHalvesAndKeepsWhoItBelongsTo(t *testing.T) {
	signedInSession := sessionExpiringAt(noon.Add(-time.Hour), noon.Add(24*time.Hour))

	renewedSession := signedInSession.WithRenewedTokens(vo.NewTokenPairVo(
		"fresh-access", noon.Add(15*time.Minute), "fresh-refresh", noon.Add(30*24*time.Hour)))

	assert.Equal(t, "fresh-access", renewedSession.AccessToken())
	assert.Equal(t, "fresh-refresh", renewedSession.RefreshToken())
	assert.Equal(t, "james@example.com", renewedSession.Email())
	assert.True(t, renewedSession.IsAccessTokenUsable(noon))
	assert.Equal(t, "access-token", signedInSession.AccessToken(),
		"換新不該改到原本那一份——半新半舊的一對正是交易服務讀成盜用的狀態")
}

func TestWhatACallerIsToldNeverIncludesEitherProof(t *testing.T) {
	signedInSession := sessionExpiringAt(noon.Add(15*time.Minute), noon.Add(30*24*time.Hour))

	signedInSessionDto := signedInSession.ToDto()

	assert.Equal(t, "james@example.com", signedInSessionDto.Email)
	assert.Equal(t, noon.Add(15*time.Minute), signedInSessionDto.AccessTokenExpiresAt)
	everythingTheCallerIsTold := fmt.Sprintf("%+v", signedInSessionDto)
	assert.NotContains(t, everythingTheCallerIsTold, "access-token")
	assert.NotContains(t, everythingTheCallerIsTold, "refresh-token")
}
