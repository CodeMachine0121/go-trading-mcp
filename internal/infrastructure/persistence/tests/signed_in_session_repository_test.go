package persistence_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sessionFor(email string, accessToken string) domains.SignedInSessionDomain {
	return domains.NewSignedInSessionDomain(email, vo.NewTokenPairVo(
		accessToken, time.Now().Add(time.Minute), accessToken+"-refresh", time.Now().Add(time.Hour)))
}

func TestAnIdentityIsFoundOnlyUnderTheConnectionItWasFiledUnder(t *testing.T) {
	repository := persistence.NewSignedInSessionRepository()
	repository.Save(vo.NewSessionKeyVo("connection-a"), sessionFor("james@example.com", "a-token"))

	found, isSignedIn := repository.Find(vo.NewSessionKeyVo("connection-a"))
	_, strangerIsSignedIn := repository.Find(vo.NewSessionKeyVo("connection-b"))

	require.True(t, isSignedIn)
	assert.Equal(t, "james@example.com", found.Email())
	assert.False(t, strangerIsSignedIn, "別的連線讀不到這一份，那是它唯一的用處")
}

func TestFilingASecondIdentityUnderOneConnectionReplacesTheFirst(t *testing.T) {
	repository := persistence.NewSignedInSessionRepository()
	repository.Save(vo.NewSessionKeyVo("connection-a"), sessionFor("james@example.com", "first"))
	repository.Save(vo.NewSessionKeyVo("connection-a"), sessionFor("somebody@example.com", "second"))

	found, _ := repository.Find(vo.NewSessionKeyVo("connection-a"))

	assert.Equal(t, "somebody@example.com", found.Email())
	assert.Equal(t, "second", found.AccessToken())
}

func TestRemovingLeavesTheConnectionAsNobody(t *testing.T) {
	repository := persistence.NewSignedInSessionRepository()
	repository.Save(vo.NewSessionKeyVo("connection-a"), sessionFor("james@example.com", "a-token"))

	repository.Remove(vo.NewSessionKeyVo("connection-a"))
	_, isSignedIn := repository.Find(vo.NewSessionKeyVo("connection-a"))

	assert.False(t, isSignedIn)
}

func TestRemovingAConnectionThatWasNeverThereIsHarmless(t *testing.T) {
	repository := persistence.NewSignedInSessionRepository()

	assert.NotPanics(t, func() { repository.Remove(vo.NewSessionKeyVo("never-seen")) })
}

func TestManyConnectionsComingAndGoingAtOnceDoNotTreadOnEachOther(t *testing.T) {
	repository := persistence.NewSignedInSessionRepository()

	var everybodyDone sync.WaitGroup
	for index := range 50 {
		everybodyDone.Add(1)

		go func() {
			defer everybodyDone.Done()

			sessionKey := vo.NewSessionKeyVo(fmt.Sprintf("connection-%d", index))
			repository.Save(sessionKey, sessionFor("person@example.com", "token"))
			repository.Find(sessionKey)
			repository.Remove(sessionKey)
		}()
	}
	everybodyDone.Wait()

	_, isSignedIn := repository.Find(vo.NewSessionKeyVo("connection-0"))
	assert.False(t, isSignedIn)
}
