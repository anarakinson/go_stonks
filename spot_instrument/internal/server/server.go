package server

import (
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/anarakinson/go_stonks/shared/pkg/interceptors"
	spot_instrument_service "github.com/anarakinson/go_stonks/spot_instrument_service/internal/app/spot_instrument"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		slog.Error("SpotInstrumentService failed to listen", "error", err)
		return err
	}

	// создаем сервер GRPC
	gs := grpc.NewServer(
		// добавляем интерцепторы
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor, // сбор данных для прометеуса
			interceptors.XRequestIDServer(),        // добавление x-request-id
			interceptors.UnaryPanicRecovery(),      // перехват и восстановление паники
		),
	)

	//--------------------------------------------//
	// создаем сервис
	spotService := spot_instrument_service.NewService(s.repo)

	// регистрируем сервис на сервере
	pb.RegisterSpotInstrumentServiceServer(gs, spotService)

	log.Printf("SpotInstrumentService started on %v", lis.Addr())
	return gs.Serve(lis)
}
