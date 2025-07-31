package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TracingInterceptor(
	ctx context.Context,
	method string, req,
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// Получаем tracer
	tracer := otel.Tracer("grpc-client")

	// Извлекаем x-request-id из контекста или метаданных
	var requestID string
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		requestIDs := md.Get("x-request-id")
		if len(requestIDs) > 0 {
			requestID = requestIDs[0]
		}
	}

	// Создаем span с x-request-id в атрибутах
	ctx, span := tracer.Start(
		ctx,
		"grpc."+method,
		trace.WithAttributes(
			attribute.String("x-request-id", requestID),
			attribute.String("grpc.method", method),
		),
	)
	defer span.End()

	// Продолжаем выполнение
	return invoker(ctx, method, req, reply, cc, opts...)
}
