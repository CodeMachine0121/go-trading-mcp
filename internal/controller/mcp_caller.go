package controller

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const bearerScheme = "bearer "

type mcpCaller struct {
	request *mcp.CallToolRequest
}

func callerOn(request *mcp.CallToolRequest) mcpCaller {
	return mcpCaller{request: request}
}

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
