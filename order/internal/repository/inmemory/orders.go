package inmemory

import (
	"fmt"
	"time"

	"github.com/anarakinson/go_stonks/order/internal/domain"
)

func (r *Repository) AddOrder(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; exists {
		return ErrOrderCollision
	}
	el := element{order, time.Now().Add(r.ttl)}
	r.orders[order.ID] = el
	r.orderId[order.ID] = true
	return nil
}

func (r *Repository) UpdateOrder(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	el, ok := r.orders[order.ID]
	if !ok {
		return fmt.Errorf("order %s does not exists or expired", order.ID)
	}
	el.order = order
	r.orders[order.ID] = el
	return nil
}

func (r *Repository) GetOrder(orderID string) (*domain.Order, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.orders[orderID]
	return v.order, ok
}

func (r *Repository) GetUserOrders(UserId string) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Order

	for _, o := range r.orders {
		if o.order.UserID == UserId {
			if time.Now().Before(o.expiredAt) {
				out = append(out, *o.order)
			}
		}
	}
	return out, nil
}
