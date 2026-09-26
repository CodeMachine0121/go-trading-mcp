package application

import (
	"context"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
)

type ConnectorAuthorizationApplication struct {
	connectorAuthorizationService *service.ConnectorAuthorizationService
}

func NewConnectorAuthorizationApplication(
	connectorAuthorizationService *service.ConnectorAuthorizationService,
) *ConnectorAuthorizationApplication {
	return &ConnectorAuthorizationApplication{connectorAuthorizationService: connectorAuthorizationService}
}

func (connectorAuthorizationApplication *ConnectorAuthorizationApplication) VerifyConnectorAuthorization(
	ctx context.Context,
	accessToken string,
) (dto.ConnectorAuthorizationDto, error) {
	return connectorAuthorizationApplication.connectorAuthorizationService.
		VerifyConnectorAuthorization(ctx, accessToken)
}
