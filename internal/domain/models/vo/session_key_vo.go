package vo

// SessionKeyVo is where one connection's identity is kept.
//
// One connection, one identity. It is a value rather than a bare string so that
// nothing can pass a trading symbol, a user's email or an access token where a
// connection was meant — the three mistakes that would each quietly hand one
// person's identity to somebody else.
type SessionKeyVo struct {
	Value string
}

// NewSessionKeyVo names one connection.
func NewSessionKeyVo(value string) SessionKeyVo {
	return SessionKeyVo{Value: value}
}
