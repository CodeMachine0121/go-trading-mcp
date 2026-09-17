package persistence

import (
	"sync"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// SignedInSessionRepository keeps each connection's identity for as long as this
// connector is running, and no longer.
//
// In memory, deliberately. Writing proofs to disk would make them outlive the process
// that earned them — recoverable from a backup, readable by anything that can read
// the file, and still valid. The price is that a restart signs everybody out, which
// costs one password and is the cheaper of the two mistakes.
//
// The lock is a read-write one because reading is what almost every ask does and
// writing happens twice a session: once on signing in, once on renewing.
type SignedInSessionRepository struct {
	guard                  sync.RWMutex
	sessionsPerSessionKeys map[string]domains.SignedInSessionDomain
}

func NewSignedInSessionRepository() *SignedInSessionRepository {
	return &SignedInSessionRepository{
		sessionsPerSessionKeys: map[string]domains.SignedInSessionDomain{},
	}
}

func (signedInSessionRepository *SignedInSessionRepository) Find(
	sessionKey vo.SessionKeyVo,
) (domains.SignedInSessionDomain, bool) {
	signedInSessionRepository.guard.RLock()
	defer signedInSessionRepository.guard.RUnlock()

	signedInSession, isSignedIn := signedInSessionRepository.sessionsPerSessionKeys[sessionKey.Value]

	return signedInSession, isSignedIn
}

func (signedInSessionRepository *SignedInSessionRepository) Save(
	sessionKey vo.SessionKeyVo,
	signedInSession domains.SignedInSessionDomain,
) {
	signedInSessionRepository.guard.Lock()
	defer signedInSessionRepository.guard.Unlock()

	signedInSessionRepository.sessionsPerSessionKeys[sessionKey.Value] = signedInSession
}

func (signedInSessionRepository *SignedInSessionRepository) Remove(sessionKey vo.SessionKeyVo) {
	signedInSessionRepository.guard.Lock()
	defer signedInSessionRepository.guard.Unlock()

	delete(signedInSessionRepository.sessionsPerSessionKeys, sessionKey.Value)
}
