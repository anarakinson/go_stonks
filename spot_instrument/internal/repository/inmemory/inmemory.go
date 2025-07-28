package inmemory

import (
	"errors"
	"github.com/anarakinson/go_stonks/spot_instrument_service/internal/domain"
)

var ErrOrderCollision = errors.New("order already exists")
var ErrMarketCollision = errors.New("market already exists")

type Repository struct {
	markets map[string]*domain.Market
}

func NewRepository() *Repository {
	return &Repository{
		markets: make(map[string]*domain.Market),
	}
}
