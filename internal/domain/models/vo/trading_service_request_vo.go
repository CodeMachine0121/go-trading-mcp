package vo

import "time"

// TradingServiceRequestVo is one ask, ready to hand to the trading service.
//
// It is the whole of what leaves the domain. Everything the assistant filled in has
// already been placed — into the address, onto the end of it, or inside it — so
// nothing downstream has to know which box went where, and nothing downstream can
// place one differently.
type TradingServiceRequestVo struct {
	Verb  RequestVerb
	Path  string
	Query map[string]string
	// Body is the ask's contents, or empty when this ability has nothing to put
	// inside. Empty means *no* contents, not an empty set of them: some of the
	// trading service's rules read "left out" and "given as nothing" differently.
	Body []byte
	// CarriesIdentity says whether this ask must say who is making it.
	CarriesIdentity bool
	// LiveUpdateWaitLimit is how long to stay on the line for an ability that
	// watches rather than asks. Zero — the ordinary case — means do not stay at all.
	LiveUpdateWaitLimit time.Duration
}
