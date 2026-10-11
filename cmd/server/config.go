package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const resourceMetadataPath = "/.well-known/oauth-protected-resource"

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
	PublicBaseUrl                string
	TradingServiceBaseUrl        string
	TradingServicePublicUrl      string
	TradingServiceRequestTimeout time.Duration
	// TradingServiceReplayTimeout covers the on-the-spot comparisons and a run looked at with a wait; it must outlast the trading service's own limits.
	TradingServiceReplayTimeout time.Duration
	LiveUpdateWaitLimit         time.Duration
}

func loadApplicationConfig() ApplicationConfig {
	return ApplicationConfig{
		ServerBindAddress:     textWithDefault("SERVER_BIND_ADDRESS", "127.0.0.1"),
		ServerPort:            textWithDefault("SERVER_PORT", "8090"),
		McpPath:               textWithDefault("MCP_PATH", "/mcp"),
		PublicBaseUrl:         strings.TrimSuffix(textWithDefault("PUBLIC_BASE_URL", "http://localhost:8090"), "/"),
		TradingServiceBaseUrl: textWithDefault("TRADING_SERVICE_BASE_URL", "http://localhost:8080"),
		TradingServicePublicUrl: strings.TrimSuffix(
			textWithDefault("TRADING_SERVICE_PUBLIC_URL", "http://localhost:8080"), "/"),
		TradingServiceRequestTimeout: time.Duration(
			wholeNumberWithDefault("TRADING_SERVICE_REQUEST_TIMEOUT_SECONDS", 30)) * time.Second,
		TradingServiceReplayTimeout: time.Duration(
			wholeNumberWithDefault("TRADING_SERVICE_REPLAY_TIMEOUT_SECONDS", 120)) * time.Second,
		LiveUpdateWaitLimit: time.Duration(
			wholeNumberWithDefault("LIVE_UPDATE_WAIT_LIMIT_SECONDS", 10)) * time.Second,
	}
}

// Configured rather than derived: Traefik rewrites the scheme.
func (applicationConfig ApplicationConfig) ProtectedResourceUrl() string {
	return applicationConfig.PublicBaseUrl + applicationConfig.McpPath
}

func (applicationConfig ApplicationConfig) ResourceMetadataUrl() string {
	return applicationConfig.PublicBaseUrl + resourceMetadataPath + applicationConfig.McpPath
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
