package service

import (
	"context"

	_interface "github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
)

// ConnectorAuthorizationService decides whether a connector authorization lets its
// bearer use this connector, asking the trading service only when no recent
// judgement is remembered.
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

// VerifyConnectorAuthorization answers ErrConnectorAuthorizationRejected for one that
// does not count here, and ErrTradingServiceUnreachable, never remembered, when no
// judgement could be had.
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
