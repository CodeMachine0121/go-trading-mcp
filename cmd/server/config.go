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
	ServerPort                   string
	McpPath                      string
	TradingServiceBaseUrl        string
	TradingServiceRequestTimeout time.Duration
	LiveUpdateWaitLimit          time.Duration
}

func loadApplicationConfig() ApplicationConfig {
	return ApplicationConfig{
		ServerPort:            textWithDefault("SERVER_PORT", "8090"),
		McpPath:               textWithDefault("MCP_PATH", "/mcp"),
		TradingServiceBaseUrl: textWithDefault("TRADING_SERVICE_BASE_URL", "http://localhost:8080"),
		TradingServiceRequestTimeout: time.Duration(
			wholeNumberWithDefault("TRADING_SERVICE_REQUEST_TIMEOUT_SECONDS", 30)) * time.Second,
		LiveUpdateWaitLimit: time.Duration(
			wholeNumberWithDefault("LIVE_UPDATE_WAIT_LIMIT_SECONDS", 10)) * time.Second,
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
