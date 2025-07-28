package inmemory

import (
	"errors"
	"order_service/internal/domain"
)

var ErrOrderCollision = errors.New("order already exists")
var ErrMarketCollision = errors.New("market already exists")

type Repository struct {
	orders map[string]*domain.Order
}

func NewRepository() *Repository {
	return &Repository{
		orders: make(map[string]*domain.Order),
	}
}
