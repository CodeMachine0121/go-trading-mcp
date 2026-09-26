package domains

import (
	"strings"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

const connectorAuthorizationVerdictMemoryLimit = time.Minute

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

func (connectorAuthorizationDomain ConnectorAuthorizationDomain) IsGrantedTo(
	protectedResourceUrl string,
) bool {
	return connectorAuthorizationDomain.isActive &&
		connectorAuthorizationDomain.audience == strings.TrimSuffix(protectedResourceUrl, "/")
}

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
