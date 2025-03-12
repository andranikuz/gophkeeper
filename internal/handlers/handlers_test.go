package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andranikuz/gophkeeper/internal/handlers"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// fakeUserRepo реализует интерфейс repository.UserRepository.
type fakeUserRepo struct {
	GetUserByUsernameFunc func(username string) (*entity.User, error)
	SaveUserFunc          func(user entity.User) error
}

func (f *fakeUserRepo) GetUserByUsername(username string) (*entity.User, error) {
	if f.GetUserByUsernameFunc != nil {
		return f.GetUserByUsernameFunc(username)
	}
	return nil, nil
}

func (f *fakeUserRepo) SaveUser(user entity.User) error {
	if f.SaveUserFunc != nil {
		return f.SaveUserFunc(user)
	}
	return nil
}

// fakeAuthenticator реализует интерфейс services.AuthenticatorInterface.
type fakeAuthenticator struct {
	token string
}

func (fa *fakeAuthenticator) GenerateToken(userID, username string) (string, error) {
	return fa.token, nil
}

func (fa *fakeAuthenticator) HashPassword(password string) (string, error) {
	// Возвращаем фиктивный хэш
	return "hashed:" + password, nil
}

func (fa *fakeAuthenticator) CheckPasswordHash(password, hash string) bool {
	return hash == "hashed:"+password
}

func (fa *fakeAuthenticator) ValidateToken(tokenStr string) (*services.Claims, error) {
	// Минимальная реализация для тестов
	return &services.Claims{
		UserID:   "dummy",
		Username: "dummy",
	}, nil
}

func (fa *fakeAuthenticator) GetUserIdFromCtx(ctx context.Context) (string, error) {
	return "dummy", nil
}

func TestLogin_Success(t *testing.T) {
	// Подготавливаем тестового пользователя.
	plainPassword := "password123"
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	testUser := &entity.User{
		ID:        "user123",
		Username:  "testuser",
		Password:  string(hashed),
		CreatedAt: time.Now(),
	}

	// Фиктивный репозиторий возвращает testUser для "testuser".
	fakeRepo := &fakeUserRepo{
		GetUserByUsernameFunc: func(username string) (*entity.User, error) {
			if username == "testuser" {
				return testUser, nil
			}
			return nil, nil
		},
	}

	// Фиктивный аутентификатор возвращает фиксированный токен.
	fakeAuth := &fakeAuthenticator{token: "testtoken"}

	// Создаем Handler.
	h := handlers.NewHandler(nil, fakeRepo, fakeAuth)

	// Формируем JSON-запрос для логина.
	loginPayload := map[string]string{
		"username": "testuser",
		"password": plainPassword,
	}
	body, err := json.Marshal(loginPayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// Вызываем метод Login.
	h.Login(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var respData map[string]string
	err = json.NewDecoder(res.Body).Decode(&respData)
	require.NoError(t, err)

	assert.Equal(t, "testtoken", respData["token"])
	assert.Equal(t, "user123", respData["user_id"])
}

func TestLogin_InvalidCredentials(t *testing.T) {
	// Подготавливаем тестового пользователя.
	plainPassword := "password123"
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	testUser := &entity.User{
		ID:        "user123",
		Username:  "testuser",
		Password:  string(hashed),
		CreatedAt: time.Now(),
	}

	fakeRepo := &fakeUserRepo{
		GetUserByUsernameFunc: func(username string) (*entity.User, error) {
			if username == "testuser" {
				return testUser, nil
			}
			return nil, nil
		},
	}

	fakeAuth := &fakeAuthenticator{token: "testtoken"}
	h := handlers.NewHandler(nil, fakeRepo, fakeAuth)

	// Передаем неверный пароль.
	loginPayload := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	body, err := json.Marshal(loginPayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	// Ожидаем статус 401 Unauthorized.
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRegister_Success(t *testing.T) {
	var savedUser entity.User
	fakeRepo := &fakeUserRepo{
		SaveUserFunc: func(user entity.User) error {
			savedUser = user
			return nil
		},
	}

	fakeAuth := &fakeAuthenticator{token: "testtoken"}
	h := handlers.NewHandler(nil, fakeRepo, fakeAuth)

	// Формируем запрос на регистрацию.
	registerPayload := map[string]string{
		"username": "newuser",
		"password": "newpassword",
	}
	body, err := json.Marshal(registerPayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	// Ожидаем статус 201 Created.
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var respData map[string]string
	err = json.NewDecoder(res.Body).Decode(&respData)
	require.NoError(t, err)
	assert.Equal(t, "User registered successfully", respData["message"])

	// Проверяем, что пользователь сохранён и его поля заполнены корректно.
	assert.Equal(t, "newuser", savedUser.Username)
	assert.NotEqual(t, "newpassword", savedUser.Password) // пароль должен быть захеширован
	assert.NotEmpty(t, savedUser.ID)
}

func TestRegister_MissingFields(t *testing.T) {
	fakeRepo := &fakeUserRepo{}
	fakeAuth := &fakeAuthenticator{token: "testtoken"}
	h := handlers.NewHandler(nil, fakeRepo, fakeAuth)

	// Отсутствует username.
	registerPayload := map[string]string{
		"username": "",
		"password": "password",
	}
	body, err := json.Marshal(registerPayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	// Ожидаем статус 400 Bad Request.
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
