package vo

// SessionGrantVo is what came back from asking the trading service for a signing-in
// or a renewal: either a pair of proofs, or its own words about why not.
//
// It carries the same three-way verdict as any other answer rather than a pair-or-
// error, because "wrong password" and "no signing keys configured" are both the
// trading service speaking, and both have to reach the person in its words.
type SessionGrantVo struct {
	Outcome TradingServiceOutcome
	Content string
	Tokens  TokenPairVo
}
