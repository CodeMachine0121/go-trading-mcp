package domains_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aReplayResultWith is a replay result with that many equity points and trades, point
// and trade n carrying n so a test can see which ones were kept.
func aReplayResultWith(equityPointCount int, closedTradeCount int, extraFields string) string {
	equityPoints := make([]string, 0, equityPointCount)
	for pointIndex := range equityPointCount {
		equityPoints = append(equityPoints, fmt.Sprintf(`{"openTime":"t%d","equity":"%d"}`, pointIndex, pointIndex))
	}

	closedTrades := make([]string, 0, closedTradeCount)
	for tradeIndex := range closedTradeCount {
		closedTrades = append(closedTrades, fmt.Sprintf(`{"profit":"%d"}`, tradeIndex))
	}

	return `{"symbol":"BTCUSDT","summary":{"finalEquity":"10123.456789012345678901","profitFactor":2.5}` +
		`,"closedTrades":[` + strings.Join(closedTrades, ",") + `]` +
		`,"equityCurve":[` + strings.Join(equityPoints, ",") + `]` + extraFields + `}`
}

type condensedReplay struct {
	Summary                    json.RawMessage   `json:"summary"`
	EquityCurve                []json.RawMessage `json:"equityCurve"`
	EquityCurvePointTotalCount *int              `json:"equityCurvePointTotalCount"`
	ClosedTrades               []json.RawMessage `json:"closedTrades"`
	ClosedTradeTotalCount      *int              `json:"closedTradeTotalCount"`
	InSample                   *condensedReplay  `json:"inSample"`
	Validation                 *condensedReplay  `json:"validation"`
}

func condensedOf(t *testing.T, content string) condensedReplay {
	t.Helper()

	condensed := condensedReplay{}
	require.NoError(t, json.Unmarshal(
		[]byte(domains.NewReplayResultDomain(content).ToCondensedContent()), &condensed))

	return condensed
}

func TestALongEquityCurveIsSampledKeepingBothEnds(t *testing.T) {
	condensed := condensedOf(t, aReplayResultWith(5000, 0, ""))

	require.Len(t, condensed.EquityCurve, 200)
	assert.JSONEq(t, `{"openTime":"t0","equity":"0"}`, string(condensed.EquityCurve[0]))
	assert.JSONEq(t, `{"openTime":"t4999","equity":"4999"}`, string(condensed.EquityCurve[199]))
	require.NotNil(t, condensed.EquityCurvePointTotalCount)
	assert.Equal(t, 5000, *condensed.EquityCurvePointTotalCount)
}

func TestAShortReplayResultIsHandedOnExactlyAsItArrived(t *testing.T) {
	for _, content := range []string{
		aReplayResultWith(200, 100, ""),
		aReplayResultWith(3, 1, ""),
	} {
		assert.Equal(t, content, domains.NewReplayResultDomain(content).ToCondensedContent())
	}
}

func TestOnlyTheMostRecentTradesAreKeptAndTheTotalIsSaid(t *testing.T) {
	condensed := condensedOf(t, aReplayResultWith(10, 350, ""))

	require.Len(t, condensed.ClosedTrades, 100)
	assert.JSONEq(t, `{"profit":"250"}`, string(condensed.ClosedTrades[0]))
	assert.JSONEq(t, `{"profit":"349"}`, string(condensed.ClosedTrades[99]))
	require.NotNil(t, condensed.ClosedTradeTotalCount)
	assert.Equal(t, 350, *condensed.ClosedTradeTotalCount)
	assert.Nil(t, condensed.EquityCurvePointTotalCount)
}

func TestTheInSampleAndValidationPartsAreCondensedByTheSameRules(t *testing.T) {
	condensed := condensedOf(t, aReplayResultWith(10, 1,
		`,"inSample":`+aReplayResultWith(50, 1, "")+`,"validation":`+aReplayResultWith(1000, 120, "")))

	require.NotNil(t, condensed.Validation)
	assert.Len(t, condensed.Validation.EquityCurve, 200)
	assert.Len(t, condensed.Validation.ClosedTrades, 100)
	require.NotNil(t, condensed.InSample)
	assert.Len(t, condensed.InSample.EquityCurve, 50)
	assert.Nil(t, condensed.InSample.EquityCurvePointTotalCount)
}

func TestTheReportCardIsNeverTouched(t *testing.T) {
	condensed := condensedOf(t, aReplayResultWith(5000, 350, ""))

	assert.Equal(t, `{"finalEquity":"10123.456789012345678901","profitFactor":2.5}`, string(condensed.Summary))
}

func TestWhatIsNotAReplayResultIsHandedOnUntouched(t *testing.T) {
	for _, content := range []string{"已完成（交易服務沒有回傳內容）", "[1,2,3]", `{"message":"驗證起點必須落在期間之內"}`} {
		assert.Equal(t, content, domains.NewReplayResultDomain(content).ToCondensedContent())
	}
}

func TestCondensingDoesNotEscapeWhatTheTradingServiceWrote(t *testing.T) {
	content := domains.NewReplayResultDomain(
		aReplayResultWith(300, 0, `,"note":"<短線> & 合約"`)).ToCondensedContent()

	assert.Contains(t, content, `"note":"<短線> & 合約"`)
}
