package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// shutdownDrainLimit is how long already-accepted requests get to finish once a stop
// is asked for.
//
// Long enough for an ordinary ask to the trading service, short enough that a stop is
// still a stop. An ability that watches has its own, shorter, wait limit, so nothing
// here waits on a line that could stay open indefinitely.
const shutdownDrainLimit = 15 * time.Second

func main() {
	_ = godotenv.Load()

	applicationConfig := loadApplicationConfig()
	httpServer := &http.Server{
		Addr:    applicationConfig.ServerBindAddress + ":" + applicationConfig.ServerPort,
		Handler: buildHttpHandler(applicationConfig),
	}

	stopSignalled, stopListening := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopListening()

	go func() {
		log.Printf("交易助理外掛啟動於 %s:%s%s（對外 %s），交易服務為 %s，外掛授權向 %s 取得",
			applicationConfig.ServerBindAddress, applicationConfig.ServerPort,
			applicationConfig.McpPath, applicationConfig.ProtectedResourceUrl(),
			applicationConfig.TradingServiceBaseUrl, applicationConfig.TradingServicePublicUrl)

		if serveError := httpServer.ListenAndServe(); !errors.Is(serveError, http.ErrServerClosed) {
			log.Fatalf("外掛停止服務：%v", serveError)
		}
	}()

	<-stopSignalled.Done()
	log.Println("收到關機訊號，等已經收下的請求回答完")

	drainCtx, stopDraining := context.WithTimeout(context.Background(), shutdownDrainLimit)
	defer stopDraining()

	if shutdownError := httpServer.Shutdown(drainCtx); shutdownError != nil {
		log.Printf("關機時還有請求沒排空：%v", shutdownError)
	}

	log.Println("已關機。")
}
