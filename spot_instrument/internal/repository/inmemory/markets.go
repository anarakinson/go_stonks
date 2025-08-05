package inmemory

import "github.com/anarakinson/go_stonks/spot_instrument/internal/domain"

func (r *Repository) AddMarket(market *domain.Market) error {
	if _, exists := r.markets[market.ID]; exists {
		return ErrMarketCollision
	}
	r.markets[market.ID] = market
	return nil
}

func (r *Repository) GetMarket(marketID string) (*domain.Market, bool) {
	v, ok := r.markets[marketID]
	return v, ok
}

func (r *Repository) GetMarkets() ([]*domain.Market, error) {
	var markets []*domain.Market
	for _, mrkt := range r.markets {
		markets = append(markets, mrkt)
	}
	return markets, nil
}
