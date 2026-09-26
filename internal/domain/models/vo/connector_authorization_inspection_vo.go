package vo

import "time"

type ConnectorAuthorizationInspectionVo struct {
	IsActive  bool
	Subject   string
	Audience  string
	ExpiresAt time.Time
}
