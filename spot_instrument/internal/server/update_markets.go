package server

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
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ctx.Done():
			// Контекст отменен, выходим из функции
			return
		case <-ticker.C:

			// удаляем старый маркет, если он четный
			if counter%2 == 0 {
				oldMarket, ok := repo.GetMarket(fmt.Sprintf("UnnamedMarket#%d", counter))
				if ok {
					oldMarket.Delete()
				}
			}

			counter++

			// создаем новый безымянный	маркет
			newMarket := domain.NewMarket(fmt.Sprintf("UnnamedMarket#%d", counter), true)

			// добавляем безымянный маркет в репозиторий
			repo.AddMarket(newMarket)

			// оповещаем редис через PubSub
			redisClient.Publish(ctx, "markets:invalidated", "markets:list")

		}
	}
}
