package tradingservice

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

type connectorAuthorizationInspectionWire struct {
	Active     bool   `json:"active"`
	Subject    string `json:"sub"`
	Audience   string `json:"aud"`
	Expiration int64  `json:"exp"`
}

func (connectorAuthorizationInspectionWire connectorAuthorizationInspectionWire) ToConnectorAuthorizationInspectionVo() vo.ConnectorAuthorizationInspectionVo {
	expiresAt := time.Time{}
	if connectorAuthorizationInspectionWire.Expiration > 0 {
		expiresAt = time.Unix(connectorAuthorizationInspectionWire.Expiration, 0)
	}

	return vo.ConnectorAuthorizationInspectionVo{
		IsActive:  connectorAuthorizationInspectionWire.Active,
		Subject:   connectorAuthorizationInspectionWire.Subject,
		Audience:  connectorAuthorizationInspectionWire.Audience,
		ExpiresAt: expiresAt,
	}
}
