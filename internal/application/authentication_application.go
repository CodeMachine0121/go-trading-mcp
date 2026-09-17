package application

import (
	"context"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
)

// AuthenticationApplication is the three things a person does with their own account
// on this connector: hand it over, give it up, and force a fresh signing-in.
type AuthenticationApplication struct {
	authenticationService *service.AuthenticationService
}

func NewAuthenticationApplication(
	authenticationService *service.AuthenticationService,
) *AuthenticationApplication {
	return &AuthenticationApplication{authenticationService: authenticationService}
}

// SignIn hands an account over to this connection.
func (authenticationApplication *AuthenticationApplication) SignIn(
	ctx context.Context,
	sessionKey string,
	signInDto dto.SignInDto,
) (dto.SignInOutcomeDto, error) {
	return authenticationApplication.authenticationService.SignIn(
		ctx, vo.NewSessionKeyVo(sessionKey), signInDto)
}

// SignOut gives up this connection's identity.
func (authenticationApplication *AuthenticationApplication) SignOut(
	ctx context.Context,
	sessionKey string,
) error {
	return authenticationApplication.authenticationService.SignOut(
		ctx, vo.NewSessionKeyVo(sessionKey))
}

// RenewSession spends this connection's renewal now rather than waiting for the
// expiry to force it.
//
// Nothing needs this — every ask renews for itself when it has to. It exists because
// a person who has just changed something about their account somewhere else wants a
// way to say "start again from now" without signing in twice.
func (authenticationApplication *AuthenticationApplication) RenewSession(
	ctx context.Context,
	sessionKey string,
) error {
	_, renewalError := authenticationApplication.authenticationService.RenewedAccessToken(
		ctx, vo.NewSessionKeyVo(sessionKey), "")

	return renewalError
}
