package mocks

import "github.com/anarakinson/go_stonks/order/internal/domain"

type Repository interface {
	AddOrder(*domain.Order) error
	GetOrder(string) (*domain.Order, bool)
	GetUserOrders(UserId string) ([]*domain.Order, error)
}

// MockRepository - мок репозитория для тестов
type MockRepository struct {
	orders map[string]*domain.Order
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		orders: make(map[string]*domain.Order),
	}
}

func (m *MockRepository) GetUserOrders(UserId string) ([]*domain.Order, error) {
	orders := []*domain.Order{}
	for _, v := range m.orders {
		orders = append(orders, v)
	}
	return orders, nil
}

func (m *MockRepository) GetOrder(id string) (*domain.Order, bool) { 
	v, ok := m.orders[id]
	return v, ok
}

func (m *MockRepository) AddOrder(order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}
