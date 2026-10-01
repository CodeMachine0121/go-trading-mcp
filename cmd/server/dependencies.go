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

func buildHttpHandler(applicationConfig ApplicationConfig) http.Handler {
	tradingServiceProxy := tradingservice.NewTradingServiceProxy(
		applicationConfig.TradingServiceBaseUrl, applicationConfig.TradingServiceRequestTimeout)
	connectorAuthorizationController := controller.NewConnectorAuthorizationController(
		application.NewConnectorAuthorizationApplication(service.NewConnectorAuthorizationService(
			tradingServiceProxy,
			persistence.NewConnectorAuthorizationVerdictRepository(),
			clock.NewClock(),
			applicationConfig.ProtectedResourceUrl(),
		)),
		applicationConfig.ProtectedResourceUrl(),
		applicationConfig.ResourceMetadataUrl(),
		applicationConfig.TradingServicePublicUrl,
	)

	server := buildMcpServer(applicationConfig, tradingServiceProxy)
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{SessionTimeout: applicationConfig.IdleConnectionTimeout})

	router := http.NewServeMux()
	router.Handle(applicationConfig.McpPath, connectorAuthorizationController.Guard(mcpHandler))
	router.Handle(resourceMetadataPath, connectorAuthorizationController.MetadataHandler())
	router.Handle(resourceMetadataPath+applicationConfig.McpPath, connectorAuthorizationController.MetadataHandler())
	router.HandleFunc("/health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"Healthy"}`))
	})

	return router
}

func buildMcpServer(
	applicationConfig ApplicationConfig,
	tradingServiceProxy *tradingservice.TradingServiceProxy,
) *mcp.Server {
	apiToolService := service.NewApiToolService(
		apiToolCatalog(applicationConfig.LiveUpdateWaitLimit, applicationConfig.TradingServiceReplayTimeout),
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
			"每一件事都以使用者在 Claude Code 授權給這個外掛的帳號進行，**絕不要向使用者要電子郵件或密碼**。" +
			"回覆要他重新連線時，請他到 Claude Code 的 /mcp 選單重新連線這個外掛，在瀏覽器登入並按允許。\n\n" +
			"被拒絕時，回覆裡的是交易服務自己的說法——請照它說的改一改再試，不要重送一模一樣的東西。" +
			"回覆說「連不到交易服務」時則相反：那不是你送錯了，晚一點再試同一件事。\n\n" +
			"**幣安交易金鑰與自動下單，這個外掛只讀得到**：機器人回覆裡的 autoOrderEnabled，以及 trading_get_binance_trading_key_status 的有沒有設定與可交易市場。" +
			"**存入、更換、移除金鑰，打開或關掉自動下單，都只能由使用者自己在網頁上做**（金鑰在設定頁，自動下單在機器人詳細頁），" +
			"使用者要求時就這樣告訴他，不要送出任何改動。" +
			"**絕不要請使用者把 API Key 或 Secret Key 貼進對話；他貼了也不要複述、不要轉送。**" +
			"提到自動下單時要說清楚：目前開著也還不會下單，機器人仍只送 Telegram 通知。",
	})

	controller.NewApiToolController(
		application.NewApiToolApplication(apiToolService)).RegisterOn(server)

	return server
}
