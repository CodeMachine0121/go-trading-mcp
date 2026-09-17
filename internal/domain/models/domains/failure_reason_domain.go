package domains

import (
	"errors"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
)

// FailureReasonDomain is one thing that went wrong, and which of the four things it
// asks the reader to do about it.
//
// It exists because that judgement was being made in three places — carrying out an
// ability, signing in, signing out — and the three had already started to word the
// same failure differently. They are four genuinely different instructions:
//
//	請先登入          你從來沒有登入過
//	登入已失效        你登入過，但那一份救不回來了
//	連不到交易服務    你送的沒有錯，是它現在不在；等一下再送同一件事
//	（其他）          這個外掛自己出了狀況
//
// Getting one of them wrong is never harmless: telling somebody who is signed in to
// sign in again makes this connector look like it forgets people, and telling
// somebody whose network blipped to sign in again makes them type a password for
// nothing.
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
	case errors.Is(failureReasonDomain.cause, ErrSignInRequired):
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeSignInRequired,
			Content: ErrSignInRequired.Error(),
		}
	case errors.Is(failureReasonDomain.cause, ErrSignInExpired):
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeSignInExpired,
			Content: ErrSignInExpired.Error(),
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
