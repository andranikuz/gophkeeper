package services

import (
	"context"

	"github.com/golang-jwt/jwt"
)

// Claims определяет структуру данных, содержащую информацию о пользователе для JWT.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// AuthenticatorInterface определяет методы для работы с аутентификацией.
type AuthenticatorInterface interface {
	// HashPassword принимает пароль в виде строки и возвращает его bcrypt-хэш.
	HashPassword(password string) (string, error)
	// CheckPasswordHash сравнивает сырой пароль с хэшированным и возвращает true, если они совпадают.
	CheckPasswordHash(password, hash string) bool
	// GenerateToken генерирует JWT-токен для пользователя с указанным идентификатором и именем.
	GenerateToken(userID string, username string) (string, error)
	// ValidateToken проверяет валидность переданного JWT-токена и возвращает данные из claims.
	ValidateToken(tokenStr string) (*Claims, error)
	// GetUserIdFromCtx извлекает userID из контекста.
	GetUserIdFromCtx(ctx context.Context) (string, error)
}
