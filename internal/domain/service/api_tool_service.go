package service

import (
	"context"
	"fmt"

	_interface "github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// ApiToolService carries out one ability under the caller's own connector
// authorization, and owns no rule about market data.
type ApiToolService struct {
	apiToolsPerNames    map[string]domains.ApiToolDomain
	tradingServiceProxy _interface.ITradingServiceProxy
}

func NewApiToolService(
	apiTools []domains.ApiToolDomain,
	tradingServiceProxy _interface.ITradingServiceProxy,
) *ApiToolService {
	apiToolsPerNames := make(map[string]domains.ApiToolDomain, len(apiTools))
	for _, apiTool := range apiTools {
		apiToolsPerNames[apiTool.Name()] = apiTool
	}

	return &ApiToolService{
		apiToolsPerNames:    apiToolsPerNames,
		tradingServiceProxy: tradingServiceProxy,
	}
}

func (apiToolService *ApiToolService) ListApiTools() []dto.ToolDefinitionDto {
	definitionDtos := make([]dto.ToolDefinitionDto, 0, len(apiToolService.apiToolsPerNames))
	for _, apiTool := range apiToolService.apiToolsPerNames {
		definitionDtos = append(definitionDtos, apiTool.ToDefinitionDto())
	}

	return definitionDtos
}

// CallApiTool carries out one ability and says how it went. A refusal is a result,
// not an error; only failing to reach the trading service comes back as one.
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

	response, sendError := apiToolService.tradingServiceProxy.Send(ctx, request, toolCallDto.AccessToken)
	if sendError != nil {
		return domains.NewFailureReasonDomain(sendError).ToToolResultDto()
	}

	if response.Outcome == vo.TradingServiceIdentityNotRecognized {
		return domains.NewFailureReasonDomain(domains.ErrReconnectRequired).ToToolResultDto()
	}

	return apiTool.Relayed(response).ToToolResultDto()
}
