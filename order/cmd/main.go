package main

import (
	"log/slog"
	"os"

	"github.com/anarakinson/go_stonks/order/internal/repository/inmemory"
	"github.com/anarakinson/go_stonks/order/internal/server"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/logger"
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
	shutdown := tracer.initTracing("client-service")
	defer shutdown() // закрытие при завершении

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
