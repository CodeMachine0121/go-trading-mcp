package main

import (
	"net/http"

	"github.com/CodeMachine0121/go-trading-mcp/internal/application"
	"github.com/CodeMachine0121/go-trading-mcp/internal/controller"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/service"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/clock"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/persistence"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/tradingservice"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// buildMcpServer wires everything together and puts every ability on one server.
//
// This is the only place that knows any concrete type. Everything above it works
// through interfaces, which is what lets the whole connector be tested without an HTTP
// server, a clock, or a trading service anywhere near it.
func buildMcpServer(applicationConfig ApplicationConfig) *mcp.Server {
	tradingServiceProxy := tradingservice.NewTradingServiceProxy(
		applicationConfig.TradingServiceBaseUrl, applicationConfig.TradingServiceRequestTimeout)
	signedInSessionRepository := persistence.NewSignedInSessionRepository()

	authenticationService := service.NewAuthenticationService(
		tradingServiceProxy, signedInSessionRepository, clock.NewClock())
	apiToolService := service.NewApiToolService(
		apiToolCatalog(applicationConfig.LiveUpdateWaitLimit),
		authenticationService,
		tradingServiceProxy,
	)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "go-trading-mcp",
		Version: "1.0.0",
		Title:   "交易助理外掛",
	}, &mcp.ServerOptions{
		Instructions: "這個外掛讓你代使用者操作交易服務：查行情、寫策略腳本、組交易策略、" +
			"重演一段行情、養策略機器人、設定通知。\n\n" +
			"交易服務內建的行情對話助手**刻意不在這裡**——代你去問另一個 AI 是在花它的錢，" +
			"而那個決定不該由你做。要用它請直接告訴使用者。\n\n" +
			"大部分能力需要身分。使用者還沒登入時，請先請他提供電子郵件與密碼並呼叫 trading_sign_in；" +
			"登入之後所有能力都會自動以他的身分進行，登入過期外掛會自己換新，不必再問他。\n\n" +
			"被拒絕時，回覆裡的是交易服務自己的說法——請照它說的改一改再試，不要重送一模一樣的東西。" +
			"回覆說「連不到交易服務」時則相反：那不是你送錯了，晚一點再試同一件事。",
	})

	controller.NewAuthenticationController(
		application.NewAuthenticationApplication(authenticationService)).RegisterOn(server)
	controller.NewApiToolController(
		application.NewApiToolApplication(apiToolService)).RegisterOn(server)

	return server
}

// buildHttpHandler puts the MCP server behind one address.
//
// Every connection gets the same server instance and its own session, which is what
// makes one person's identity unreachable from another's connection: the identity is
// filed under the session, and the session is the transport's own idea of who is on
// the other end.
func buildHttpHandler(applicationConfig ApplicationConfig, server *mcp.Server) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{
			// 沒有這一行，每一段連過的連線都會被留著，直到行程結束為止——
			// 一個只會長大、永遠不會縮小的表。閒置到期的代價只是重新登入一次。
			SessionTimeout: applicationConfig.IdleConnectionTimeout,
		})

	router := http.NewServeMux()
	router.Handle(applicationConfig.McpPath, mcpHandler)
	router.HandleFunc("/health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"Healthy"}`))
	})

	return router
}
