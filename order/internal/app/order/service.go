package order_service

import (
	"context"

	"github.com/anarakinson/go_stonks/order_service/internal/domain"
	order_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Repository interface {
	AddOrder(*domain.Order) error
	GetOrder(string) (*domain.Order, bool)
	GetUserOrders(UserId string) ([]*domain.Order, error)
}

type Service struct {
	order_pb.UnimplementedOrderServiceServer
	orders               Repository
	spotInstrumentClient spot_inst_pb.SpotInstrumentServiceClient
}

func NewService(spotClient spot_inst_pb.SpotInstrumentServiceClient, repo Repository) *Service {
	return &Service{
		orders:               repo,
		spotInstrumentClient: spotClient,
	}
}

// GetMarkets - получает список доступных рынков у Spot service и возвращает клиенту
func (s *Service) GetMarkets(ctx context.Context, req *order_pb.GetMarketsRequest) (*order_pb.GetMarketsResponse, error) {

	logger.Log.Info("order service - GetMarkets request")

	// проверяем, существует ли рынок и доступен ли
	marketsResp, err := s.spotInstrumentClient.ViewMarkets(ctx, &spot_inst_pb.ViewMarketsRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check markets: %v", err)
	}

	// возвращаем список существующих маркетов (id)
	logger.Log.Info("order service - GetMarkets response")
	return &order_pb.GetMarketsResponse{Markets: marketsResp.Markets}, nil
}

// GetOrderStatus - возвращает статус заказа
func (s *Service) GetOrderStatus(ctx context.Context, req *order_pb.GetOrderStatusRequest) (*order_pb.GetOrderStatusResponse, error) {

	logger.Log.Info("order service - GetOrderStatus request")

	// получаем данные из хранилища
	order, exists := s.orders.GetOrder(req.OrderId)

	// проверяем, ордер существует и принадлежит запрашивающему пользователю
	if !exists || order.UserID != req.UserId {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}
	// возвращаем респонс
	logger.Log.Info("order service - GetOrderStatus response")
	return &order_pb.GetOrderStatusResponse{Order: OrderToProto(order)}, nil
}

// CreateOrder - создает заказ и помещает в хранилище
func (s *Service) CreateOrder(ctx context.Context, req *order_pb.CreateOrderRequest) (*order_pb.CreateOrderResponse, error) {

	logger.Log.Info("order service - CreateOrder request")

	// проверяем, существует ли рынок и доступен ли
	marketsResp, err := s.spotInstrumentClient.ViewMarkets(ctx, &spot_inst_pb.ViewMarketsRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check markets: %v", err)
	}

	// проходим по всем существующим маркетам и сверяем ID
	var marketExists bool
	for _, market := range marketsResp.Markets {
		if market.Id == req.MarketId {
			marketExists = true
			break
		}
	}

	// если не найден - значит маркет недоступен или не существует
	if !marketExists {
		return nil, status.Errorf(codes.NotFound, "market not found or disabled")
	}
	order := domain.NewOrder(req.UserId, req.MarketId, req.OrderType, req.Price, req.Quantity)

	err = s.orders.AddOrder(order)
	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "order already exists")
	}

	logger.Log.Info("order service - CreateOrder response")
	return &order_pb.CreateOrderResponse{Status: order.Status, OrderId: order.ID}, nil

}

// GetOrderStatus - возвращает статус заказа
func (s *Service) GetUserOrders(ctx context.Context, req *order_pb.GetUserOrdersRequest) (*order_pb.GetUserOrdersResponse, error) {

	logger.Log.Info("order service - GetUserOrders request")

	// получаем данные из хранилища
	orders, err := s.orders.GetUserOrders(req.UserId)
	// проверяем
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check orders: %v", err)
	}

	// преобразуем заказы в прото формат
	var protoOrders []*order_pb.Order
	for _, o := range orders {
		protoOrders = append(protoOrders, OrderToProto(o))
	}

	// возвращаем респонс
	logger.Log.Info("order service - GetUserOrders response")
	return &order_pb.GetUserOrdersResponse{Orders: protoOrders}, nil
}
