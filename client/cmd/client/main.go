package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/anarakinson/go_stonks/stonks_client/internal/user_handler"
	"github.com/anarakinson/go_stonks/stonks_client/pkg/interceptors"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
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
	// создаем соединение
	conn, err := grpc.NewClient(
		os.Getenv("TARGET_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),                                           // Ожидать подключения
		grpc.WithUnaryInterceptor(interceptors.XRequestIDClient()), // x-request-id interceptor
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

		// получаем от пользователя данные и создаем на их основе структуру заказа
		order, err := uHandler.GetOrder(userID)
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
			log.Printf("CreateOrder failed: %v", err)
		} else {
			fmt.Printf("Order created: %v", resp)
		}

		// -------------------------------------- //
		// получаем список заказов пользователя
		fmt.Println("\n***\n")
		respOrders, err := uHandler.GetUserOrders(order.UserID)
		if err != nil {
			log.Printf("GetUserOrders failed: %v", err)
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
