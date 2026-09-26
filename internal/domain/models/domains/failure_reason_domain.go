package domains

import (
	"errors"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
)

type FailureReasonDomain struct {
	cause error
}

// NewFailureReasonDomain takes whatever went wrong.
func NewFailureReasonDomain(cause error) FailureReasonDomain {
	return FailureReasonDomain{cause: cause}
}

// ToToolResultDto is this failure in the shape a caller is given it.
func (failureReasonDomain FailureReasonDomain) ToToolResultDto() dto.ToolResultDto {
	switch {
	case errors.Is(failureReasonDomain.cause, ErrReconnectRequired):
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeReconnectRequired,
			Content: ErrReconnectRequired.Error(),
		}
	case errors.Is(failureReasonDomain.cause, ErrTradingServiceUnreachable):
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeTradingServiceUnreachable,
			Content: failureReasonDomain.cause.Error(),
		}
	default:
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeTradingServiceUnreachable,
			Content: ErrTradingServiceUnreachable.Error() + "：" + failureReasonDomain.cause.Error(),
		}
	}
}
