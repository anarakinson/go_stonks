package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/anarakinson/go_stonks/order/internal/repository/inmemory"
	"github.com/anarakinson/go_stonks/order/internal/server"
	"github.com/redis/go-redis/v9"

	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/anarakinson/go_stonks_shared/pkg/metrics"
	"github.com/anarakinson/go_stonks_shared/pkg/tracing"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

func main() {

	//--------------------------------------------//
	// Канал для graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	//--------------------------------------------//
	// загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "error", err)
		return
	}

	//--------------------------------------------//
	// инициализируем логгер
	if err := logger.Init("production"); err != nil {
		slog.Error("Unable to init zap-logger", "error", err)
		return
	}
	defer logger.Sync()

	//--------------------------------------------//
	// создаем и запускаем сервер сборки метрик
	go func() {
		if err := metrics.RunMetricsServer(); err != nil {
			logger.Log.Error("Metrics server error:", zap.Error(err))
		}
	}()

	//--------------------------------------------//
	// инициализация трейсинга jaegar
	jaegarAddr := fmt.Sprintf("%s:%s", os.Getenv("JAEGER_HOST"), os.Getenv("JAEGER_PORT"))
	tp, err := tracing.InitTracerProvider(jaegarAddr, "order-service", "1.0.0", "development", nil)
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}()

	//--------------------------------------------//
	// создаем клиент редиса
	redisAddr := os.Getenv("REDIS_ADDRESS")
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		slog.Error("Error loading REDIS_DB env variable", "error", err)
		return
	}

	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     redisAddr,
			Password: redisPass,
			DB:       redisDB,
		},
	)
	defer redisClient.Close()

	// пингуем редис с повторными попытками
	ctx := context.Background()
	maxAttempts, err := strconv.Atoi(os.Getenv("REDIS_PING_NUM"))
	if err != nil {
		maxAttempts = 5
	}
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		_, err = redisClient.Ping(ctx).Result()
		if err == nil {
			logger.Log.Info("Redis connection successful")
			break
		}

		logger.Log.Warn("Redis ping failed",
			zap.Int("attempt", attempt),
			zap.Error(err))

		if attempt < maxAttempts {
			logger.Log.Info("Retrying...", zap.Duration("delay", retryDelay))
			time.Sleep(retryDelay)
			// Увеличиваем задержку перед следующей попыткой
			retryDelay *= 2
		} else {
			logger.Log.Fatal("All Redis connection attempts failed", zap.Error(err))
		}
	}

	//--------------------------------------------//
	// создаем хранилище
	repo := inmemory.NewRepository(3 * time.Minute)
	defer repo.Stop()

	//--------------------------------------------//
	// создаем и запускаем сервер
	serv := server.NewServer(os.Getenv("PORT"), repo, redisClient)
	// запускаем пинг редис сервиса каждые 15 секунд
	go serv.StartRedisMonitor(ctx, 15*time.Second)
	// запускаем сервер c грейсфул шатдаун
	errChan := make(chan error, 1)
	go func() {
		err = serv.Run()
		if err != nil {
			logger.Log.Error(
				"Failed on serve",
				zap.Error(err),
			)
			errChan <- err
			return
		}
	}()

	// грейсфул шатдаун
	// Ждем либо сигнал завершения, либо ошибку сервера
	select {
	case err := <-errChan:
		logger.Log.Error("Server error", zap.Error(err))
	case <-shutdown:
		logger.Log.Info("Server is shutting down...")
		serv.Shutdown()
	}
}
