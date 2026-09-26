package domains_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
)

const thisConnector = "https://trading-mcp.example.com/mcp"

var judgedAt = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

func TestAConnectorAuthorizationCountsOnlyWhenLiveAndIssuedForThisConnector(t *testing.T) {
	testCases := []struct {
		name          string
		inspection    vo.ConnectorAuthorizationInspectionVo
		expectedGrant bool
	}{
		{"有效且發給這個外掛", vo.ConnectorAuthorizationInspectionVo{IsActive: true, Audience: thisConnector}, true},
		{"對象多了結尾斜線", vo.ConnectorAuthorizationInspectionVo{IsActive: true, Audience: thisConnector + "/"}, true},
		{"發給別的服務", vo.ConnectorAuthorizationInspectionVo{IsActive: true, Audience: "https://elsewhere.example.com/mcp"}, false},
		{"沒有對象（網站登入的憑證）", vo.ConnectorAuthorizationInspectionVo{IsActive: true, Audience: ""}, false},
		{"交易服務判定失效", vo.ConnectorAuthorizationInspectionVo{IsActive: false, Audience: thisConnector}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			authorization := domains.NewConnectorAuthorizationDomain(testCase.inspection)

			assert.Equal(t, testCase.expectedGrant, authorization.IsGrantedTo(thisConnector))
		})
	}
}

func TestThisConnectorsAddressWithATrailingSlashIsStillTheSameConnector(t *testing.T) {
	authorization := domains.NewConnectorAuthorizationDomain(
		vo.ConnectorAuthorizationInspectionVo{IsActive: true, Audience: thisConnector})

	assert.True(t, authorization.IsGrantedTo(thisConnector+"/"))
}

func TestAJudgementIsRememberedForAMinuteAtMostAndNeverPastTheExpiry(t *testing.T) {
	testCases := []struct {
		name                    string
		inspection              vo.ConnectorAuthorizationInspectionVo
		expectedRememberedUntil time.Time
	}{
		{"到期還很久", vo.ConnectorAuthorizationInspectionVo{IsActive: true, ExpiresAt: judgedAt.Add(15 * time.Minute)},
			judgedAt.Add(time.Minute)},
		{"二十秒後到期", vo.ConnectorAuthorizationInspectionVo{IsActive: true, ExpiresAt: judgedAt.Add(20 * time.Second)},
			judgedAt.Add(20 * time.Second)},
		{"失效的判定", vo.ConnectorAuthorizationInspectionVo{IsActive: false},
			judgedAt.Add(time.Minute)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			authorization := domains.NewConnectorAuthorizationDomain(testCase.inspection)

			assert.Equal(t, testCase.expectedRememberedUntil, authorization.RememberedUntil(judgedAt))
		})
	}
}

func TestACountingAuthorizationSaysWhoseItIsAndUntilWhen(t *testing.T) {
	authorizationDto := domains.NewConnectorAuthorizationDomain(vo.ConnectorAuthorizationInspectionVo{
		IsActive: true, Subject: "42", Audience: thisConnector, ExpiresAt: judgedAt,
	}).ToDto()

	assert.Equal(t, "42", authorizationDto.Subject)
	assert.Equal(t, judgedAt, authorizationDto.ExpiresAt)
}
