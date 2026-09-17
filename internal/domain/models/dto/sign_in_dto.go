package dto

// SignInDto is one attempt to hand an account over to this connector.
//
// The password is here and nowhere else: it is read once, exchanged for a pair of
// proofs, and never written down. Nothing that outlives this value holds it.
type SignInDto struct {
	Email    string
	Password string
}
