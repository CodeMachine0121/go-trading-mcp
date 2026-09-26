package dto

import "time"

type ConnectorAuthorizationDto struct {
	Subject   string
	ExpiresAt time.Time
}
