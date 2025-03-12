package grpcserver

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/andranikuz/gophkeeper/internal/auth"
)

// JwtUnaryInterceptor проверяет JWT для униарных gRPC вызовов.
func JwtUnaryInterceptor(authenticator *auth.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// Извлекаем метаданные из контекста.
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return nil, errors.New("authorization token is not supplied")
		}

		tokenStr := authHeaders[0]
		parts := strings.Split(tokenStr, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return nil, errors.New("invalid authorization format")
		}
		tokenStr = parts[1]

		// Проверяем токен.
		claims, err := authenticator.ValidateToken(tokenStr)
		if err != nil {
			return nil, errors.New("invalid token")
		}
		ctx = context.WithValue(ctx, auth.ContextKeyUserID, claims.UserID)
		// Токен действителен, вызываем обработчик.
		return handler(ctx, req)
	}
}

// JwtStreamInterceptor проверяет JWT для стримовых gRPC вызовов и добавляет userID в контекст.
func JwtStreamInterceptor(authenticator *auth.Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return errors.New("missing metadata")
		}
		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return errors.New("authorization token is not supplied")
		}
		tokenStr := authHeaders[0]
		parts := strings.Split(tokenStr, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return errors.New("invalid authorization format")
		}
		tokenStr = parts[1]

		// Проверяем токен.
		claims, err := authenticator.ValidateToken(tokenStr)
		if err != nil {
			return errors.New("invalid token")
		}
		// Пробрасываем userID в контекст стрима.
		newCtx := context.WithValue(stream.Context(), auth.ContextKeyUserID, claims.UserID)
		wrapped := grpcServerStreamWithContext{ServerStream: stream, ctx: newCtx}
		return handler(srv, wrapped)
	}
}

// grpcServerStreamWithContext оборачивает grpc.ServerStream, позволяя изменять контекст.
type grpcServerStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (w grpcServerStreamWithContext) Context() context.Context {
	return w.ctx
}
