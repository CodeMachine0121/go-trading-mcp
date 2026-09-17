package controller

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// signInRequest is what a caller sends to hand their account over.
//
// The password is a field here and is read exactly once. Nothing that outlives this
// call holds it — not the answer, not the identity that is kept, not the log.
type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthenticationController puts the three things a person does with their own account
// in front of the assistant.
//
// These three are the connector's own rather than the trading service's relayed. They
// look like the trading service's — signing in, renewing, signing out — but each does
// something extra that only this connector can do: keep the pair of proofs, keep them
// out of the answer, and tie them to this connection and no other.
type AuthenticationController struct {
	authenticationApplication *application.AuthenticationApplication
}

func NewAuthenticationController(
	authenticationApplication *application.AuthenticationApplication,
) *AuthenticationController {
	return &AuthenticationController{authenticationApplication: authenticationApplication}
}

// RegisterOn puts the three abilities on the given server.
func (authenticationController *AuthenticationController) RegisterOn(server *mcp.Server) {
	server.AddTool(&mcp.Tool{
		Name: "trading_sign_in",
		Description: "用電子郵件與密碼登入交易服務，把身分交給這個外掛保管。" +
			"登入之後，所有需要身分的能力都會自動以你的身分進行，不必每次重報。" +
			"登入過期時外掛會自己換新，你不會被打斷。" +
			"\n\n回覆只會說你是誰與這份登入到期於何時，**永遠不含任何憑證**；密碼用完即丟。" +
			"\n\n同一個連線再登入一次別的帳號，就換成那一位。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"email":    map[string]any{"type": "string", "description": "交易服務的帳號電子郵件"},
				"password": map[string]any{"type": "string", "description": "該帳號的密碼"},
			},
			"required": []string{"email", "password"},
		},
	}, authenticationController.signIn)

	server.AddTool(&mcp.Tool{
		Name: "trading_sign_out",
		Description: "登出：把整條換發鏈作廢，並清掉這個連線保管的身分。" +
			"本來就沒登入時也回成功——目的已經達成了。" +
			"\n\n注意：登入憑證本身撤不掉，所以最長還有十幾分鐘它仍然通得過。這是交易服務的設計，不是這裡的疏漏。",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	}, authenticationController.signOut)

	server.AddTool(&mcp.Tool{
		Name: "trading_renew_session",
		Description: "立刻換一份新的登入，不等它過期。" +
			"平常不需要用——每一件事都會在需要時自己換。" +
			"它是給「我剛在別處改了帳號設定，想從現在重新開始」用的。",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	}, authenticationController.renewSession)
}

func (authenticationController *AuthenticationController) signIn(
	ctx context.Context,
	request *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	var signInRequest signInRequest
	if decodeError := json.Unmarshal(request.Params.Arguments, &signInRequest); decodeError != nil {
		return refusal("送來的欄位不是一組可以讀的資料：" + decodeError.Error()), nil
	}

	outcomeDto, signInError := authenticationController.authenticationApplication.SignIn(
		ctx, sessionKeyOf(request), dto.SignInDto{
			Email:    signInRequest.Email,
			Password: signInRequest.Password,
		})
	if signInError != nil {
		return refusal(domains.ErrTradingServiceUnreachable.Error() + "：" + signInError.Error()), nil
	}

	if outcomeDto.Outcome != dto.ToolOutcomeSucceeded {
		return refusal(outcomeDto.Content), nil
	}

	return signedInAnswer(outcomeDto.Session), nil
}

func (authenticationController *AuthenticationController) signOut(
	ctx context.Context,
	request *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	signOutError := authenticationController.authenticationApplication.SignOut(
		ctx, sessionKeyOf(request))
	if signOutError != nil {
		return refusal(domains.ErrTradingServiceUnreachable.Error() + "：" + signOutError.Error()), nil
	}

	return answer("已登出。這個連線之後要身分的能力都會請你先登入。"), nil
}

func (authenticationController *AuthenticationController) renewSession(
	ctx context.Context,
	request *mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	renewalError := authenticationController.authenticationApplication.RenewSession(
		ctx, sessionKeyOf(request))

	switch {
	case renewalError == nil:
		return answer("已換到一份新的登入。"), nil
	case errors.Is(renewalError, domains.ErrSignInRequired):
		return refusal(domains.ErrSignInRequired.Error()), nil
	case errors.Is(renewalError, domains.ErrSignInExpired):
		return refusal(domains.ErrSignInExpired.Error()), nil
	default:
		return refusal(domains.ErrTradingServiceUnreachable.Error() + "：" + renewalError.Error()), nil
	}
}
