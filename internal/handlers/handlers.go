package handlers

import (
	"github.com/andranikuz/gophkeeper/internal/auth"
	"github.com/andranikuz/gophkeeper/internal/sqlite"
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"github.com/go-chi/chi/v5"
)

// Handler инкапсулирует хранилище данных и настройки авторизации.
type Handler struct {
	DataItemRepo  *sqlite.DataItemRepository
	UserRepo      *sqlite.UserRepository
	Authenticator *auth.Authenticator
}

// NewHandler создаёт новый Handler.
func NewHandler(
	dataItemRepo *sqlite.DataItemRepository,
	userRepo *sqlite.UserRepository,
	authenticator *auth.Authenticator,
) *Handler {
	return &Handler{
		DataItemRepo:  dataItemRepo,
		UserRepo:      userRepo,
		Authenticator: authenticator,
	}
}

// RegisterRoutes регистрирует маршруты с использованием chi и применяет middleware авторизации.
func (h *Handler) RegisterRoutes() chi.Router {
	r := chi.NewRouter()
	// Публичные маршруты.
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	logger.InfoLogger.Println("Routes registered successfully")

	return r
}
