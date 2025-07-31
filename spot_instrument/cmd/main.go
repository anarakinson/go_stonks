package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/repository/inmemory"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/server"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/logger"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/tracing"
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
	// инициализация трейсинга jaegar
	tp, err := tracing.InitTracerProvider("order-service")
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}()

	//--------------------------------------------//
	// создаем хранилище с моками рынков
	repo := inmemory.NewRepository()
	repo.AddMarket(domain.NewMarket("BTC-USDT", true))
	repo.AddMarket(domain.NewMarket("BTC-USDC", true))
	repo.AddMarket(domain.NewMarket("ETH-USDT", false))
	repo.AddMarket(domain.NewMarket("ETH-USDC", true))
	solMarket := domain.NewMarket("SOL/USDT", true)
	solMarket.Delete()
	repo.AddMarket(solMarket)

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
