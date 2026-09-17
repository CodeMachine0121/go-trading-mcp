package controller

import (
	"fmt"
	"strings"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// bearerScheme is how a caller's own proof announces itself, lower-cased because the
// scheme is compared without regard to case.
const bearerScheme = "bearer "

// sessionKeyOf is which connection this call came in on.
//
// It is the transport's own idea of the connection, so one caller can never be handed
// another's identity by asking nicely: nothing in the call's contents influences it.
func sessionKeyOf(request *mcp.CallToolRequest) string {
	if request.Session == nil {
		return ""
	}

	return request.Session.ID()
}

// suppliedAccessTokenOf is a proof the caller brought along, if they brought one.
//
// Read from the request's own headers rather than from a box on the form, for the
// same reason: a form that could name its own identity is a form that could name
// somebody else's.
func suppliedAccessTokenOf(request *mcp.CallToolRequest) string {
	extra := request.GetExtra()
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

// answer is what the assistant is shown when the thing was done.
func answer(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

// refusal is what it is shown when it was not, marked as such so the assistant reads
// it as something to act on rather than as the answer it asked for.
func refusal(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
		IsError: true,
	}
}

// signedInAnswer is what signing in says: who you now are and until when, and
// deliberately not a single character of either proof.
func signedInAnswer(signedInSessionDto dto.SignedInSessionDto) *mcp.CallToolResult {
	return answer(fmt.Sprintf(
		"已登入：%s。這份登入到 %s 為止；過期時外掛會自己換新，你不必重登。",
		signedInSessionDto.Email,
		signedInSessionDto.AccessTokenExpiresAt.Format("2006-01-02 15:04:05 MST"),
	))
}
