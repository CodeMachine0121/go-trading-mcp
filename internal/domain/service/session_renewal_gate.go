package service

import (
	"sync"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// sessionRenewalGate makes sure one connection's renewal proof is spent once, however
// many asks hit the expiry at the same moment.
//
// This is not a performance measure. A renewal proof is single-use: the trading
// service treats a second use of a spent one as theft and answers by voiding the
// whole chain — which signs the real person out. Two asks renewing in parallel is
// therefore not "one wasted call", it is the user being logged out for doing two
// things at once.
//
// It holds a lock per connection rather than one lock for everybody, because two
// people renewing at the same time is ordinary and must not queue behind each other.
// It carries no business rule of its own: it only decides who waits.
type sessionRenewalGate struct {
	guard        sync.Mutex
	locksPerKeys map[string]*sync.Mutex
}

func newSessionRenewalGate() *sessionRenewalGate {
	return &sessionRenewalGate{locksPerKeys: map[string]*sync.Mutex{}}
}

// Enter hands back this connection's lock, already held. The caller releases it.
//
// Handing back a held lock rather than taking a callback keeps the caller's own
// control flow readable, and the caller is a single method that defers the release
// on its first line.
func (sessionRenewalGate *sessionRenewalGate) Enter(sessionKey vo.SessionKeyVo) *sync.Mutex {
	sessionRenewalGate.guard.Lock()

	lockForKey, isKnown := sessionRenewalGate.locksPerKeys[sessionKey.Value]
	if !isKnown {
		lockForKey = &sync.Mutex{}
		sessionRenewalGate.locksPerKeys[sessionKey.Value] = lockForKey
	}

	sessionRenewalGate.guard.Unlock()

	lockForKey.Lock()

	return lockForKey
}
