package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// ConnectorAuthorizationController is this connector's face as an OAuth protected
// resource: the metadata that points Claude Code at the authorization server, and the
// guard that turns away every call without a connector authorization that counts here.
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

// verify leaves an unreachable trading service as a plain error, so it is not answered
// as a rejected authorization that would send the person through signing in again.
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
		return nil, verificationError
	}

	return &auth.TokenInfo{UserID: authorizationDto.Subject, Expiration: authorizationDto.ExpiresAt}, nil
}
