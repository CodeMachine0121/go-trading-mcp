package tradingservice

import (
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// sessionTokensWire is a pair of proofs exactly as the trading service words them.
//
// It exists so the trading service's spelling stops here. The domain's pair names the
// two expiries symmetrically; this one does not, because the trading service calls the
// first one simply "expiresAt". Letting that asymmetry inwards would mean every reader
// of a signing-in has to remember which of the two the plain name refers to.
type sessionTokensWire struct {
	AccessToken           string    `json:"accessToken"`
	ExpiresAt             time.Time `json:"expiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

// ToTokenPairVo is this pair in the shape the domain holds it.
func (sessionTokensWire sessionTokensWire) ToTokenPairVo() vo.TokenPairVo {
	return vo.NewTokenPairVo(
		sessionTokensWire.AccessToken,
		sessionTokensWire.ExpiresAt,
		sessionTokensWire.RefreshToken,
		sessionTokensWire.RefreshTokenExpiresAt,
	)
}
