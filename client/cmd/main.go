package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/anarakinson/go_stonks/stonks_client/internal/client"
	"go.uber.org/zap"

	"github.com/anarakinson/go_stonks_shared/pkg/grpc_helpers"
	"github.com/anarakinson/go_stonks_shared/pkg/interceptors"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/anarakinson/go_stonks_shared/pkg/tracing"

	"github.com/joho/godotenv"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
)

func main() {

	//--------------------------------------------//
	// Канал для graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error, 1)

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
		logger.Log.Error("Failed to init tracer", zap.Error(err))
		return
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Log.Error("Failed to shutdown tracer", zap.Error(err))
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
		interceptors.RetryInterceptor(3) // retry-интерсептор на три попытки
	)
	if err != nil {
		logger.Log.Error("Connection failed", zap.Error(err))
		return
	}
	defer conn.Close()

	gclient := pb.NewOrderServiceClient(conn)

	// создбаем обработчик для взаимодействия с клиентом
	cl := client.NewClient(gclient)

	// создаем контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// запускаем обработчик
	go func() {
		err = cl.HandleUserInput(ctx)
		if err != nil {
			logger.Log.Error("Error handling user input", zap.Error(err))
		}
		errChan <- err
	}()

	// грейсфул шатдаун
	// Ждем либо сигнал завершения, либо ошибку сервера
	select {
	case err := <-errChan:
		logger.Log.Error("Client error", zap.Error(err))
	case <-shutdown:
		logger.Log.Info("Server is shutting down...")
	}

}
