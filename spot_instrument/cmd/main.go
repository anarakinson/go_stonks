package main

import (
	"log"
	"log/slog"
	"os"
	"github.com/anarakinson/go_stonks/spot_instrument_service/internal/domain"
	"github.com/anarakinson/go_stonks/spot_instrument_service/internal/repository/inmemory"
	"github.com/anarakinson/go_stonks/spot_instrument_service/internal/server"

	"github.com/joho/godotenv"
)

func main() {

	//--------------------------------------------//
	// загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file: %v", err)
		return
	}

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
		log.Fatalf("failed on serve: %v", err)
	}
}
