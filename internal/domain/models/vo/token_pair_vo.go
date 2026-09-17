package vo

import "time"

// TokenPairVo is one signing-in's worth of proof: the short-lived one carried on
// every ask, and the long-lived one that only ever buys a new pair.
//
// They travel together because they are issued together and replaced together. A
// renewal that kept one and replaced the other would leave a pair whose two halves
// came from different signings-in, and the trading service voids a whole renewal
// chain the moment it sees the older half twice.
type TokenPairVo struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

// NewTokenPairVo takes one pair exactly as the trading service issued it.
func NewTokenPairVo(
	accessToken string,
	accessTokenExpiresAt time.Time,
	refreshToken string,
	refreshTokenExpiresAt time.Time,
) TokenPairVo {
	return TokenPairVo{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}
}
