package dto

import "encoding/json"

type ToolCallDto struct {
	ToolName    string
	Arguments   map[string]json.RawMessage
	AccessToken string
}
