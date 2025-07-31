package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

func InitTracerProvider(serviceName string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Подключаемся к Jaeger через OTLP/gRPC (порт 4317)
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("jaeger:4317"), // или localhost:4317
		otlptracegrpc.WithInsecure(),              // для тестов (без TLS)
	)
	if err != nil {
		return nil, err
	}

	// Ресурсы трейсов (метаданные сервиса)
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment("production"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Настраиваем TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(1.0), // 100% трейсов (для прода — 0.1)
			),
		),
	)

	// Устанавливаем глобальные настройки
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp, nil
}
