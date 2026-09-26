package vo

import "github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"

// TradingServiceOutcome is the trading service's verdict on one ask, in the only
// three flavours the domain has to tell apart.
//
// Being unreachable is deliberately not one of them: that is not the trading
// service's verdict, it is the absence of one, and it travels back as an error.
type TradingServiceOutcome string

const (
	// TradingServiceSucceeded means it did the thing and this is what came back.
	TradingServiceSucceeded TradingServiceOutcome = "succeeded"
	// TradingServiceRefused means it declined, and Content says why in its own words.
	TradingServiceRefused TradingServiceOutcome = "refused"
	// TradingServiceIdentityNotRecognized is the one refusal the person fixes by reconnecting, not by rewording.
	TradingServiceIdentityNotRecognized TradingServiceOutcome = "identityNotRecognized"
)

// TradingServiceResponseVo is what came back from one ask.
//
// Content is the trading service's own words, carried through untouched. Rewording
// it here would put a second author on a sentence the assistant has to act on, and
// the second author is the one who does not know the rules.
type TradingServiceResponseVo struct {
	Outcome TradingServiceOutcome
	Content string
}

// ToToolResultDto is this answer in the shape the caller is given it, in the trading
// service's own words either way.
func (tradingServiceResponseVo TradingServiceResponseVo) ToToolResultDto() dto.ToolResultDto {
	if tradingServiceResponseVo.Outcome == TradingServiceSucceeded {
		return dto.ToolResultDto{
			Outcome: dto.ToolOutcomeSucceeded,
			Content: tradingServiceResponseVo.Content,
		}
	}

	return dto.ToolResultDto{
		Outcome: dto.ToolOutcomeRefusedByTradingService,
		Content: tradingServiceResponseVo.Content,
	}
}
