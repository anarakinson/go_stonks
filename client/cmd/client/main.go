package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/anarakinson/go_stonks/stonks_client/internal/user_handler"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/interceptors"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/logger"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/tracing"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	// создаем соединение
	target_address := os.Getenv("TARGET_ADDR")
	fmt.Println(target_address)

	conn, err := grpc.NewClient(
		target_address,
		// настройки соединения (без шифрования)
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// поддержка соединения
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second, // проверка соединения каждые 10с
			Timeout: 5 * time.Second,  // разрыв при зависании на 5с
		}),
		// балансировщик нагрузки
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		// Параметры подключения
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,       // минимальное время попытки подключения
			Backoff:           backoff.DefaultConfig, // экспоненциальная задержка между попытками
		}),
		// OpenTelemetry трассировщик
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		// добавляем интерсепторы
		grpc.WithChainUnaryInterceptor(
			interceptors.XRequestIDClient(), // x-request-id interceptor
		),
	)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	// ----------------------------------------- //
	// начинаем взаимодействие с сервисом
	// создаем хендлер пользовательского ввода на основе клиента
	uHandler := user_handler.NewUserHandler(client)

	// Запрашиваем UserId
	userID, err := uHandler.GetUserID()
	if err != nil {
		if errors.Is(err, user_handler.ErrFinish) {
			fmt.Println("End")
			return
		}
		log.Fatalf("Stdin error: %v", err)
	}
	if userID == "exit" {
		fmt.Println("End")
		return
	}

	// переходим в бесконечный цикл. получаем данные - отправляем запрос на сервис
	for {

		// получаем маркеты от внешнего сервиса
		markets, err := uHandler.GetMarkets()
		if err != nil {
			fmt.Println("Error get available markets.")
			return
		}

		// получаем от пользователя данные и создаем на их основе структуру заказа
		order, err := uHandler.GetOrder(userID, markets)
		if err != nil {
			if errors.Is(err, user_handler.ErrFinish) {
				fmt.Println("End")
				return
			}
			log.Fatalf("Stdin error: %v", err)
		}

		// -------------------------------------- //
		// отправляем запрос к сервису

		// создаем заказ на основе введенных данных
		resp, err := uHandler.CreateOrderRequest(order)
		if err != nil {
			fmt.Println("Error creating order. Try again")
			continue
		} else {
			fmt.Printf("Order created: %v", resp)
		}

		// -------------------------------------- //
		// получаем список заказов пользователя
		fmt.Println("\n***\n")
		respOrders, err := uHandler.GetUserOrders(order.UserID)
		if err != nil {
			fmt.Println("Error getting user orders")
			continue
		}
		for _, o := range respOrders.Orders {
			fmt.Println("User orders:")
			fmt.Println(o)
		}

		// переходим на следующую итерацию цикла
		fmt.Println("\n***\n")
	}

}
