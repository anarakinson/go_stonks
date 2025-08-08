package order_service

import (
	"context"
	"sync"
	"time"

	"github.com/anarakinson/go_stonks/order/internal/domain"
	market_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/market"
	order_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"go.uber.org/zap"

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
	mu                   sync.RWMutex
	updatesChannels      map[string]chan *order_pb.Order
}

func NewService(spotClient spot_inst_pb.SpotInstrumentServiceClient, repo Repository) *Service {
	return &Service{
		orders:               repo,
		spotInstrumentClient: spotClient,
		updatesChannels:      make(map[string]chan *order_pb.Order),
	}
}

// GetMarkets - получает список доступных рынков у Spot service и возвращает клиенту
func (s *Service) GetMarkets(ctx context.Context, req *order_pb.GetMarketsRequest) (*order_pb.GetMarketsResponse, error) {
	logger.Log.Info("received msg", zap.String("user role", req.UserRoles.String()))

	// проверяем, существует ли рынок и доступен ли
	marketsResp, err := s.spotInstrumentClient.ViewMarkets(ctx, &spot_inst_pb.ViewMarketsRequest{UserRoles: req.UserRoles})
	if err != nil {
		logger.Log.Error("Failed to check markets", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to check market service: %v", err)
	}

	// возвращаем клиенту только маркеты, который подходят по роли
	markets := []*market_pb.Market{}
	for _, m := range marketsResp.Markets {
		for _, role := range m.AllowedRoles {
			if role == req.UserRoles {
				markets = append(markets, m)
				break
			}
		}
	}

	// возвращаем список существующих маркетов (id)
	return &order_pb.GetMarketsResponse{Markets: markets}, nil
}

// GetOrderStatus - возвращает статус заказа
func (s *Service) GetOrderStatus(ctx context.Context, req *order_pb.GetOrderStatusRequest) (*order_pb.GetOrderStatusResponse, error) {

	// получаем данные из хранилища
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, exists := s.orders.GetOrder(req.OrderId)

	// проверяем, ордер существует и принадлежит запрашивающему пользователю
	if !exists || order.UserID != req.UserId {
		logger.Log.Info("Order does not found")
		return nil, status.Errorf(codes.NotFound, "order not found")
	}
	// возвращаем респонс
	return &order_pb.GetOrderStatusResponse{Order: OrderToProto(order)}, nil
}

// CreateOrder - создает заказ и помещает в хранилище
func (s *Service) CreateOrder(ctx context.Context, req *order_pb.CreateOrderRequest) (*order_pb.CreateOrderResponse, error) {

	// проверяем, существует ли рынок и доступен ли
	marketsResp, err := s.spotInstrumentClient.ViewMarkets(ctx, &spot_inst_pb.ViewMarketsRequest{UserRoles: req.UserRoles})
	if err != nil {
		logger.Log.Error("Failed to check markets", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to check markets: %v", err)
	}

	// проходим по всем существующим маркетам и сверяем ID
	var marketExists bool
	for _, market := range marketsResp.Markets {
		if market.Id == req.MarketId {
			// проверяем, есть ли роль юзера в списке доступа маркета
			for _, role := range market.AllowedRoles {
				if role == req.UserRoles {
					// помечаем маркет, как существующий и прерываемся
					marketExists = true
					break
				}
			}
		}
	}

	// если не найден - значит маркет недоступен или не существует
	if !marketExists {
		logger.Log.Info("Market does not exists")
		return nil, status.Errorf(codes.NotFound, "market not found or disabled")
	}

	// формируем ордер
	order := domain.NewOrder(req.UserId, req.MarketId, req.OrderType, req.Price, req.Quantity)
	// Создаём канал для обновлений заказа
	updatesCh := make(chan *order_pb.Order)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updatesChannels[order.ID] = updatesCh
	// отправляем ордер на обработку
	go s.processOrder(order.ID)

	err = s.orders.AddOrder(order)
	if err != nil {
		logger.Log.Error("Error adding order to repository", zap.Error(err))
		return nil, status.Errorf(codes.AlreadyExists, "order already exists")
	}

	return &order_pb.CreateOrderResponse{Status: order.Status, OrderId: order.ID}, nil

}

// GetOrderStatus - возвращает статус заказа
func (s *Service) GetUserOrders(ctx context.Context, req *order_pb.GetUserOrdersRequest) (*order_pb.GetUserOrdersResponse, error) {

	// получаем данные из хранилища
	s.mu.RLock()
	defer s.mu.RUnlock()
	orders, err := s.orders.GetUserOrders(req.UserId)
	// проверяем
	if err != nil {
		logger.Log.Error("Error getting order from repository", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to check orders: %v", err)
	}

	// преобразуем заказы в прото формат
	var protoOrders []*order_pb.Order
	for _, o := range orders {
		protoOrders = append(protoOrders, OrderToProto(o))
	}

	// возвращаем респонс
	return &order_pb.GetUserOrdersResponse{Orders: protoOrders}, nil
}

// стрим состояния заказа
func (s *Service) StreamOrderUpdates(req *order_pb.GetOrderStatusRequest, stream order_pb.OrderService_StreamOrderUpdatesServer) error {
	orderID := req.GetOrderId()

	s.mu.Lock()
	updatesCh, exists := s.updatesChannels[orderID]
	s.mu.Unlock()

	if !exists {
		logger.Log.Error("Required order does not exists", zap.String("order ID", orderID))
		return status.Errorf(
			codes.NotFound,
			"order with ID '%s' does not exist",
			orderID,
		)
	}

	// Читаем из канала и отправляем клиенту
	for order := range updatesCh {
		if err := stream.Send(&order_pb.GetOrderStatusResponse{Order: order}); err != nil {
			logger.Log.Error("Stream error", zap.Error(err))
			return err
		}
	}

	return nil
}

// вспомогательная функция, имитирующая бурную деятельность по обработке поступившего заказа
func (s *Service) processOrder(id string) {
	time.Sleep(10 * time.Second)
	s.mu.Lock()
	defer s.mu.Unlock()
	defer delete(s.updatesChannels, id) // удаляем канал из мапы
	defer close(s.updatesChannels[id])  // закрываем канал
	// получаем ордер из базы данных
	order, ok := s.orders.GetOrder(id)
	if !ok {
		logger.Log.Error("error updating order status")
		return
	}
	// обновляем статус
	order.Status = "done"
	// отправляем ордер в канал обработки
	s.updatesChannels[id] <- OrderToProto(order)
}
