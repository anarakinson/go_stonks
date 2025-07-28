package spot_instrument

import (
	"context"
	"github.com/anarakinson/go_stonks/spot_instrument_service/internal/domain"
	market_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/market"

	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Repository interface {
	AddMarket(*domain.Market) error
	GetMarket(string) (*domain.Market, bool)
	GetAvailableMarkets() ([]*domain.Market, error)
}

type Service struct {
	spot_inst_pb.UnimplementedSpotInstrumentServiceServer
	markets Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		markets: repo,
	}
}

func (s *Service) ViewMarkets(ctx context.Context, req *spot_inst_pb.ViewMarketsRequest) (*spot_inst_pb.ViewMarketsResponse, error) {

	var availableMarkes []*market_pb.Market
	// получаем доступные маркеты из хранилища
	available, err := s.markets.GetAvailableMarkets()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check markets: %v", err)
	}

	// преобразуем domain.Market в pb.Market
	for _, mrkt := range available {
		availableMarkes = append(
			availableMarkes,
			MarketToProto(mrkt),
		)
	}

	return &spot_inst_pb.ViewMarketsResponse{Markets: availableMarkes}, nil
}
