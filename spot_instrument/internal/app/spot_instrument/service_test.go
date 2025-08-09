package spot_instrument_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/anarakinson/go_stonks/spot_instrument/internal/app/spot_instrument"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/mocks"
	spot_inst_pb "github.com/anarakinson/go_stonks/stonks_pb/gen/spot_instrument/v1"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestService_ViewMarkets(t *testing.T) {
	//--------------------------------------------//
	// инициализируем логгер
	if err := logger.Init("test"); err != nil {
		slog.Error("Unable to init zap-logger", "error", err)
		return
	}
	defer logger.Sync()

	// засекаем время
	deletedTime := time.Now()
	tests := []struct {
		name     string
		markets  []*domain.Market
		expected int // ожидаемое количество активных рынков
	}{
		{
			name: "only active markets",
			markets: []*domain.Market{
				{ID: "1", Name: "BTC/USD", Enabled: true, DeletedAt: nil},
				{ID: "2", Name: "ETH/USD", Enabled: true, DeletedAt: nil},
			},
			expected: 2,
		},
		{
			name: "mixed active and inactive markets",
			markets: []*domain.Market{
				{ID: "1", Name: "BTC/USD", Enabled: true, DeletedAt: nil},
				{ID: "2", Name: "ETH/USD", Enabled: false, DeletedAt: nil},
				{ID: "3", Name: "XRP/USD", Enabled: true, DeletedAt: &deletedTime},
				{ID: "4", Name: "LTC/USD", Enabled: true, DeletedAt: nil},
			},
			expected: 2, // только 1 и 4 рынки активны
		},
		{
			name:     "no markets",
			markets:  []*domain.Market{},
			expected: 0,
		},
		{
			name: "all markets inactive",
			markets: []*domain.Market{
				{ID: "1", Name: "BTC/USD", Enabled: false, DeletedAt: nil},
				{ID: "2", Name: "ETH/USD", Enabled: true, DeletedAt: &deletedTime},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок репозитория
			repo := mocks.NewMockRepository(tt.markets)
			// Создаем сервис
			service := spot_instrument.NewService(repo)

			// Вызываем метод
			resp, err := service.ViewMarkets(context.Background(), &spot_inst_pb.ViewMarketsRequest{})

			fmt.Println(resp)

			// Проверяем что нет ошибки
			assert.NoError(t, err)
			// Проверяем количество возвращенных рынков
			assert.Equal(t, tt.expected, len(resp.Markets))

			// Дополнительно проверяем что все возвращенные рынки активны
			for _, m := range resp.Markets {
				assert.True(t, m.Enabled, "market should be enabled")
				assert.Nil(t, m.DeletedAt, "market should not be deleted")
			}
		})
	}
}
