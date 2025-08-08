package inmemory

import "github.com/anarakinson/go_stonks/order/internal/domain"

func (r *Repository) AddOrder(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; exists {
		return ErrOrderCollision
	}
	r.orders[order.ID] = order
	return nil
}

func (r *Repository) UpdateOrder(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

func (r *Repository) GetOrder(orderID string) (*domain.Order, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.orders[orderID]
	return v, ok
}

func (r *Repository) GetUserOrders(UserId string) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Order

	for _, o := range r.orders {
		if o.UserID == UserId {
			out = append(out, *o)
		}
	}
	return out, nil
}
