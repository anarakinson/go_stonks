package main

import (
	"log"
	"os"

	"log/slog"

	"order_service/internal/repository/inmemory"
	"order_service/internal/server"

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
	// создаем хранилище
	repo := inmemory.NewRepository()

	//--------------------------------------------//
	// создаем и запускаем сервер
	serv := server.NewServer(os.Getenv("PORT"), repo)
	err = serv.Run()
	if err != nil {
		log.Fatalf("failed on serve: %v", err)
	}
}
