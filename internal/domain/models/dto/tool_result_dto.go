package dto

// Distinct failures because each asks for a different fix: reword, reconnect, or retry later.
type ToolOutcome string

const (
	// ToolOutcomeSucceeded means it was done, and Content is what came back.
	ToolOutcomeSucceeded ToolOutcome = "succeeded"
	// ToolOutcomeRefusedByTradingService means the trading service declined, in its
	// own words.
	ToolOutcomeRefusedByTradingService ToolOutcome = "refusedByTradingService"
	// ToolOutcomeTradingServiceUnreachable means this connector could not reach the
	// trading service at all. Explicitly not the caller's fault.
	ToolOutcomeTradingServiceUnreachable ToolOutcome = "tradingServiceUnreachable"
	// ToolOutcomeReconnectRequired means the person must reconnect from Claude Code.
	ToolOutcomeReconnectRequired ToolOutcome = "reconnectRequired"
	// ToolOutcomeUnknownTool means this connector does not do that.
	ToolOutcomeUnknownTool ToolOutcome = "unknownTool"
	// ToolOutcomeInvalidArguments means a required box was left empty, and Content
	// says which one.
	ToolOutcomeInvalidArguments ToolOutcome = "invalidArguments"
)

// ToolResultDto is how one attempt ended and what there is to show for it.
//
// A refusal is a result, not an error: the trading service answered, and its answer
// is the most useful thing the assistant can be given. Only the connector's own
// breakages travel as errors.
type ToolResultDto struct {
	Outcome ToolOutcome
	Content string
}

// Succeeded reports whether there is an answer here rather than a reason there is not.
func (toolResultDto ToolResultDto) Succeeded() bool {
	return toolResultDto.Outcome == ToolOutcomeSucceeded
}
