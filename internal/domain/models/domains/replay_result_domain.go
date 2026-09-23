package domains

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

const (
	// condensedEquityCurvePointLimit is how many equity curve points an assistant is
	// handed at most. Two hundred still draws the shape of a curve; fifty thousand
	// fill the room the assistant has for reading the report card.
	condensedEquityCurvePointLimit = 200
	// condensedClosedTradeLimit is how many of the most recent trades an assistant is
	// handed at most.
	condensedClosedTradeLimit = 100
)

// ReplayResultDomain is one replay result as the trading service answered it, and the
// one thing this connector does to it before an assistant reads it: make a long replay
// small enough to read.
//
// Only the two lists that grow with the length of the replay are touched — the equity
// curve is sampled evenly, first and last point always kept, and only the most recent
// trades are kept — and wherever either was shortened, the result says how long it
// was. The in-sample and validation parts are replay results of their own and are
// condensed by the same rules. **Nothing else is touched**: the report card is copied
// through as the trading service wrote it, figure for figure, because this connector
// explains and relays; it does not reinterpret.
//
// It is read as raw JSON fields rather than decoded into figures, which is what lets
// the report card's exact decimals pass through untouched.
type ReplayResultDomain struct {
	content string
}

func NewReplayResultDomain(content string) ReplayResultDomain {
	return ReplayResultDomain{content: content}
}

// ToCondensedContent is the replay result an assistant is handed. A result with
// nothing too long in it comes back exactly as it arrived, and so does anything that
// is not a replay result at all — condensing must never be the reason an answer does
// not reach the assistant.
func (replayResultDomain ReplayResultDomain) ToCondensedContent() string {
	fields := map[string]json.RawMessage{}
	if decodeError := json.Unmarshal([]byte(replayResultDomain.content), &fields); decodeError != nil {
		return replayResultDomain.content
	}

	// shortenedLists are the lists written back shorter, joined by hand below so that
	// every element is copied through exactly as the trading service wrote it.
	shortenedLists := map[string][]json.RawMessage{}

	equityPoints := []json.RawMessage{}
	if json.Unmarshal(fields["equityCurve"], &equityPoints) == nil &&
		len(equityPoints) > condensedEquityCurvePointLimit {
		sampledPoints := make([]json.RawMessage, 0, condensedEquityCurvePointLimit)
		lastIndex := len(equityPoints) - 1
		for sampleIndex := range condensedEquityCurvePointLimit {
			sampledPoints = append(sampledPoints,
				equityPoints[sampleIndex*lastIndex/(condensedEquityCurvePointLimit-1)])
		}

		shortenedLists["equityCurve"] = sampledPoints
		fields["equityCurvePointTotalCount"] = json.RawMessage(strconv.Itoa(len(equityPoints)))
	}

	closedTrades := []json.RawMessage{}
	if json.Unmarshal(fields["closedTrades"], &closedTrades) == nil &&
		len(closedTrades) > condensedClosedTradeLimit {
		shortenedLists["closedTrades"] = closedTrades[len(closedTrades)-condensedClosedTradeLimit:]
		fields["closedTradeTotalCount"] = json.RawMessage(strconv.Itoa(len(closedTrades)))
	}

	for listName, keptElements := range shortenedLists {
		joined := bytes.Buffer{}
		joined.WriteByte('[')
		for elementIndex, element := range keptElements {
			if elementIndex > 0 {
				joined.WriteByte(',')
			}
			joined.Write(element)
		}
		joined.WriteByte(']')
		fields[listName] = joined.Bytes()
	}
	isShortened := len(shortenedLists) > 0

	for _, partName := range []string{"inSample", "validation"} {
		part, hasPart := fields[partName]
		if !hasPart {
			continue
		}

		condensedPart := NewReplayResultDomain(string(part)).ToCondensedContent()
		if condensedPart != string(part) {
			fields[partName] = json.RawMessage(condensedPart)
			isShortened = true
		}
	}

	if !isShortened {
		return replayResultDomain.content
	}

	// Written back without HTML escaping, the way the trading service wrote it: an
	// escaped "<" or "&" in a strategy's name would reach the assistant as a literal
	// escape sequence.
	buffer := bytes.Buffer{}
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	// Every field is JSON the trading service wrote or a list joined from its own
	// elements, so there is nothing here that can fail to encode.
	_ = encoder.Encode(fields)

	return strings.TrimSuffix(buffer.String(), "\n")
}
