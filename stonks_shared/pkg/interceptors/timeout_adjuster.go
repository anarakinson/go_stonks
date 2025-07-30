package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// TimeoutAdjusterInterceptor - перехватывает контекст и уменьшает до указанного размера
func TimeoutAdjusterInterceptor(fraction float64) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		adjustedCtx := trimContextTimeout(ctx, fraction)
		return handler(adjustedCtx, req)
	}
}

// trimContextTimeout - уменьшает таймаут до указанной доли, чтобы сервис успел ответить до обрыва соединения
func trimContextTimeout(ctx context.Context, fraction float64) context.Context {
	if deadline, ok := ctx.Deadline(); ok {
		timeRemaining := time.Until(deadline)
		newTimeout := time.Duration(float64(timeRemaining) * fraction)

		// Создаём дочерний контекст с новым таймаутом
		// Он автоматически отменится при отмене родительского ctx
		ctx, _ = context.WithTimeout(ctx, newTimeout)

	}
	return ctx
}
