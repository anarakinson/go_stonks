package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/anarakinson/go_stonks/stonks_client/internal/client"

	"github.com/anarakinson/go_stonks_shared/pkg/grpc_helpers"
	"github.com/anarakinson/go_stonks_shared/pkg/interceptors"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/anarakinson/go_stonks_shared/pkg/tracing"

	"github.com/joho/godotenv"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
)

func main() {
	//--------------------------------------------//
	// загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file:", "error", err)
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
	jaegarAddr := fmt.Sprintf("%s:%s", os.Getenv("JAEGER_HOST"), os.Getenv("JAEGER_PORT"))
	tp, err := tracing.InitTracerProvider(jaegarAddr, "client-service", "1.0.0", "development", nil)
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}()

	//--------------------------------------------//
	// создаем соединение
	target_address := os.Getenv("TARGET_ADDR")
	fmt.Println(target_address)

	// Без TLS (для тестов)
	conn, err := grpc_helpers.NewGRPCClient(
		target_address,
		nil, // TLS настройки
		// интерсепторы
		interceptors.XRequestIDClient(), // x-request-id interceptor
	)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	gclient := pb.NewOrderServiceClient(conn)

	cl := client.NewClient(gclient)

	err = cl.HandleUserInput()
	if err != nil {
		slog.Error("Error handling user input", "error", err)
	}

}
