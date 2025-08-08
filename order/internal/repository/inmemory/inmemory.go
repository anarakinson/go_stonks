package inmemory

import (
	"errors"
	"sync"

	"github.com/anarakinson/go_stonks/order/internal/domain"
)

var ErrOrderCollision = errors.New("order already exists")
var ErrMarketCollision = errors.New("market already exists")

type Repository struct {
	orders  map[string]*domain.Order
	orderId map[string]bool
	mu sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{
		orders:  make(map[string]*domain.Order),
		orderId: make(map[string]bool),
	}
}
