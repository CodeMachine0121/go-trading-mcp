package domains

import (
	"strings"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

const connectorAuthorizationVerdictMemoryLimit = time.Minute

// ConnectorAuthorizationDomain is one connector authorization as the trading service
// judged it, and whether that judgement lets it act on this connector.
type ConnectorAuthorizationDomain struct {
	isActive  bool
	subject   string
	audience  string
	expiresAt time.Time
}

func NewConnectorAuthorizationDomain(
	inspection vo.ConnectorAuthorizationInspectionVo,
) ConnectorAuthorizationDomain {
	return ConnectorAuthorizationDomain{
		isActive:  inspection.IsActive,
		subject:   inspection.Subject,
		audience:  strings.TrimSuffix(strings.TrimSpace(inspection.Audience), "/"),
		expiresAt: inspection.ExpiresAt,
	}
}

// IsGrantedTo reports whether this authorization is live and was issued for exactly
// this protected resource; a trailing slash does not make it a different one.
func (connectorAuthorizationDomain ConnectorAuthorizationDomain) IsGrantedTo(
	protectedResourceUrl string,
) bool {
	return connectorAuthorizationDomain.isActive &&
		connectorAuthorizationDomain.audience == strings.TrimSuffix(protectedResourceUrl, "/")
}

// RememberedUntil is how long this judgement may be reused without asking again:
// at most a minute, and never past the authorization's own expiry.
func (connectorAuthorizationDomain ConnectorAuthorizationDomain) RememberedUntil(now time.Time) time.Time {
	memoryLimit := now.Add(connectorAuthorizationVerdictMemoryLimit)
	if connectorAuthorizationDomain.isActive &&
		!connectorAuthorizationDomain.expiresAt.IsZero() &&
		connectorAuthorizationDomain.expiresAt.Before(memoryLimit) {
		return connectorAuthorizationDomain.expiresAt
	}

	return memoryLimit
}

func (connectorAuthorizationDomain ConnectorAuthorizationDomain) ToDto() dto.ConnectorAuthorizationDto {
	return dto.ConnectorAuthorizationDto{
		Subject:   connectorAuthorizationDomain.subject,
		ExpiresAt: connectorAuthorizationDomain.expiresAt,
	}
}
