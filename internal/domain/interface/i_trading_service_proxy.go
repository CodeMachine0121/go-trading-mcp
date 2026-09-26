package _interface

import (
	"context"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

//go:generate go tool mockgen -source=i_trading_service_proxy.go -destination=mocks/mock_i_trading_service_proxy.go -package=mocks

// ITradingServiceProxy is the one way out of this connector.
//
// Everything about how the trading service is actually spoken to lives behind it —
// including the one ability that stays on a line instead of asking once. The domain
// only ever sends a request and gets a verdict back; the staying is written on the
// request as a wait limit and honoured on the other side. That is what keeps the
// domain free of a branch it would otherwise have to remember to keep in step.
type ITradingServiceProxy interface {
	// Send carries out one ask. The error is reserved for not reaching the trading
	// service at all; a refusal is an answer and comes back as one.
	Send(
		ctx context.Context,
		request vo.TradingServiceRequestVo,
		accessToken string,
	) (vo.TradingServiceResponseVo, error)
}
