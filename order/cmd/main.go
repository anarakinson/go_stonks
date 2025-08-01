package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/anarakinson/go_stonks/order/internal/repository/inmemory"
	"github.com/anarakinson/go_stonks/order/internal/server"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/logger"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/tracing"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/metrics"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

func main() {

	//--------------------------------------------//
	// загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "error", err)
		return
	}

	//--------------------------------------------//
	// инициализируем логгер
	if err := logger.Init("production"); err != nil {
		slog.Error("Unable to init zap-logger", "error", err)
		return
	}
	defer logger.Sync()

	//--------------------------------------------//
	// создаем и запускаем сервер сборки метрик
	go func() {
		if err := metrics.RunMetricsServer(); err != nil {
			logger.Log.Error("Metrics server error:", zap.Error(err))
		}
	}()

	//--------------------------------------------//
	// инициализация трейсинга jaegar
	tp, err := tracing.InitTracerProvider("jaeger:4317", "order-service", "1.0.0", "development", nil)
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}()

	//--------------------------------------------//
	// создаем хранилище
	repo := inmemory.NewRepository()

	//--------------------------------------------//
	// создаем и запускаем сервер
	serv := server.NewServer(os.Getenv("PORT"), repo)
	err = serv.Run()
	if err != nil {
		logger.Log.Error(
			"Failed on serve",
			zap.Error(err),
		)
		return
	}
}
