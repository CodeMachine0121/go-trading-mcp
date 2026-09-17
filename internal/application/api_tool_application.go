package application

import (
	"context"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
)

// ApiToolApplication is what an assistant's two questions come down to: what can you
// do, and please do this one.
type ApiToolApplication struct {
	apiToolService *service.ApiToolService
}

func NewApiToolApplication(apiToolService *service.ApiToolService) *ApiToolApplication {
	return &ApiToolApplication{apiToolService: apiToolService}
}

// ListApiTools is everything this connector can do.
func (apiToolApplication *ApiToolApplication) ListApiTools() []dto.ToolDefinitionDto {
	return apiToolApplication.apiToolService.ListApiTools()
}

// CallApiTool carries out one ability.
func (apiToolApplication *ApiToolApplication) CallApiTool(
	ctx context.Context,
	toolCallDto dto.ToolCallDto,
) dto.ToolResultDto {
	return apiToolApplication.apiToolService.CallApiTool(ctx, toolCallDto)
}
