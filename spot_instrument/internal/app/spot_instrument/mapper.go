package spot_instrument

import (
	"spot_instrument_service/internal/domain"
	prototime "spot_instrument_service/pkg/proto_time"
	market_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/market"
)

// MarketToProto преобразует доменную сущность Market в proto-структуру
func MarketToProto(m *domain.Market) *market_pb.Market {
	return &market_pb.Market{
		Id:        m.ID,
		Name:      m.Name,
		Enabled:   m.Enabled,
		DeletedAt: prototime.ToProtoTime(m.DeletedAt),
	}
}

// ProtoToMarket proto-структуру в преобразует доменную сущность Market
func ProtoToMarket(m *market_pb.Market) *domain.Market {
	return &domain.Market{
		ID:        m.Id,
		Name:      m.Name,
		Enabled:   m.Enabled,
		DeletedAt: prototime.FromProtoTime(m.DeletedAt),
	}
}
