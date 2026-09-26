package dto

import "time"

// ConnectorAuthorizationDto is a connector authorization that counts here: whose it
// is, and until when.
type ConnectorAuthorizationDto struct {
	Subject   string
	ExpiresAt time.Time
}
