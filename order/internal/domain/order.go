package domain

import (
	"github.com/google/uuid"
)

type Order struct {
	ID        string
	UserID    string
	MarketID  string
	OrderType string
	Status    string
	Price     float64
	Quantity  float64
}

func NewOrder(
	UserID,
	MarketID,
	OrderType string,
	Price,
	Quantity float64,
) *Order {
	return &Order{
		ID:        uuid.New().String(),
		UserID:    UserID,
		MarketID:  MarketID,
		OrderType: OrderType,
		Price:     Price,
		Quantity:  Quantity,
		Status:    "created",
	}
}
