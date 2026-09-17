package controller

import (
	"log/slog"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
)

// recordAttempt leaves a trace of one attempt: which ability, and how it ended.
//
// **Three fields, and the choice of three is the whole design.** What is left out —
// the filled-in boxes, the answer, the proof it travelled under, who it belonged to —
// is left out because a record is the one artifact that outlives the process, gets
// copied into a bug report and pasted into a chat. A password that reached a log has
// been disclosed, and no later deletion undoes that.
//
// What is kept is enough to answer the only questions anybody asks of a connector
// that is misbehaving: which ability, how often, and which kind of failure. The kind
// matters more than the message: "refused" is the trading service disagreeing and
// needs no action here, while "unreachable" is this machine failing to reach that one.
func recordAttempt(toolName string, resultDto dto.ToolResultDto) {
	slog.Info("代辦一件事",
		slog.String("ability", toolName),
		slog.Bool("succeeded", resultDto.Succeeded()),
		slog.String("outcome", string(resultDto.Outcome)),
	)
}
