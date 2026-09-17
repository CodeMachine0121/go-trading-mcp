package vo

// RequestVerb is what a request asks the trading service to do.
//
// It is deliberately not an HTTP method. The domain of this connector is "relay one
// ask to the trading service", and a domain that spelled its asks GET and DELETE
// would be a domain that knows how it travels — which is the proxy's business, not
// the domain's. The proxy owns the one table that turns these four into methods.
type RequestVerb string

const (
	// RequestVerbRead asks for something without changing it.
	RequestVerbRead RequestVerb = "read"
	// RequestVerbSubmit hands something over: a new record, or an action to perform.
	RequestVerbSubmit RequestVerb = "submit"
	// RequestVerbReplace rewrites something that already exists.
	RequestVerbReplace RequestVerb = "replace"
	// RequestVerbRemove takes something away.
	RequestVerbRemove RequestVerb = "remove"
)
