package controller

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// replyTo is how anything this connector has to say reaches the assistant — and, on
// the way past, what leaves a trace of having happened.
//
// One function for every ability and every failure, so that "was this a success?" is
// answered in one place. Marking it matters more than it looks: an assistant reads an
// unmarked refusal as the answer it asked for, and goes on to summarise a number that
// was never returned.
//
// The trace is taken here rather than at each of the four call sites for the same
// reason: a record written in four places is a record missing from one of them, and
// the one it is missing from is always the failure nobody can reproduce.
func replyTo(toolName string, resultDto dto.ToolResultDto) *mcp.CallToolResult {
	recordAttempt(toolName, resultDto)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultDto.Content}},
		IsError: !resultDto.Succeeded(),
	}
}
