package mocks

import (
	"context"

	market_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/market/v1"
	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument/v1"
	"google.golang.org/grpc"
)

// MockSpotInstrumentService реализует интерфейс SpotInstrumentServiceClient
type MockSpotInstrumentService struct {
	ViewMarketsFunc func(ctx context.Context, in *spot_inst_pb.ViewMarketsRequest, opts ...grpc.CallOption) (*spot_inst_pb.ViewMarketsResponse, error)
}

// ViewMarkets вызывает мок-функцию
func (m *MockSpotInstrumentService) ViewMarkets(ctx context.Context, in *spot_inst_pb.ViewMarketsRequest, opts ...grpc.CallOption) (*spot_inst_pb.ViewMarketsResponse, error) {
	if m.ViewMarketsFunc != nil {
		return m.ViewMarketsFunc(ctx, in, opts...)
	}
	return &spot_inst_pb.ViewMarketsResponse{}, nil
}

// NewMockSpotInstrumentService создает новый мок с дефолтным поведением
func NewMockSpotInstrumentService() *MockSpotInstrumentService {
	return &MockSpotInstrumentService{
		ViewMarketsFunc: func(ctx context.Context, in *spot_inst_pb.ViewMarketsRequest, opts ...grpc.CallOption) (*spot_inst_pb.ViewMarketsResponse, error) {
			return &spot_inst_pb.ViewMarketsResponse{
				Markets: []*market_pb.Market{
					{Id: "1", Name: "BTC/USD", Enabled: true,
						AllowedRoles: []market_pb.UserRole{market_pb.UserRole_ROLE_BASIC}}, // доступен
				},
			}, nil
		},
	}
}
