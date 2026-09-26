package vo

import "time"

// ConnectorAuthorizationInspectionVo is the trading service's own judgement of one
// connector authorization, before this connector decides whether it counts here.
type ConnectorAuthorizationInspectionVo struct {
	IsActive  bool
	Subject   string
	Audience  string
	ExpiresAt time.Time
}
