package mocks

import (
	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
)

// MockRepository - мок репозитория для тестов
type MockRepository struct {
	markets []*domain.Market
}

func NewMockRepository(markets []*domain.Market) *MockRepository {
	return &MockRepository{
		markets: markets,
	}
}

func (m *MockRepository) GetMarkets() ([]domain.Market, error) {
	markets := make([]domain.Market, 0, len(m.markets))
	for _, m := range m.markets {
		markets = append(markets, *m)
	}
	return markets, nil
}

func (m *MockRepository) GetMarket(string) (*domain.Market, bool) { return nil, false }
func (m *MockRepository) AddMarket(*domain.Market) error          { return nil }
