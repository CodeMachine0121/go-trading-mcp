package dto

import "encoding/json"

// ToolCallDto is one request to have an ability carried out.
//
// The filled-in boxes arrive as the raw JSON they came in as. The application layer
// does not read them and has no business shape for them — it is the domain that
// wraps them into something with behaviour.
//
// It carries two ways of saying who is asking, and they are not interchangeable.
// SessionKey names an identity this connector is holding and may renew on the
// caller's behalf. SuppliedAccessToken is one the caller brought along, which this
// connector uses as given and never renews — the renewal half is not in its hands.
type ToolCallDto struct {
	ToolName            string
	Arguments           map[string]json.RawMessage
	SessionKey          string
	SuppliedAccessToken string
}
