package inmemory

import (
	"errors"
	"sync"
	"time"

	"github.com/anarakinson/go_stonks/order/internal/domain"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"go.uber.org/zap"
)

var ErrOrderCollision = errors.New("order already exists")
var ErrMarketCollision = errors.New("market already exists")

type element struct {
	order     *domain.Order
	expiredAt time.Time
}

type Repository struct {
	orders   map[string]element
	orderId  map[string]bool
	ttl      time.Duration
	stopChan chan struct{}
	mu       sync.RWMutex
}

func NewRepository(ttl time.Duration) *Repository {
	r := &Repository{
		orders:  make(map[string]element),
		orderId: make(map[string]bool),
		ttl:     ttl,
	}
	go r.startCleaning()
	return r
}

func (r *Repository) Stop() {
	r.stopChan <- struct{}{}
}

func (r *Repository) startCleaning() {
	logger.Log.Info("Repository cleaning process started")
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	keysToDelete := []string{}

	for {
		select {
		case <-r.stopChan:
			logger.Log.Info("Repository cleaning process stopped")
			return
		case <-ticker.C:
			keysToDelete = keysToDelete[:0]
			// блокировка на чтение и поиск кандидатов на удаление
			r.mu.RLock()
			for k, el := range r.orders {
				if time.Now().After(el.expiredAt) {
					keysToDelete = append(keysToDelete, k)
				}
			}
			r.mu.RUnlock()

			// полная блокировка и удаление всех ключей
			logger.Log.Info("Deleting elements from repository", zap.Int("keys number", len(keysToDelete)))
			r.mu.Lock()
			for _, k := range keysToDelete {
				delete(r.orders, k)
				delete(r.orderId, k)
			}
			r.mu.Unlock()
		}
	}
}
