package spot_instrument

import (
	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
	prototime "github.com/anarakinson/go_stonks/spot_instrument/pkg/proto_time"
	market_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/market/v1"
)

// MarketToProto преобразует доменную сущность Market в proto-структуру
func MarketToProto(m *domain.Market) *market_pb.Market {
	AllowedRoles := []market_pb.UserRole{}
	for _, role := range m.AvailableRoles {
		switch role {
		case domain.UserRole_BASIC:
			AllowedRoles = append(AllowedRoles, market_pb.UserRole_ROLE_BASIC)
		case domain.UserRole_PROFESSIONAL:
			AllowedRoles = append(AllowedRoles, market_pb.UserRole_ROLE_PROFESSIONAL)
		case domain.UserRole_WHALE:
			AllowedRoles = append(AllowedRoles, market_pb.UserRole_ROLE_WHALE)
		}
	}

	return &market_pb.Market{
		Id:           m.ID,
		Name:         m.Name,
		Enabled:      m.Enabled,
		DeletedAt:    prototime.ToProtoTime(m.DeletedAt),
		AllowedRoles: AllowedRoles,
	}
}

// ProtoToMarket proto-структуру в преобразует доменную сущность Market
func ProtoToMarket(m *market_pb.Market) *domain.Market {
	AllowedRoles := []domain.UserRole{}
	for _, role := range m.AllowedRoles {
		switch role {
		case market_pb.UserRole_ROLE_BASIC:
			AllowedRoles = append(AllowedRoles, domain.UserRole_BASIC)
		case market_pb.UserRole_ROLE_PROFESSIONAL:
			AllowedRoles = append(AllowedRoles, domain.UserRole_PROFESSIONAL)
		case market_pb.UserRole_ROLE_WHALE:
			AllowedRoles = append(AllowedRoles, domain.UserRole_WHALE)
		}
	}

	return &domain.Market{
		ID:             m.Id,
		Name:           m.Name,
		Enabled:        m.Enabled,
		DeletedAt:      prototime.FromProtoTime(m.DeletedAt),
		AvailableRoles: AllowedRoles,
	}
}
