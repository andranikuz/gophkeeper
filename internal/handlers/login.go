package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/andranikuz/gophkeeper/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// Login реализует аутентификацию пользователя.
// При успешном логине генерируется JWT-токен и возвращается вместе с userID.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// Декодируем JSON-пейлоад запроса, содержащий имя пользователя и пароль.
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// Проверяем, что оба поля не пусты.
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Получаем пользователя из базы по имени.
	user, err := h.UserRepo.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Сравниваем хэшированный пароль пользователя с введённым.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT-токен с информацией о пользователе.
	// Токен содержит user.ID, user.Username и срок действия (например, 24 часа).
	token, err := h.Authenticator.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	// Логируем успешный вход пользователя.
	logger.InfoLogger.Printf("User logged in: %s", user.Username)

	// Отправляем клиенту JSON-ответ с токеном и userID.
	json.NewEncoder(w).Encode(map[string]string{
		"token":   token,
		"user_id": user.ID,
	})
}
