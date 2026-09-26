package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

type ConnectorAuthorizationController struct {
	connectorAuthorizationApplication *application.ConnectorAuthorizationApplication
	protectedResourceUrl              string
	resourceMetadataUrl               string
	authorizationServerUrl            string
}

func NewConnectorAuthorizationController(
	connectorAuthorizationApplication *application.ConnectorAuthorizationApplication,
	protectedResourceUrl string,
	resourceMetadataUrl string,
	authorizationServerUrl string,
) *ConnectorAuthorizationController {
	return &ConnectorAuthorizationController{
		connectorAuthorizationApplication: connectorAuthorizationApplication,
		protectedResourceUrl:              protectedResourceUrl,
		resourceMetadataUrl:               resourceMetadataUrl,
		authorizationServerUrl:            authorizationServerUrl,
	}
}

func (connectorAuthorizationController *ConnectorAuthorizationController) Guard(handler http.Handler) http.Handler {
	return auth.RequireBearerToken(
		connectorAuthorizationController.verify,
		&auth.RequireBearerTokenOptions{ResourceMetadataURL: connectorAuthorizationController.resourceMetadataUrl},
	)(handler)
}

func (connectorAuthorizationController *ConnectorAuthorizationController) MetadataHandler() http.Handler {
	return auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               connectorAuthorizationController.protectedResourceUrl,
		AuthorizationServers:   []string{connectorAuthorizationController.authorizationServerUrl},
		BearerMethodsSupported: []string{"header"},
	})
}

// Unreachable must not read as a rejection, or the person is sent to sign in again.
func (connectorAuthorizationController *ConnectorAuthorizationController) verify(
	ctx context.Context,
	accessToken string,
	_ *http.Request,
) (*auth.TokenInfo, error) {
	authorizationDto, verificationError := connectorAuthorizationController.connectorAuthorizationApplication.
		VerifyConnectorAuthorization(ctx, accessToken)
	if errors.Is(verificationError, domains.ErrConnectorAuthorizationRejected) {
		return nil, fmt.Errorf("%w：%s", auth.ErrInvalidToken, verificationError.Error())
	}

	if verificationError != nil {
		slog.Error("確認外掛授權失敗", slog.String("error", verificationError.Error()))

		return nil, domains.ErrTradingServiceUnreachable
	}

	return &auth.TokenInfo{UserID: authorizationDto.Subject, Expiration: authorizationDto.ExpiresAt}, nil
}
