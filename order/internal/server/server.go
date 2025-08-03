package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	order_service "github.com/anarakinson/go_stonks/order/internal/app/order"
	order_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"
	"github.com/redis/go-redis/v9"

	"github.com/anarakinson/go_stonks_shared/pkg/grpc_helpers"
	"github.com/anarakinson/go_stonks_shared/pkg/interceptors"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"go.uber.org/zap"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type Server struct {
	port string
	repo order_service.Repository
}

func NewServer(port string, repo order_service.Repository) *Server {
	return &Server{
		port: port,
		repo: repo,
	}
}

func (s *Server) Run() error {

	//--------------------------------------------//
	// слушаем порт
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT")))
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}

	//--------------------------------------------//
	// создаем клиент редиса
	redisAddr := os.Getenv("REDIS_ADDRESS")
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return err
	}
	fmt.Println(redisAddr, redisPass, redisDB)
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     "redis:6379",
			Password: redisPass,
			DB:       redisDB,
		},
	)
	// пингуем редис
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// Создаем интерсептор
	cacheInterceptor := interceptors.NewRedisCacheInterceptor(redisClient)
	err = cacheInterceptor.Subscribe("markets:list", "markets:invalidated")
	if err != nil {
		return fmt.Errorf("failed on subscribe: %w", err)
	}

	// создаем сервер GRPC
	gs := grpc.NewServer(
		// OpenTelemetry трассировщик
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// добавляем интерцепторы
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,           // сбор данных для прометеуса
			interceptors.UnaryLoggingInterceptor(logger.Log), // логирование запросов и ошибок
			interceptors.XRequestIDServer(),                  // добавление x-request-id
			interceptors.UnaryPanicRecoveryInterceptor(),     // перехват и восстановление паники
		),
	)

	//--------------------------------------------//
	// создаем соединение с spot_instrument server
	logger.Log.Info(
		"Create connection to spot insrtument service",
		zap.String("service address", os.Getenv("SPOT_INSTRUMENT_ADDR")),
	)
	// Без TLS (для тестов)
	spotConn, err := grpc_helpers.NewGRPCClient(
		os.Getenv("SPOT_INSTRUMENT_ADDR"),
		nil, // TLS настройки
		// интерсепторы
		interceptors.XRequestIDClient(), // x-request-id interceptor
		cacheInterceptor.Unary(
			"markets:list",
			spot_inst_pb.SpotInstrumentService_ViewMarkets_FullMethodName,
			5*time.Minute,
		),
		// interceptors.TimeoutAdjusterClientInterceptor(0.8), // интерсептор для уменьшения времени таймаута контекта
	)
	if err != nil {
		return fmt.Errorf("order service connection failed: %w", err)
	}
	defer spotConn.Close()
	// проверка доступности спот сервиса
	spotClient := spot_inst_pb.NewSpotInstrumentServiceClient(spotConn)

	//--------------------------------------------//
	// создаем сервис
	orderService := order_service.NewService(spotClient, s.repo)
	// регистрируем
	order_pb.RegisterOrderServiceServer(gs, orderService)

	logger.Log.Info(
		"Order service started",
		zap.String("listening address", fmt.Sprintf("%v", lis.Addr())),
	)

	return gs.Serve(lis)

}
