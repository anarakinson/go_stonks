package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"

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
	tp, err := tracing.InitTracerProvider("jaeger:4317", "spot_instrument-service", "1.0.0", "development", nil)
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
	fmt.Println(redisAddr, redisPass, redisDB)
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     "redis:6379",
			Password: redisPass,
			DB:       redisDB,
		},
	)
	// пингуем редис
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Log.Error("redis ping failed", zap.Error(err))
		return
	}

	//--------------------------------------------//
	// создаем хранилище с моками рынков
	repo := inmemory.NewRepository()
	repo.AddMarket(domain.NewMarket("BTC-USDT", true))
	repo.AddMarket(domain.NewMarket("BTC-USDC", true))
	repo.AddMarket(domain.NewMarket("ETH-USDT", false))
	repo.AddMarket(domain.NewMarket("ETH-USDC", true))
	solMarket := domain.NewMarket("SOL/USDT", true)
	solMarket.Delete()
	repo.AddMarket(solMarket)

	// запускаем фоновое обновление маркетов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go server.StartUpdatingMarkets(ctx, repo, redisClient)

	//--------------------------------------------//
	// создаем и запускаем сервер
	serv := server.NewServer(os.Getenv("PORT"), repo)
	err = serv.Run()
	if err != nil {
		logger.Log.Error(
			"Failed on serve",
			zap.Error(err),
		)
		return
	}
}
