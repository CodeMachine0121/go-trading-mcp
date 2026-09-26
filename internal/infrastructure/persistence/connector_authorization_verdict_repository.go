package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
)

// Sweeping only past this size keeps Save constant-time under the lock in ordinary use.
const sweepThreshold = 10_000

type rememberedVerdict struct {
	verdict         domains.ConnectorAuthorizationDomain
	rememberedUntil time.Time
}

// Keyed by SHA-256 so the token itself is never held.
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

func (connectorAuthorizationVerdictRepository *ConnectorAuthorizationVerdictRepository) Save(
	accessToken string,
	verdict domains.ConnectorAuthorizationDomain,
	now time.Time,
) {
	connectorAuthorizationVerdictRepository.guard.Lock()
	defer connectorAuthorizationVerdictRepository.guard.Unlock()

	if len(connectorAuthorizationVerdictRepository.rememberedPerFingerprints) >= sweepThreshold {
		for fingerprint, remembered := range connectorAuthorizationVerdictRepository.rememberedPerFingerprints {
			if !now.Before(remembered.rememberedUntil) {
				delete(connectorAuthorizationVerdictRepository.rememberedPerFingerprints, fingerprint)
			}
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
