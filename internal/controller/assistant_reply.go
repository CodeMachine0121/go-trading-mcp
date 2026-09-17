package controller

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// replyTo is how anything this connector has to say reaches the assistant.
//
// One function for every ability and every failure, so that "was this a success?" is
// answered in one place. Marking it matters more than it looks: an assistant reads an
// unmarked refusal as the answer it asked for, and goes on to summarise a number that
// was never returned.
func replyTo(resultDto dto.ToolResultDto) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultDto.Content}},
		IsError: !resultDto.Succeeded(),
	}
}
