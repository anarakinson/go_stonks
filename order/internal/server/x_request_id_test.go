package server_test

import (
	"context"
	"net"
	"testing"

	order_service "github.com/anarakinson/go_stonks/order/internal/app/order"
	"github.com/anarakinson/go_stonks/order/internal/mocks"
	market_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/market/v1"
	order_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order/v1"
	"github.com/anarakinson/go_stonks_shared/pkg/interceptors"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

func TestOrderService_XRequestIDPassedCorrectly(t *testing.T) {
	// Инициализируем логгер
	if err := logger.Init("test"); err != nil {
		t.Fatalf("Unable to init zap-logger: %v", err)
	}
	defer logger.Sync()

	// Инициализируем моки
	repo := mocks.NewMockRepository()
	mockSpot := mocks.NewMockSpotInstrumentService()

	// Создаем тестируемый сервис
	service := order_service.NewService(mockSpot, repo)

	// Переменная для хранения полученного x-request-id
	var receivedRequestID string

	// Перехватчик для записи x-request-id
	testInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Извлекаем метаданные
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ids := md.Get("x-request-id")
			if len(ids) > 0 {
				receivedRequestID = ids[0]
			}
		}
		// Продолжаем обработку запроса
		return handler(ctx, req)
	}

	// Настраиваем gRPC сервер с нашим перехватчиком
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.XRequestIDServer(),
			testInterceptor, // наш кастомный перехватчик
		),
	)
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
		grpc.WithUnaryInterceptor(interceptors.XRequestIDClient()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := order_pb.NewOrderServiceClient(conn)

	// Вызываем метод с x-request-id
	testRequestID := "test-request-123"
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-request-id", testRequestID))
	_, err = client.GetMarkets(ctx, &order_pb.GetMarketsRequest{UserRoles: market_pb.UserRole_ROLE_BASIC})
	require.NoError(t, err)

	// Проверяем, что сервер получил правильный x-request-id
	assert.Equal(t, testRequestID, receivedRequestID, "x-request-id from server and from client are not equal")
}
