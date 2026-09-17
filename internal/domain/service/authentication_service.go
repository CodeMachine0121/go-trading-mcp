package service

import (
	"context"

	_interface "github.com/CodeMachine0121/go-trading-mcp/internal/domain/interface"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// AuthenticationService holds who each connection is, and keeps that usable.
//
// Keeping it usable is the whole of the difficulty. A signing-in lasts about fifteen
// minutes and a person works for hours, so expiry is the ordinary case, not the
// exceptional one — and the renewal that fixes it can only be spent once. Both of
// those live in here, so that everything else in this connector can simply ask for a
// proof and get one.
type AuthenticationService struct {
	tradingServiceProxy       _interface.ITradingServiceProxy
	signedInSessionRepository _interface.ISignedInSessionRepository
	clock                     _interface.IClock
	renewalGate               *sessionRenewalGate
}

func NewAuthenticationService(
	tradingServiceProxy _interface.ITradingServiceProxy,
	signedInSessionRepository _interface.ISignedInSessionRepository,
	clock _interface.IClock,
) *AuthenticationService {
	return &AuthenticationService{
		tradingServiceProxy:       tradingServiceProxy,
		signedInSessionRepository: signedInSessionRepository,
		clock:                     clock,
		renewalGate:               newSessionRenewalGate(),
	}
}

// SignIn hands an account over and keeps what comes back.
//
// The password travels no further than this call. What is kept is the pair of
// proofs; what is answered is who you now are and until when — never the proofs
// themselves, because an answer is something an assistant reads back and stores.
//
// Signing in again on a connection that already has an identity replaces it outright.
// There is no such thing as two identities on one connection: the next ask would have
// to choose between them, and nothing on that ask says which.
func (authenticationService *AuthenticationService) SignIn(
	ctx context.Context,
	sessionKey vo.SessionKeyVo,
	signInDto dto.SignInDto,
) (dto.SignInOutcomeDto, error) {
	sessionGrant, sendError := authenticationService.tradingServiceProxy.SignIn(ctx, signInDto)
	if sendError != nil {
		return dto.SignInOutcomeDto{}, sendError
	}

	if sessionGrant.Outcome != vo.TradingServiceSucceeded {
		return dto.SignInOutcomeDto{
			Outcome: dto.ToolOutcomeRefusedByTradingService,
			Content: sessionGrant.Content,
		}, nil
	}

	authenticationService.forgetWhatNobodyCanUse()

	signedInSession := domains.NewSignedInSessionDomain(signInDto.Email, sessionGrant.Tokens)
	authenticationService.signedInSessionRepository.Save(sessionKey, signedInSession)

	return dto.SignInOutcomeDto{
		Outcome: dto.ToolOutcomeSucceeded,
		Session: signedInSession.ToDto(),
	}, nil
}

// SignOut gives up this connection's identity.
//
// Signing out one that was never there is success: the aim — that this connection no
// longer acts as anybody — has been achieved, and saying otherwise would make the
// safe thing to do look like a mistake.
func (authenticationService *AuthenticationService) SignOut(
	ctx context.Context,
	sessionKey vo.SessionKeyVo,
) error {
	signedInSession, isSignedIn := authenticationService.signedInSessionRepository.Find(sessionKey)
	if !isSignedIn {
		return nil
	}

	authenticationService.signedInSessionRepository.Remove(sessionKey)

	_, sendError := authenticationService.tradingServiceProxy.RevokeSession(
		ctx, signedInSession.RefreshToken())

	return sendError
}

// UsableAccessToken hands back a proof that can be carried right now, renewing first
// if the one in hand has run out.
//
// This is the method the rest of the connector uses, and it is deliberately the only
// shape of the question: asking "is it expired?" and then "renew it" separately would
// let two asks answer the first question, both act on it, and spend the single-use
// renewal twice.
func (authenticationService *AuthenticationService) UsableAccessToken(
	ctx context.Context,
	sessionKey vo.SessionKeyVo,
) (string, error) {
	signedInSession, isSignedIn := authenticationService.signedInSessionRepository.Find(sessionKey)
	if !isSignedIn {
		return "", domains.ErrSignInRequired
	}

	if signedInSession.IsAccessTokenUsable(authenticationService.clock.Now()) {
		return signedInSession.AccessToken(), nil
	}

	return authenticationService.renewOnce(ctx, sessionKey, signedInSession.AccessToken())
}

// RenewedAccessToken replaces a proof that has just been rejected, whatever the
// expiry time said about it.
//
// It exists for the one case the arithmetic cannot see: the trading service itself
// says it does not recognise us, while our own sums said there were minutes left.
// Clocks disagree and a signing-in can be revoked from another device — either way its
// answer is the one that counts.
//
// The rejected proof is named rather than implied, so that this can tell "nobody has
// renewed yet" from "somebody already did while I was waiting". Passing none means
// renew regardless, which is what a person asking for a fresh signing-in outright
// means.
func (authenticationService *AuthenticationService) RenewedAccessToken(
	ctx context.Context,
	sessionKey vo.SessionKeyVo,
	rejectedAccessToken string,
) (string, error) {
	if _, isSignedIn := authenticationService.signedInSessionRepository.Find(sessionKey); !isSignedIn {
		return "", domains.ErrSignInRequired
	}

	return authenticationService.renewOnce(ctx, sessionKey, rejectedAccessToken)
}

// renewOnce is the renewal itself, behind this connection's gate.
//
// It is private and shared by both public ways in, which is what makes "spent once"
// true: both doors lead here, and this one lets one caller through at a time.
//
// It re-reads the identity **after** taking the gate, and compares what it finds with
// the proof the caller found wanting. A different one means somebody renewed while
// this caller waited — so this caller takes theirs and renews nothing. Renewing on top
// of that is exactly the double-spend the gate exists to prevent, and the trading
// service answers a double-spend by voiding the chain and signing the real person out.
func (authenticationService *AuthenticationService) renewOnce(
	ctx context.Context,
	sessionKey vo.SessionKeyVo,
	rejectedAccessToken string,
) (string, error) {
	lockForKey := authenticationService.renewalGate.Enter(sessionKey)
	defer lockForKey.Unlock()

	signedInSession, isSignedIn := authenticationService.signedInSessionRepository.Find(sessionKey)
	if !isSignedIn {
		return "", domains.ErrSignInRequired
	}

	if rejectedAccessToken != "" && signedInSession.AccessToken() != rejectedAccessToken {
		return signedInSession.AccessToken(), nil
	}

	if !signedInSession.IsRenewable(authenticationService.clock.Now()) {
		authenticationService.signedInSessionRepository.Remove(sessionKey)

		return "", domains.ErrSignInExpired
	}

	sessionGrant, sendError := authenticationService.tradingServiceProxy.RenewSession(
		ctx, signedInSession.RefreshToken())
	if sendError != nil {
		return "", sendError
	}

	if sessionGrant.Outcome != vo.TradingServiceSucceeded {
		authenticationService.signedInSessionRepository.Remove(sessionKey)

		return "", domains.ErrSignInExpired
	}

	renewedSession := signedInSession.WithRenewedTokens(sessionGrant.Tokens)
	authenticationService.signedInSessionRepository.Save(sessionKey, renewedSession)

	return renewedSession.AccessToken(), nil
}

// forgetWhatNobodyCanUse drops every identity that can no longer do anything.
//
// "Can no longer do anything" is not a new rule invented for tidiness — it is
// IsRenewable, the same line that already decides whether somebody has to sign in
// again. Once the renewal half is gone, the pair cannot be refreshed and the short
// half has long since lapsed: what is left is a credential that would be refused if
// it were used, and that nobody will ever come back for.
//
// It runs on signing in rather than on a timer. Signing in is the only moment the
// store grows, so it is the only moment the growth needs answering — and doing it
// here means no goroutine to start, stop, or leak in its own right.
func (authenticationService *AuthenticationService) forgetWhatNobodyCanUse() {
	now := authenticationService.clock.Now()

	authenticationService.signedInSessionRepository.RemoveUnusable(
		func(signedInSession domains.SignedInSessionDomain) bool {
			return !signedInSession.IsRenewable(now)
		})
}
