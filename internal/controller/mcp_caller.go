package controller

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const bearerScheme = "bearer "

// mcpCaller is who is on the other end of one call, read only from the call's own
// envelope and never from the form the assistant filled in.
type mcpCaller struct {
	request *mcp.CallToolRequest
}

func callerOn(request *mcp.CallToolRequest) mcpCaller {
	return mcpCaller{request: request}
}

// AccessToken is the connector authorization this call arrived under, forwarded to the
// trading service as is.
func (mcpCaller mcpCaller) AccessToken() string {
	extra := mcpCaller.request.GetExtra()
	if extra == nil || extra.Header == nil {
		return ""
	}

	authorization := extra.Header.Get("Authorization")
	if len(authorization) <= len(bearerScheme) ||
		!strings.EqualFold(authorization[:len(bearerScheme)], bearerScheme) {
		return ""
	}

	return strings.TrimSpace(authorization[len(bearerScheme):])
}
