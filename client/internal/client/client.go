package client

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/anarakinson/go_stonks/stonks_client/internal/user_handler"
	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
)

type Client struct {
	gclient pb.OrderServiceClient
}

func NewClient(client pb.OrderServiceClient) *Client {
	return &Client{
		gclient: client,
	}
}

func (c *Client) HandleUserInput() error {
	// ----------------------------------------- //
	// начинаем взаимодействие с сервисом
	// создаем хендлер пользовательского ввода на основе клиента
	uHandler := user_handler.NewUserHandler(c.gclient)

	// Запрашиваем UserId
	userID, err := uHandler.GetUserID()
	if err != nil {
		if errors.Is(err, user_handler.ErrFinish) {
			fmt.Println("End")
			return err
		}
		log.Fatalf("Stdin error: %v", err)
	}
	if userID == "exit" {
		fmt.Println("End")
		return fmt.Errorf("user interruption")
	}

	// Запрашиваем User Role
	userRole, err := uHandler.GetUserRole()
	if err != nil {
		if errors.Is(err, user_handler.ErrFinish) {
			fmt.Println("End")
			return fmt.Errorf("user interruption")
		}
		log.Fatalf("Stdin error: %v", err)
	}

	// переходим в бесконечный цикл. получаем данные - отправляем запрос на сервис
	for {

		// получаем маркеты от внешнего сервиса
		markets, err := uHandler.GetMarkets(userRole)
		if err != nil {
			fmt.Println("Error get available markets.")
			return err
		}

		// получаем от пользователя данные и создаем на их основе структуру заказа
		order, err := uHandler.GetOrder(userID, markets)
		if err != nil {
			if errors.Is(err, user_handler.ErrFinish) {
				fmt.Println("End")
				return fmt.Errorf("user interruption")
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
		// Подписываемся на обновления по созданному заказу
		fmt.Println("\n\nWaiting for order processing done")
		stream, err := c.gclient.StreamOrderUpdates(context.Background(), &pb.GetOrderStatusRequest{
			UserId:  userID,
			OrderId: resp.OrderId, // указываем ордер из ответа
		})
		if err != nil {
			log.Fatalf("StreamOrderUpdates failed: %v", err)
		}

		// ждем, когда заказ обработается
		for {
			update, err := stream.Recv()
			if err != nil {
				log.Printf("Stream closed: %v", err)
				break
			}

			fmt.Printf("\nOrder id: %s \nOrder type: %s, \nStatus: %s\n",
				update.Order.Id,
				update.Order.OrderType,
				update.Order.Status,
			)
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
