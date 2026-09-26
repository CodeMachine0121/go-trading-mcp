package _interface

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
)

//go:generate go tool mockgen -source=i_connector_authorization_verdict_repository.go -destination=mocks/mock_i_connector_authorization_verdict_repository.go -package=mocks

type IConnectorAuthorizationVerdictRepository interface {
	Find(accessToken string, now time.Time) (domains.ConnectorAuthorizationDomain, bool)
	Save(accessToken string, verdict domains.ConnectorAuthorizationDomain, now time.Time)
}
