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

func (r *Repository) GetAvailableMarkets() ([]*domain.Market, error) {
	var available []*domain.Market
	for _, mrkt := range r.markets {
		if mrkt.Enabled && mrkt.DeletedAt == nil {
			available = append(available, mrkt)
		}
	}
	return available, nil
}
