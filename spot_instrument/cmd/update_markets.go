package main

import (
	"context"
	"fmt"
	"time"

	"github.com/anarakinson/go_stonks/spot_instrument/internal/app/spot_instrument"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
	"github.com/redis/go-redis/v9"
)

// функция периодически добавляет и удаляет новый маркет в репозиторий
func StartUpdatingMarkets(ctx context.Context, repo spot_instrument.Repository, redisClient *redis.Client) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ctx.Done():
			// Контекст отменен, выходим из функции
			return
		case <-ticker.C:

			// создаем новый безымянный	маркет
			solMarket := domain.NewMarket(fmt.Sprintf("UnnamedMarket#%d", counter), true)

			// добавляем безымянный маркет в репозиторий
			repo.AddMarket(solMarket)

			// ждем минуту
			time.Sleep(2 * time.Minute)
			// удаляем новый маркет
			solMarket.Delete()

			// оповещаем редис через PubSub
			redisClient.Publish(ctx, "markets:invalidated", "markets:list")

		}
	}
}
