package controller

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// bearerScheme is how a caller's own proof announces itself, lower-cased because the
// scheme is compared without regard to case.
const bearerScheme = "bearer "

// mcpCaller is who is on the other end of one call.
//
// Both answers it gives are read from the call's own envelope rather than from
// anything the caller wrote in the form — and that is the whole of why they live
// together here. A form that could name its own connection, or its own identity, is a
// form that could name somebody else's; keeping both readings in one place means
// there is one place to check that neither ever does.
type mcpCaller struct {
	request *mcp.CallToolRequest
}

func callerOn(request *mcp.CallToolRequest) mcpCaller {
	return mcpCaller{request: request}
}

// SessionKey is which connection this call came in on, and therefore whose identity
// this connector may use.
//
// There is deliberately no fallback for a call without one. A blank key is not
// "anonymous" — it is a drawer every such caller would share, which is the one
// failure this design exists to make impossible. A call always has a session here,
// and if that stops being true it should stop loudly.
func (mcpCaller mcpCaller) SessionKey() string {
	return mcpCaller.request.Session.ID()
}

// SuppliedAccessToken is a proof the caller brought along, or nothing.
//
// Anything that is not a bearer proof counts as nothing. Passing one on regardless
// would have the trading service reject it for a reason nobody could act on, when the
// truthful answer is simply that this caller has not said who they are.
func (mcpCaller mcpCaller) SuppliedAccessToken() string {
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
