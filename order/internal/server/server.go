package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"log/slog"

	order_service "github.com/anarakinson/go_stonks/order_service/internal/app/order"
	"github.com/anarakinson/go_stonks/order_service/pkg/interceptors"
	order_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		slog.Error("Order service failed to listen", "error", err)
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
	// создаем соединение с spot_instrument server
	spotConn, err := grpc.NewClient(
		// fmt.Sprintf("spot_instrument:%s", os.Getenv("SPOT_INSTRUMENT_PORT")),
		os.Getenv("SPOT_INSTRUMENT_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
		grpc.WithUnaryInterceptor(interceptors.XRequestIDClient()), // x-request-id interceptor

	)
	if err != nil {
		log.Fatalf("Order service connection failed: %v", err)
	}
	defer spotConn.Close()
	// проверка доступности спот сервиса
	spotClient := spot_inst_pb.NewSpotInstrumentServiceClient(spotConn)

	//--------------------------------------------//
	// создаем сервис
	orderService := order_service.NewService(spotClient, s.repo)
	// регистрируем
	order_pb.RegisterOrderServiceServer(gs, orderService)

	log.Printf("Order service started on %v", lis.Addr())
	return gs.Serve(lis)

}
