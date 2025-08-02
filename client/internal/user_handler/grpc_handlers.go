package user_handler

import (
	"context"
	"time"

	"github.com/anarakinson/go_stonks/stonks_client/internal/domain"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"go.uber.org/zap"

	pb_market "github.com/anarakinson/go_stonks/stonks_pb/gen/market"
	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
)

func (h *UserHandler) CreateOrderRequest(order *domain.Order) (*pb.CreateOrderResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log.Info("Create order")
	resp, err := h.client.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId:    order.UserID,
		MarketId:  order.MarketID,
		OrderType: order.OrderType,
		Price:     order.Price,
		Quantity:  order.Quantity,
	})
	if err != nil {
		logger.Log.Error(
			"Error CreateOrder",
			zap.Error(err),
		)
		return nil, err
	}

	return resp, nil

}

func (h *UserHandler) GetUserOrders(userId string) (*pb.GetUserOrdersResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log.Info("Get user orders")
	resp, err := h.client.GetUserOrders(ctx, &pb.GetUserOrdersRequest{UserId: userId})

	if err != nil {
		logger.Log.Error(
			"Error GetUserOrders",
			zap.Error(err),
		)
		return nil, err
	}

	return resp, nil

}

func (h *UserHandler) GetMarkets() ([]*pb_market.Market, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log.Info("Get available markets")
	markets, err := h.client.GetMarkets(ctx, &pb.GetMarketsRequest{})

	if err != nil {
		logger.Log.Error(
			"Error GetMarkets",
			zap.Error(err),
		)
		return nil, err
	}

	return markets.Markets, nil
}
