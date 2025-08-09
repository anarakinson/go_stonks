package user_handler

import (
	"context"
	"time"

	"github.com/anarakinson/go_stonks/stonks_client/internal/domain"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"go.uber.org/zap"

	pb_market "github.com/anarakinson/go_stonks/stonks_pb/gen/market/v1"
	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order/v1"
)

func (h *UserHandler) CreateOrderRequest(ctx context.Context, order *domain.Order, timeout time.Duration) (*pb.CreateOrderResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Проверка, не отменён ли уже исходный контекст
	if err := ctx.Err(); err != nil {
		logger.Log.Error(
			"Context cancelled before request",
			zap.Error(err),
		)
		return nil, err
	}

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

func (h *UserHandler) GetUserOrders(ctx context.Context, userId string, timeout time.Duration) (*pb.GetUserOrdersResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

    // Проверка, не отменён ли уже исходный контекст
    if err := ctx.Err(); err != nil {
        logger.Log.Error(
            "Context cancelled before request",
            zap.Error(err),
        )
        return nil, err
    }

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

func (h *UserHandler) GetMarkets(ctx context.Context, userRole pb_market.UserRole, timeout time.Duration) ([]*pb_market.Market, error) {

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Проверка, не отменён ли уже исходный контекст
    if err := ctx.Err(); err != nil {
        logger.Log.Error(
            "Context cancelled before request",
            zap.Error(err),
        )
        return nil, err
    }

	logger.Log.Info("Get available markets")
	markets, err := h.client.GetMarkets(ctx, &pb.GetMarketsRequest{UserRoles: userRole})

	if err != nil {
		logger.Log.Error(
			"Error GetMarkets",
			zap.Error(err),
		)
		return nil, err
	}

	return markets.Markets, nil
}
