package service

import (
	"context"
	"fmt"

	_interface "github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// ApiToolService carries out one ability, start to finish.
//
// Start to finish is the point. Finding the ability, working out whose identity to
// use, renewing it if it has run out, turning the filled-in form into an ask, sending
// it, and deciding what one particular refusal means — a caller that had to sequence
// those is a caller that could sequence them differently, and one of the orderings
// spends a single-use renewal twice.
//
// It owns no rule about market data. Every such rule belongs to the trading service,
// and this service's job is to let those rules answer for themselves.
type ApiToolService struct {
	apiToolsPerNames      map[string]domains.ApiToolDomain
	authenticationService *AuthenticationService
	tradingServiceProxy   _interface.ITradingServiceProxy
}

// NewApiToolService takes the whole catalogue at once, so that what this connector
// can do is settled at assembly and cannot drift afterwards.
func NewApiToolService(
	apiTools []domains.ApiToolDomain,
	authenticationService *AuthenticationService,
	tradingServiceProxy _interface.ITradingServiceProxy,
) *ApiToolService {
	apiToolsPerNames := make(map[string]domains.ApiToolDomain, len(apiTools))
	for _, apiTool := range apiTools {
		apiToolsPerNames[apiTool.Name()] = apiTool
	}

	return &ApiToolService{
		apiToolsPerNames:      apiToolsPerNames,
		authenticationService: authenticationService,
		tradingServiceProxy:   tradingServiceProxy,
	}
}

// ListApiTools is everything this connector can do, as the assistant is told about it.
func (apiToolService *ApiToolService) ListApiTools() []dto.ToolDefinitionDto {
	definitionDtos := make([]dto.ToolDefinitionDto, 0, len(apiToolService.apiToolsPerNames))
	for _, apiTool := range apiToolService.apiToolsPerNames {
		definitionDtos = append(definitionDtos, apiTool.ToDefinitionDto())
	}

	return definitionDtos
}

// CallApiTool carries out one ability and says how it went.
//
// A refusal is a result, not an error. The trading service answered, and its answer
// is the single most useful thing an assistant can be handed — it is what tells it
// whether to change the form and try again, or to stop. Only failing to reach the
// trading service at all comes back as nothing-happened.
func (apiToolService *ApiToolService) CallApiTool(
	ctx context.Context,
	toolCallDto dto.ToolCallDto,
) dto.ToolResultDto {
	apiTool, isKnown := apiToolService.apiToolsPerNames[toolCallDto.ToolName]
	if !isKnown {
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeUnknownTool,
			Content: fmt.Sprintf("這個外掛不會「%s」這件事", toolCallDto.ToolName),
		}
	}

	request, buildError := apiTool.BuildRequest(
		domains.NewToolArgumentsDomain(toolCallDto.Arguments))
	if buildError != nil {
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeInvalidArguments,
			Content: buildError.Error(),
		}
	}

	accessToken, identityFailure, isIdentified := apiToolService.identityFor(ctx, apiTool, toolCallDto)
	if !isIdentified {
		return identityFailure
	}

	response, sendError := apiToolService.tradingServiceProxy.Send(ctx, request, accessToken)
	if sendError != nil {
		return domains.NewFailureReasonDomain(sendError).ToToolResultDto()
	}

	if response.Outcome != vo.TradingServiceIdentityNotRecognized {
		return apiTool.Relayed(response).ToToolResultDto()
	}

	return apiToolService.retryWithRenewedIdentity(
		ctx, apiTool, request, toolCallDto, accessToken, response)
}

// identityFor settles whose identity this ask travels under.
//
// The caller's own proof wins when there is one: that is the identity the person
// named out loud on this very ask, and second-guessing it would mean acting as
// somebody they did not name. It is used exactly as given — the renewal half is not
// in this connector's hands, so there is nothing here that could renew it.
func (apiToolService *ApiToolService) identityFor(
	ctx context.Context,
	apiTool domains.ApiToolDomain,
	toolCallDto dto.ToolCallDto,
) (string, dto.ToolResultDto, bool) {
	if toolCallDto.SuppliedAccessToken != "" {
		return toolCallDto.SuppliedAccessToken, dto.ToolResultDto{}, true
	}

	if !apiTool.RequiresSignIn() {
		return "", dto.ToolResultDto{}, true
	}

	accessToken, identityError := apiToolService.authenticationService.UsableAccessToken(
		ctx, vo.NewSessionKeyVo(toolCallDto.SessionKey))
	if identityError != nil {
		return "", domains.NewFailureReasonDomain(identityError).ToToolResultDto(), false
	}

	return accessToken, dto.ToolResultDto{}, true
}

// retryWithRenewedIdentity is what happens when the trading service says it does not
// know who we are, although our own arithmetic said the proof was still good.
//
// Its answer wins over the arithmetic — clocks disagree, and a signing-in can be
// revoked from somewhere else. So the renewal is spent for real and the ask is made
// once more. Once, not in a loop: a second rejection is the trading service saying the
// same thing twice, and a third ask would only be slower.
//
// A proof the caller brought along is never retried. There is no renewal half to
// spend, so the retry would send the identical rejected proof again.
func (apiToolService *ApiToolService) retryWithRenewedIdentity(
	ctx context.Context,
	apiTool domains.ApiToolDomain,
	request vo.TradingServiceRequestVo,
	toolCallDto dto.ToolCallDto,
	rejectedAccessToken string,
	firstResponse vo.TradingServiceResponseVo,
) dto.ToolResultDto {
	if toolCallDto.SuppliedAccessToken != "" {
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeSignInExpired,
			Content: domains.ErrSignInExpired.Error(),
		}
	}

	renewedAccessToken, renewalError := apiToolService.authenticationService.RenewedAccessToken(
		ctx, vo.NewSessionKeyVo(toolCallDto.SessionKey), rejectedAccessToken)
	if renewalError != nil {
		return domains.NewFailureReasonDomain(renewalError).ToToolResultDto()
	}

	response, sendError := apiToolService.tradingServiceProxy.Send(ctx, request, renewedAccessToken)
	if sendError != nil {
		return domains.NewFailureReasonDomain(sendError).ToToolResultDto()
	}

	if response.Outcome == vo.TradingServiceIdentityNotRecognized {
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeSignInExpired,
			Content: firstResponse.Content,
		}
	}

	return apiTool.Relayed(response).ToToolResultDto()
}
