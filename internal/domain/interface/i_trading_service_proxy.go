package _interface

import (
	"context"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
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
//
// Signing in, renewing and revoking are named separately rather than being three
// more requests, because they are the only three answers this connector reads for
// itself. Everything else it carries through untouched.
type ITradingServiceProxy interface {
	// Send carries out one ask. The error is reserved for not reaching the trading
	// service at all; a refusal is an answer and comes back as one.
	Send(
		ctx context.Context,
		request vo.TradingServiceRequestVo,
		accessToken string,
	) (vo.TradingServiceResponseVo, error)

	// SignIn exchanges an account for a pair of proofs.
	SignIn(ctx context.Context, signInDto dto.SignInDto) (vo.SessionGrantVo, error)

	// RenewSession spends a renewal proof on a fresh pair. The spent one is void the
	// moment this returns, whatever it returns.
	RenewSession(ctx context.Context, refreshToken string) (vo.SessionGrantVo, error)

	// RevokeSession voids a whole renewal chain. Revoking one that is already gone is
	// success, not failure — the aim has been achieved.
	RevokeSession(ctx context.Context, refreshToken string) (vo.TradingServiceResponseVo, error)
}
