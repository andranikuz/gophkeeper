package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	// ContextKeyUserID используется для хранения ID пользователя в контексте.
	ContextKeyUserID = "userID"
)

// Authenticator инкапсулирует логику работы с аутентификацией.
type Authenticator struct {
	tokenSecret     string
	tokenExpiration time.Duration
}

// NewAuthenticator создаёт новый экземпляр Authenticator.
// tokenSecret — секрет для подписи JWT, tokenExpirationSeconds — время жизни токена в секундах.
func NewAuthenticator(tokenSecret string, tokenExpirationSeconds int) *Authenticator {
	return &Authenticator{
		tokenSecret:     tokenSecret,
		tokenExpiration: time.Duration(tokenExpirationSeconds) * time.Second,
	}
}

// HashPassword принимает пароль в виде строки и возвращает его bcrypt-хэш.
func (a *Authenticator) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash сравнивает сырой пароль с хэшированным и возвращает true, если они совпадают.
func (a *Authenticator) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Claims определяет структуру данных, содержащую информацию о пользователе для JWT.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenerateToken генерирует JWT-токен для пользователя с указанным идентификатором и именем.
// Токен подписывается секретным ключом и имеет ограниченный срок действия.
func (a *Authenticator) GenerateToken(userID string, username string) (string, error) {
	expirationTime := time.Now().Add(a.tokenExpiration)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "gophkeeper",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateToken проверяет валидность переданного JWT-токена.
// Если токен корректен, возвращаются данные из claims, иначе — ошибка.
func (a *Authenticator) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписи является HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.tokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// Middleware возвращает middleware для проверки JWT-токена.
// secret — секрет для проверки подписи JWT.
func (a *Authenticator) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			tokenStr := parts[1]
			claims, err := a.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
