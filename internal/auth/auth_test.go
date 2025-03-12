package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndCheckPassword(t *testing.T) {
	a := NewAuthenticator("secret", 60)
	password := "password123"

	// Хэширование пароля
	hash, err := a.HashPassword(password)
	require.NoError(t, err, "Ошибка при хэшировании пароля")
	require.NotEmpty(t, hash, "Хэш не должен быть пустым")

	// Убедимся, что хэш не равен исходному паролю
	assert.NotEqual(t, password, hash)

	// Проверка совпадения пароля и хэша
	assert.True(t, a.CheckPasswordHash(password, hash), "Пароль должен совпадать с хэшем")
	assert.False(t, a.CheckPasswordHash("wrongpassword", hash), "Неверный пароль не должен совпадать с хэшем")
}

func TestGenerateAndValidateToken(t *testing.T) {
	tokenSecret := "secret"
	a := NewAuthenticator(tokenSecret, 60)
	userID := "user123"
	username := "john_doe"

	// Генерация токена
	token, err := a.GenerateToken(userID, username)
	require.NoError(t, err, "Ошибка при генерации токена")
	require.NotEmpty(t, token, "Токен не должен быть пустым")

	// Валидация токена
	claims, err := a.ValidateToken(token)
	require.NoError(t, err, "Ошибка при валидации токена")
	require.NotNil(t, claims, "Claims не должны быть nil")

	assert.Equal(t, userID, claims.UserID, "Некорректный userID в claims")
	assert.Equal(t, username, claims.Username, "Некорректное имя пользователя в claims")
	assert.Equal(t, "gophkeeper", claims.Issuer, "Некорректный issuer в claims")
	// Проверка, что время истечения токена установлено в будущем
	assert.True(t, claims.ExpiresAt > time.Now().Unix(), "Время истечения токена должно быть в будущем")
}

func TestValidateTokenInvalid(t *testing.T) {
	a := NewAuthenticator("secret", 60)

	// Передача явно некорректного токена
	_, err := a.ValidateToken("invalid.token.here")
	assert.Error(t, err, "Ожидается ошибка при валидации некорректного токена")

	// Генерация токена с правильным секретом, а затем проверка с неправильным секретом
	token, err := a.GenerateToken("user123", "john_doe")
	require.NoError(t, err)

	// Создаем новый экземпляр с другим секретом
	aWrong := NewAuthenticator("wrongsecret", 60)
	_, err = aWrong.ValidateToken(token)
	assert.Error(t, err, "Ожидается ошибка при валидации токена с неправильным секретом")
}

func TestValidateExpiredToken(t *testing.T) {
	// Устанавливаем отрицательное время жизни токена, чтобы он сразу считался просроченным
	a := NewAuthenticator("secret", -1)
	token, err := a.GenerateToken("user123", "john_doe")
	require.NoError(t, err)

	_, err = a.ValidateToken(token)
	assert.Error(t, err, "Ожидается ошибка при валидации просроченного токена")
	// Возможное сообщение об ошибке может содержать "expired"
	assert.Contains(t, err.Error(), "expired", "Ошибка должна указывать на истечение срока действия токена")
}

func TestGetUserIdFromCtx(t *testing.T) {
	a := NewAuthenticator("secret", 60)

	// Контекст с корректным значением userID
	ctx := context.WithValue(context.Background(), ContextKeyUserID, "user123")
	userID, err := a.GetUserIdFromCtx(ctx)
	require.NoError(t, err, "Ошибка при получении userID из контекста")
	assert.Equal(t, "user123", userID, "Неверный userID, полученный из контекста")

	// Контекст без userID
	ctxEmpty := context.Background()
	_, err = a.GetUserIdFromCtx(ctxEmpty)
	assert.Error(t, err, "Ожидается ошибка при отсутствии userID в контексте")
	assert.Equal(t, "userID not found in context", err.Error(), "Неверное сообщение об ошибке при отсутствии userID")
}
