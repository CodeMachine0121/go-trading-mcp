package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
)

type rememberedVerdict struct {
	verdict         domains.ConnectorAuthorizationDomain
	rememberedUntil time.Time
}

// ConnectorAuthorizationVerdictRepository keeps judgements in memory under the
// token's SHA-256 fingerprint, so no connector authorization itself is ever held.
type ConnectorAuthorizationVerdictRepository struct {
	guard                     sync.Mutex
	rememberedPerFingerprints map[string]rememberedVerdict
}

func NewConnectorAuthorizationVerdictRepository() *ConnectorAuthorizationVerdictRepository {
	return &ConnectorAuthorizationVerdictRepository{
		rememberedPerFingerprints: map[string]rememberedVerdict{},
	}
}

func (connectorAuthorizationVerdictRepository *ConnectorAuthorizationVerdictRepository) Find(
	accessToken string,
	now time.Time,
) (domains.ConnectorAuthorizationDomain, bool) {
	connectorAuthorizationVerdictRepository.guard.Lock()
	defer connectorAuthorizationVerdictRepository.guard.Unlock()

	remembered, isRemembered := connectorAuthorizationVerdictRepository.
		rememberedPerFingerprints[connectorAuthorizationVerdictRepository.fingerprintOf(accessToken)]
	if !isRemembered || !now.Before(remembered.rememberedUntil) {
		return domains.ConnectorAuthorizationDomain{}, false
	}

	return remembered.verdict, true
}

// Save also forgets every judgement already past its time, so the store never only grows.
func (connectorAuthorizationVerdictRepository *ConnectorAuthorizationVerdictRepository) Save(
	accessToken string,
	verdict domains.ConnectorAuthorizationDomain,
	now time.Time,
) {
	connectorAuthorizationVerdictRepository.guard.Lock()
	defer connectorAuthorizationVerdictRepository.guard.Unlock()

	for fingerprint, remembered := range connectorAuthorizationVerdictRepository.rememberedPerFingerprints {
		if !now.Before(remembered.rememberedUntil) {
			delete(connectorAuthorizationVerdictRepository.rememberedPerFingerprints, fingerprint)
		}
	}

	connectorAuthorizationVerdictRepository.
		rememberedPerFingerprints[connectorAuthorizationVerdictRepository.fingerprintOf(accessToken)] =
		rememberedVerdict{verdict: verdict, rememberedUntil: verdict.RememberedUntil(now)}
}

func (connectorAuthorizationVerdictRepository *ConnectorAuthorizationVerdictRepository) fingerprintOf(
	accessToken string,
) string {
	fingerprint := sha256.Sum256([]byte(accessToken))

	return hex.EncodeToString(fingerprint[:])
}
