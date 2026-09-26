package dto

import "encoding/json"

// ToolCallDto is one request to have an ability carried out, under the connector
// authorization the caller brought along.
type ToolCallDto struct {
	ToolName    string
	Arguments   map[string]json.RawMessage
	AccessToken string
}
