package domains

import "errors"

// ErrTradingServiceUnreachable is what a trading service that did not answer at all
// gets said about it.
//
// It is kept apart from every refusal on purpose. A refusal means change what you
// sent; this means what you sent never arrived, so sending something different will
// not help and sending the same thing later will. Told as one sentence, an assistant
// spends its retries rewriting a request that was fine.
var ErrTradingServiceUnreachable = errors.New("連不到交易服務")
