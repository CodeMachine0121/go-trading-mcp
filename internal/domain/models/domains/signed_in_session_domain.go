package domains

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// accessTokenExpiryMargin is how early a signing-in is treated as spent.
//
// Without it, a proof handed over at the last legal instant arrives expired, because
// the journey takes time and two clocks never agree exactly. The cost of being early
// is one extra renewal; the cost of being late is a refusal the user sees.
const accessTokenExpiryMargin = 30 * time.Second

// SignedInSessionDomain is one connection's identity, and everything about whether
// it can still be used.
//
// Renewing returns a new one rather than changing this one. A half-renewed identity —
// new proof, old renewal half — is the exact state the trading service reads as theft
// and answers by voiding the whole chain and signing the real person out.
type SignedInSessionDomain struct {
	email  string
	tokens vo.TokenPairVo
}

// NewSignedInSessionDomain takes one signing-in as the trading service granted it.
func NewSignedInSessionDomain(email string, tokens vo.TokenPairVo) SignedInSessionDomain {
	return SignedInSessionDomain{email: email, tokens: tokens}
}

// Email is who this identity belongs to.
func (signedInSessionDomain SignedInSessionDomain) Email() string {
	return signedInSessionDomain.email
}

// AccessToken is the proof to carry on an ask.
func (signedInSessionDomain SignedInSessionDomain) AccessToken() string {
	return signedInSessionDomain.tokens.AccessToken
}

// RefreshToken is the half that buys a new pair. It is used once and never again.
func (signedInSessionDomain SignedInSessionDomain) RefreshToken() string {
	return signedInSessionDomain.tokens.RefreshToken
}

// IsAccessTokenUsable reports whether an ask can be made right now without renewing
// first.
func (signedInSessionDomain SignedInSessionDomain) IsAccessTokenUsable(now time.Time) bool {
	return now.Add(accessTokenExpiryMargin).Before(signedInSessionDomain.tokens.AccessTokenExpiresAt)
}

// IsRenewable reports whether there is still a way back from an expired signing-in.
//
// This is the line between "the user need not know" and "the user has to sign in
// again", and the two must never be told as the same thing.
func (signedInSessionDomain SignedInSessionDomain) IsRenewable(now time.Time) bool {
	return now.Before(signedInSessionDomain.tokens.RefreshTokenExpiresAt)
}

// WithRenewedTokens is this identity carrying a freshly granted pair. Both halves are
// replaced together, because they were granted together.
func (signedInSessionDomain SignedInSessionDomain) WithRenewedTokens(
	tokens vo.TokenPairVo,
) SignedInSessionDomain {
	signedInSessionDomain.tokens = tokens

	return signedInSessionDomain
}

// ToDto is this identity in the shape a caller is told about it — who, and until
// when. Never the proofs.
func (signedInSessionDomain SignedInSessionDomain) ToDto() dto.SignedInSessionDto {
	return dto.SignedInSessionDto{
		Email:                signedInSessionDomain.email,
		AccessTokenExpiresAt: signedInSessionDomain.tokens.AccessTokenExpiresAt,
	}
}
