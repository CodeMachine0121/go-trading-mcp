package dto

// SignInOutcomeDto is how one attempt to hand an account over ended.
//
// It carries the same vocabulary of outcomes as any other attempt, because the
// caller does the same thing with it: show the answer, or show the trading service's
// own words about why there isn't one. Session is filled in only when it succeeded,
// and it never contains a proof.
type SignInOutcomeDto struct {
	Outcome ToolOutcome
	Content string
	Session SignedInSessionDto
}
