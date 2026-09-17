package dto

// ToolOutcome is how one attempt at an ability ended.
//
// Six words rather than a success flag and a message, because the four failures ask
// the reader to do four different things: change what you sent, sign in, sign in
// again, or wait and retry. One word for all of them would leave the assistant
// guessing which, and guessing wrong is either an endless retry or a needless
// password prompt.
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
	// ToolOutcomeSignInRequired means this ability needs to know who is asking and
	// nobody has said.
	ToolOutcomeSignInRequired ToolOutcome = "signInRequired"
	// ToolOutcomeSignInExpired means somebody said, but that signing-in is past
	// saving and the person has to sign in again.
	ToolOutcomeSignInExpired ToolOutcome = "signInExpired"
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
