package domains

import "encoding/json"

// ToolArgumentsDomain is what the assistant filled in this time, and everything
// about how those pieces of JSON read once they leave JSON.
//
// The values are kept raw. This connector never decides what a value *means* — that
// belongs to the trading service — so reading them into Go types would only create a
// second opinion that can one day disagree with the first. What it does know is how
// a value reads in an address and how a handful of them read as contents, and both
// of those are about the values themselves, which is why they live here.
type ToolArgumentsDomain struct {
	values map[string]json.RawMessage
}

// NewToolArgumentsDomain takes the filled-in boxes, always as a set rather than
// sometimes nothing: filling nothing in is an empty set, not an absence.
func NewToolArgumentsDomain(values map[string]json.RawMessage) ToolArgumentsDomain {
	if values == nil {
		return ToolArgumentsDomain{values: map[string]json.RawMessage{}}
	}

	return ToolArgumentsDomain{values: values}
}

// Has reports whether this box was filled in at all.
func (toolArgumentsDomain ToolArgumentsDomain) Has(name string) bool {
	_, isFilledIn := toolArgumentsDomain.values[name]

	return isFilledIn
}

// PlainTextAt is how one box reads once it is out of JSON and into an address.
//
// A string loses its quotation marks; everything else is already written the way it
// belongs there. Leaving the marks on would send `"BTCUSDT"` as a trading symbol,
// and the trading service would — correctly — not find it.
func (toolArgumentsDomain ToolArgumentsDomain) PlainTextAt(name string) string {
	rawValue, isFilledIn := toolArgumentsDomain.values[name]
	if !isFilledIn {
		return ""
	}

	var text string
	if json.Unmarshal(rawValue, &text) == nil {
		return text
	}

	return string(rawValue)
}

// EncodedSubset is the named boxes gathered back into one piece of JSON, ready to be
// the contents of an ask.
//
// Names that were not filled in are left out rather than sent as nothing: the
// trading service reads "left out" and "given as nothing" differently in places, so
// filling the gaps here would be answering questions nobody asked. For the same
// reason, nothing to gather is nothing at all, not an empty set.
func (toolArgumentsDomain ToolArgumentsDomain) EncodedSubset(names []string) []byte {
	subset := make(map[string]json.RawMessage, len(names))
	for _, name := range names {
		if rawValue, isFilledIn := toolArgumentsDomain.values[name]; isFilledIn {
			subset[name] = rawValue
		}
	}

	if len(subset) == 0 {
		return nil
	}

	encodedSubset, encodeError := json.Marshal(subset)
	if encodeError != nil {
		return nil
	}

	return encodedSubset
}
