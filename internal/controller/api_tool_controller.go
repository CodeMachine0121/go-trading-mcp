package controller

import (
	"context"
	"encoding/json"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ApiToolController puts every ability in front of the assistant and turns one call
// into one ask.
//
// It registers the whole catalogue itself rather than handing the assembly root a
// list to loop over, because registering and describing are the same act: the shape
// the assistant sees is computed from the same declaration that will later build the
// request. Split apart, a box could be described one way and placed another.
type ApiToolController struct {
	apiToolApplication *application.ApiToolApplication
}

func NewApiToolController(apiToolApplication *application.ApiToolApplication) *ApiToolController {
	return &ApiToolController{apiToolApplication: apiToolApplication}
}

// RegisterOn puts every ability on the given server.
func (apiToolController *ApiToolController) RegisterOn(server *mcp.Server) {
	for _, definitionDto := range apiToolController.apiToolApplication.ListApiTools() {
		server.AddTool(apiToolController.toMcpTool(definitionDto), apiToolController.handle)
	}
}

// handle carries out whichever ability was called.
//
// One handler for all of them. The ability's name arrives on the call, and the
// catalogue already knows everything else about it, so a handler per ability would be
// fifty copies of these twelve lines.
func (apiToolController *ApiToolController) handle(
	ctx context.Context,
	request *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	arguments := map[string]json.RawMessage{}
	if len(request.Params.Arguments) > 0 {
		if decodeError := json.Unmarshal(request.Params.Arguments, &arguments); decodeError != nil {
			return replyTo(dto.ToolResultDto{
				Outcome: dto.ToolOutcomeInvalidArguments,
				Content: "送來的欄位不是一組可以讀的資料：" + decodeError.Error(),
			}), nil
		}
	}

	caller := callerOn(request)

	return replyTo(apiToolController.apiToolApplication.CallApiTool(ctx, dto.ToolCallDto{
		ToolName:            request.Params.Name,
		Arguments:           arguments,
		SessionKey:          caller.SessionKey(),
		SuppliedAccessToken: caller.SuppliedAccessToken(),
	})), nil
}

// toMcpTool is one ability in the shape the assistant is shown it: its name, what it
// is for, and the form to fill in.
//
// Describing and shaping are one act, not two. Both are computed from the same
// declaration that will later place the values, so what the assistant is shown and
// what is accepted cannot drift apart — and the note about signing in is appended
// here rather than written out by each of the fifty declarations, because a sentence
// written fifty times is a sentence missing from one of them.
func (apiToolController *ApiToolController) toMcpTool(
	definitionDto dto.ToolDefinitionDto,
) *mcp.Tool {
	properties := map[string]any{}
	required := make([]string, 0, len(definitionDto.Parameters))

	for _, parameter := range definitionDto.Parameters {
		properties[parameter.Name] = map[string]any{
			"type":        parameter.Kind,
			"description": parameter.Description,
		}

		if parameter.IsRequired {
			required = append(required, parameter.Name)
		}
	}

	description := definitionDto.Description
	if definitionDto.RequiresSignIn {
		description += "\n\n（需要身分：請先用 trading_sign_in 登入，" +
			"或在請求標頭帶上自己的 Authorization: Bearer 憑證。）"
	}

	return &mcp.Tool{
		Name:        definitionDto.Name,
		Description: description,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": properties,
			"required":   required,
		},
	}
}
