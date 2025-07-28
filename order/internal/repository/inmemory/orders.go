package inmemory

import "github.com/anarakinson/go_stonks/order_service/internal/domain"

func (r *Repository) AddOrder(order *domain.Order) error {
	if _, exists := r.orders[order.ID]; exists {
		return ErrOrderCollision
	}
	r.orders[order.ID] = order
	return nil
}

func (r *Repository) GetOrder(orderID string) (*domain.Order, bool) {
	v, ok := r.orders[orderID]
	return v, ok
}

func (r *Repository) GetUserOrders(UserId string) ([]*domain.Order, error) {
	var out []*domain.Order
	for _, o := range r.orders {
		if o.UserID == UserId {
			out = append(out, o)
		}
	}
	return out, nil
}
