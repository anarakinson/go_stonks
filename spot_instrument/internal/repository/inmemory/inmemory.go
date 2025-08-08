package inmemory

import (
	"errors"
	"sync"

	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
)

var ErrOrderCollision = errors.New("order already exists")
var ErrMarketCollision = errors.New("market already exists")

type Repository struct {
	markets map[string]*domain.Market
	mu      sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{
		markets: make(map[string]*domain.Market),
	}
}
