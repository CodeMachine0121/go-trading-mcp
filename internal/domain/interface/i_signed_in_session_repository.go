package _interface

import (
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

//go:generate go tool mockgen -source=i_signed_in_session_repository.go -destination=mocks/mock_i_signed_in_session_repository.go -package=mocks

// ISignedInSessionRepository is where one connection's identity is kept while the
// connector is running.
//
// Keyed by connection and by nothing else. That is the whole of what stops one
// person's identity being handed to another, so there is deliberately no way to ask
// for "the" identity, or for one by email.
type ISignedInSessionRepository interface {
	Find(sessionKey vo.SessionKeyVo) (domains.SignedInSessionDomain, bool)
	Save(sessionKey vo.SessionKeyVo, signedInSession domains.SignedInSessionDomain)
	Remove(sessionKey vo.SessionKeyVo)
}
