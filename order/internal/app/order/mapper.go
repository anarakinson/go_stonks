package order_service

import (
	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order"
	"order_service/internal/domain"
)

// OrderToProto преобразует доменную сущность Order в proto-структуру
func OrderToProto(o *domain.Order) *pb.Order {
	return &pb.Order{
		Id:        o.ID,
		UserId:    o.UserID,
		MarketId:  o.MarketID,
		OrderType: o.OrderType,
		Price:     o.Price,
		Quantity:  o.Quantity,
		Status:    o.Status,
	}
}

// ProtoToOrder proto-структуру в преобразует доменную сущность Order
func ProtoToOrder(p *pb.Order) *domain.Order {
	return &domain.Order{
		ID:        p.Id,
		UserID:    p.UserId,
		MarketID:  p.MarketId,
		OrderType: p.OrderType,
		Status:    p.Status,
		Price:     p.Price,
		Quantity:  p.Quantity,
	}
}
