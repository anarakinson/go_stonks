package server

import (
	"fmt"
	"net"

	spot_instrument_service "github.com/anarakinson/go_stonks/spot_instrument/internal/app/spot_instrument"

	"github.com/anarakinson/go_stonks/stonks_shared/pkg/interceptors"
	"github.com/anarakinson/go_stonks/stonks_shared/pkg/logger"
	"go.uber.org/zap"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type Server struct {
	port string
	repo spot_instrument_service.Repository
}

func NewServer(port string, repo spot_instrument_service.Repository) *Server {
	return &Server{
		port: port,
		repo: repo,
	}
}

func (s *Server) Run() error {
	//--------------------------------------------//
	// слушаем порт
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", s.port))
	if err != nil {
		logger.Log.Error("Order service failed to listen", zap.Error(err))
		return err
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
	// создаем сервис
	spotService := spot_instrument_service.NewService(s.repo)

	// регистрируем сервис на сервере
	pb.RegisterSpotInstrumentServiceServer(gs, spotService)

	logger.Log.Info(
		"Order service started",
		zap.String("listening address", fmt.Sprintf("%v", lis.Addr())),
	)

	return gs.Serve(lis)
}
