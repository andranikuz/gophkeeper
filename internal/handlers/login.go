package handlers

import (
	"encoding/json"
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// Login реализует аутентификацию пользователя.
// Здесь должна быть логика проверки учетных данных и генерации JWT.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Получаем пользователя из базы.
	user, err := h.UserRepo.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Сравниваем хэшированный пароль.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT-токен
	token, err := h.Authenticator.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	logger.InfoLogger.Printf("User logged in: %s", user.Username)
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
