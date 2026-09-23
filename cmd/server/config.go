package main

import (
	"os"
	"strconv"
	"time"
)

// ApplicationConfig is everything about this connector that differs between one
// machine and another.
//
// Every setting has a default, so the connector runs with no `.env` at all. The one
// that matters is the trading service's address: point it at the wrong place and
// nothing here works, which is why it is first.
type ApplicationConfig struct {
	ServerBindAddress            string
	ServerPort                   string
	McpPath                      string
	TradingServiceBaseUrl        string
	TradingServiceRequestTimeout time.Duration
	// TradingServiceReplayTimeout is how long the replay abilities wait for their
	// answer. It is longer than the usual wait, and longer than the trading service's
	// own allowance for a whole replay, so that the trading service is the one to say
	// a replay ran out of time.
	TradingServiceReplayTimeout time.Duration
	LiveUpdateWaitLimit         time.Duration
	IdleConnectionTimeout       time.Duration
}

func loadApplicationConfig() ApplicationConfig {
	return ApplicationConfig{
		// 預設只聽 loopback。這個外掛的 MCP 端點**沒有任何門鎖**——連得到的人就開得了
		// 一段連線並以自己的帳號登入。聽在每一張網卡上，等於在咖啡廳的 wifi 上開一扇
		// 沒有鎖的門，而且沒有任何跡象顯示它開著。要對外開放請是個明確的動作。
		ServerBindAddress:     textWithDefault("SERVER_BIND_ADDRESS", "127.0.0.1"),
		ServerPort:            textWithDefault("SERVER_PORT", "8090"),
		McpPath:               textWithDefault("MCP_PATH", "/mcp"),
		TradingServiceBaseUrl: textWithDefault("TRADING_SERVICE_BASE_URL", "http://localhost:8080"),
		TradingServiceRequestTimeout: time.Duration(
			wholeNumberWithDefault("TRADING_SERVICE_REQUEST_TIMEOUT_SECONDS", 30)) * time.Second,
		TradingServiceReplayTimeout: time.Duration(
			wholeNumberWithDefault("TRADING_SERVICE_REPLAY_TIMEOUT_SECONDS", 120)) * time.Second,
		LiveUpdateWaitLimit: time.Duration(
			wholeNumberWithDefault("LIVE_UPDATE_WAIT_LIMIT_SECONDS", 10)) * time.Second,
		IdleConnectionTimeout: time.Duration(
			wholeNumberWithDefault("IDLE_CONNECTION_TIMEOUT_MINUTES", 60)) * time.Minute,
	}
}

func textWithDefault(name string, fallback string) string {
	declared := os.Getenv(name)
	if declared == "" {
		return fallback
	}

	return declared
}

// wholeNumberWithDefault reads a number, and treats an unreadable or unusable one as
// not having been said at all.
//
// Falling back rather than refusing to start is the right trade for a personal tool:
// a typo in one timing should not take the whole connector down, and every one of
// these has a default that works.
func wholeNumberWithDefault(name string, fallback int) int {
	declared := os.Getenv(name)
	if declared == "" {
		return fallback
	}

	wholeNumber, parseError := strconv.Atoi(declared)
	if parseError != nil || wholeNumber <= 0 {
		return fallback
	}

	return wholeNumber
}
