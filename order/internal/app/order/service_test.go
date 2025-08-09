package order_service_test

import (
	"context"
	"log/slog"
	"net"
	"testing"

	order_service "github.com/anarakinson/go_stonks/order/internal/app/order"
	"github.com/anarakinson/go_stonks/order/internal/mocks"
	market_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/market/v1"
	order_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order/v1"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestService_CreateOrder_InvalidMarket(t *testing.T) {
	// Инициализируем логгер
	if err := logger.Init("test"); err != nil {
		slog.Error("Unable to init zap-logger", "error", err)
		return
	}
	defer logger.Sync()

	// Инициализируем моки
	repo := mocks.NewMockRepository()
	mockSpot := mocks.NewMockSpotInstrumentService()

	// Создаем тестируемый сервис
	service := order_service.NewService(mockSpot, repo)

	// Настраиваем gRPC сервер
	server := grpc.NewServer()
	order_pb.RegisterOrderServiceServer(server, service)

	// Запускаем in-memory сервер
	listener := bufconn.Listen(1024 * 1024)
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	// Создаем клиент
	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := order_pb.NewOrderServiceClient(conn)

	// Тестовые случаи

	// Проверка добавления заказа по несуществующему маркету
	tests := []struct {
		name        string
		marketId    string
		expectedErr string
	}{
		{
			name:        "Non-existent market",
			marketId:    "999",
			expectedErr: "market not found or disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &order_pb.CreateOrderRequest{
				UserId:    "user1",
				MarketId:  tt.marketId,
				OrderType: "limit",
				Price:     100.0,
				Quantity:  1.0,
				UserRoles: market_pb.UserRole_ROLE_BASIC,
			}

			resp, err := client.CreateOrder(context.Background(), req)
			// проверяем, что метод вернул ошибку
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
			assert.Nil(t, resp)

			// Проверяем что заказ не добавился
			orders, err := repo.GetUserOrders("user1")

			assert.NoError(t, err)
			assert.Empty(t, orders)
		})
	}

	// Проверка успешного создания заказа
	t.Run("Valid active market", func(t *testing.T) {
		req := &order_pb.CreateOrderRequest{
			UserId:    "user1",
			MarketId:  "1", // активный рынок
			OrderType: "limit",
			Price:     100.0,
			Quantity:  1.0,
			UserRoles: market_pb.UserRole_ROLE_BASIC,
		}

		resp, err := client.CreateOrder(context.Background(), req)
		// проверяем, что метод не вернул ошибку, а ответ содержит айди и статус
		if assert.NoError(t, err) {
			assert.NotNil(t, resp)
			assert.NotEmpty(t, resp.OrderId)
			assert.Equal(t, "created", resp.Status)

			// Проверяем что заказ добавился
			orders, err := repo.GetUserOrders("user1")
			assert.NoError(t, err)
			if assert.Len(t, orders, 1) {
				assert.Equal(t, resp.OrderId, orders[0].ID)
			}
		}
	})
}
