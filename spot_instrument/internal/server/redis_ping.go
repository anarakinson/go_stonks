package server

import (
	"context"
	"time"

	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// проверяет здоровье редис подключения (пингует редис)
// в случае проблем пытается восстановить соединение
// принимает
// контекст
// интервал, с которым происходит пинг
func (s *Server) StartRedisMonitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Проверяем соединение с таймаутом
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			_, err := s.redisClient.Ping(pingCtx).Result()
			cancel()

			if err != nil {
				logger.Log.Error("Redis connection lost", zap.Error(err))

				// Попытка восстановления соединения
				logger.Log.Warn("Attempting to reconnect")

				// Создаем новый клиент (старый может быть в невалидном состоянии)
				newClient := redis.NewClient(s.redisClient.Options())

				// Проверяем новое соединение
				pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				_, err := newClient.Ping(pingCtx).Result()
				cancel()

				if err == nil {
					// Успешное восстановление
					s.redisClient = newClient // Заменяем клиент
					logger.Log.Info("Redis connection restored")
				}
			} else {
				logger.Log.Debug("Redis heartbeat OK")
			}
		}
	}
}

/*
как альтернатива, можно указать какое то количество попыток переподключения,
и после того, как они закончатся, останавливать сервер

				// Попытка восстановления соединения
				for attempt := 1; attempt <= maxRetries; attempt++ {

					// Экспоненциальная задержка
					time.Sleep(baseDelay * time.Duration(attempt))

					// Создаем новый клиент (старый может быть в невалидном состоянии)
					newClient := redis.NewClient(s.redisClient.Options())

					// Проверяем новое соединение
					pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
					_, err := newClient.Ping(pingCtx).Result()
					cancel()

					if err == nil {
						// Успешное восстановление
						s.redisClient = newClient // Заменяем клиент
						logger.Log.Info("Redis connection restored")
						break
					}

					if attempt == maxRetries {
						logger.Log.Error("Failed to restore Redis connection after retries")
					}

*/
