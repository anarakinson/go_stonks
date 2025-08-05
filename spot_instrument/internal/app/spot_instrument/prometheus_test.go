package spot_instrument_test

import (
	"context"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/anarakinson/go_stonks/spot_instrument/internal/app/spot_instrument"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/mocks"
	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestViewMarkets_PrometheusMetrics(t *testing.T) {
	// Инициализация логгера
	if err := logger.Init("test"); err != nil {
		slog.Error("Unable to init zap-logger", "error", err)
		return
	}
	defer logger.Sync()

	// 1. Подготовка мока репозитория
	repo := mocks.NewMockRepository(
		[]*domain.Market{
			{ID: "1", Name: "BTC/USD", Enabled: true, DeletedAt: nil},
		},
	)

	// 2. Создаем тестовый сервис
	service := spot_instrument.NewService(repo)

	// 3. Настройка Prometheus
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	registry := prometheus.NewRegistry()
	registry.MustRegister(grpcMetrics)

	// 4. Создаем тестовый gRPC сервер
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
	)
	spot_inst_pb.RegisterSpotInstrumentServiceServer(server, service)

	// 5. Запускаем тестовый сервер
	lis, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer server.Stop()

	// 6. Создаем клиент
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := spot_inst_pb.NewSpotInstrumentServiceClient(conn)

	// 7. Получаем начальное значение счетчика
	metricName := "grpc_server_handled_total"
	before := getCounterValue(registry, metricName, "ViewMarkets")

	// 8. Вызываем метод через gRPC клиент
	_, err = client.ViewMarkets(context.Background(), &spot_inst_pb.ViewMarketsRequest{})
	assert.NoError(t, err)

	// 9. Даем время для обработки метрик
	time.Sleep(100 * time.Millisecond)

	// 10. Проверяем что счетчик увеличился
	after := getCounterValue(registry, metricName, "ViewMarkets")
	assert.Equal(t, before+1, after, "Счетчик вызовов должен увеличиться на 1")
}

// Вспомогательная функция для получения значения счетчика
func getCounterValue(registry *prometheus.Registry, metricName, method string) int {
	metrics, _ := registry.Gather()

	for _, mf := range metrics {
		if mf.GetName() == metricName {
			for _, m := range mf.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "grpc_method" && l.GetValue() == method {
						return int(m.GetCounter().GetValue())
					}
				}
			}
		}
	}
	return 0
}
