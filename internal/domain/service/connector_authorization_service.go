package service

import (
	"context"

	_interface "github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
)

type ConnectorAuthorizationService struct {
	tradingServiceProxy                     _interface.ITradingServiceProxy
	connectorAuthorizationVerdictRepository _interface.IConnectorAuthorizationVerdictRepository
	clock                                   _interface.IClock
	protectedResourceUrl                    string
}

func NewConnectorAuthorizationService(
	tradingServiceProxy _interface.ITradingServiceProxy,
	connectorAuthorizationVerdictRepository _interface.IConnectorAuthorizationVerdictRepository,
	clock _interface.IClock,
	protectedResourceUrl string,
) *ConnectorAuthorizationService {
	return &ConnectorAuthorizationService{
		tradingServiceProxy:                     tradingServiceProxy,
		connectorAuthorizationVerdictRepository: connectorAuthorizationVerdictRepository,
		clock:                                   clock,
		protectedResourceUrl:                    protectedResourceUrl,
	}
}

func (connectorAuthorizationService *ConnectorAuthorizationService) VerifyConnectorAuthorization(
	ctx context.Context,
	accessToken string,
) (dto.ConnectorAuthorizationDto, error) {
	now := connectorAuthorizationService.clock.Now()

	verdict, isRemembered := connectorAuthorizationService.connectorAuthorizationVerdictRepository.Find(accessToken, now)
	if !isRemembered {
		inspection, inspectionError := connectorAuthorizationService.tradingServiceProxy.
			InspectConnectorAuthorization(ctx, accessToken)
		if inspectionError != nil {
			return dto.ConnectorAuthorizationDto{}, inspectionError
		}

		verdict = domains.NewConnectorAuthorizationDomain(inspection)
		connectorAuthorizationService.connectorAuthorizationVerdictRepository.Save(accessToken, verdict, now)
	}

	if !verdict.IsGrantedTo(connectorAuthorizationService.protectedResourceUrl) {
		return dto.ConnectorAuthorizationDto{}, domains.ErrConnectorAuthorizationRejected
	}

	return verdict.ToDto(), nil
}
