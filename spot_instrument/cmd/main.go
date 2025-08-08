package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/anarakinson/go_stonks/spot_instrument/internal/app/spot_instrument"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/domain"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/repository/inmemory"
	"github.com/anarakinson/go_stonks/spot_instrument/internal/server"
	"github.com/anarakinson/go_stonks_shared/pkg/logger"
	"github.com/anarakinson/go_stonks_shared/pkg/metrics"
	"github.com/anarakinson/go_stonks_shared/pkg/tracing"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

func main() {

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
	tp, err := tracing.InitTracerProvider(jaegarAddr, "spot_instrument-service", "1.0.0", "development", nil)
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
	// создаем хранилище с моками рынков
	repo := inmemory.NewRepository()
	repo.AddMarket(domain.NewMarket("BTC-USDT", true, domain.UserRole_BASIC, domain.UserRole_PROFESSIONAL, domain.UserRole_WHALE))
	repo.AddMarket(domain.NewMarket("BTC-USDC", true, domain.UserRole_BASIC, domain.UserRole_PROFESSIONAL, domain.UserRole_WHALE))
	repo.AddMarket(domain.NewMarket("ETH-USDT", false, domain.UserRole_PROFESSIONAL, domain.UserRole_WHALE))
	repo.AddMarket(domain.NewMarket("ETH-USDC", true, domain.UserRole_PROFESSIONAL, domain.UserRole_WHALE))
	solMarket := domain.NewMarket("SOL/USDT", true, domain.UserRole_WHALE)
	solMarket.Delete()
	repo.AddMarket(solMarket)

	// запускаем фоновое обновление маркетов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go spot_instrument.StartUpdatingMarkets(ctx, repo, redisClient)

	//--------------------------------------------//
	// создаем и запускаем сервер
	serv := server.NewServer(os.Getenv("PORT"), repo, redisClient)
	// запускаем пинг редис сервиса каждые 15 секунд
	go serv.StartRedisMonitor(ctx, 15*time.Second)
	// запускаем сервер
	err = serv.Run()
	if err != nil {
		logger.Log.Error(
			"Failed on serve",
			zap.Error(err),
		)
		return
	}
}
