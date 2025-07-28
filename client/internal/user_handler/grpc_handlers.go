package user_handler

import (
	"context"
	"log"
	"time"

	"github.com/anarakinson/go_stonks/stonks_client/internal/domain"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
)

func (h *UserHandler) CreateOrderRequest(order *domain.Order) (*pb.CreateOrderResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("create order")
	resp, err := h.client.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId:    order.UserID,
		MarketId:  order.MarketID,
		OrderType: order.OrderType,
		Price:     order.Price,
		Quantity:  order.Quantity,
	})

	return resp, err

}

func (h *UserHandler) GetUserOrders(userId string) (*pb.GetUserOrdersResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetUserOrders(ctx, &pb.GetUserOrdersRequest{UserId: userId})
	return resp, err

}
