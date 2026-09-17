package dto

import "time"

// SignedInSessionDto is what a caller is told after signing in: who they now are,
// and until when.
//
// It carries no proofs, and that absence is the design. The proofs are this
// connector's to hold; putting them in an answer would print them into a transcript
// an assistant reads back, quotes, and keeps.
type SignedInSessionDto struct {
	Email                string
	AccessTokenExpiresAt time.Time
}
