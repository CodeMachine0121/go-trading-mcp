package domains

import "errors"

var ErrConnectorAuthorizationRejected = errors.New("外掛授權無效，或不是發給這個外掛的")

var ErrReconnectRequired = errors.New(
	"交易服務不認得這份外掛授權。請到 Claude Code 的 /mcp 選單重新連線這個外掛（重新授權），再試一次")
