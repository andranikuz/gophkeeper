package handlers

import (
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"github.com/andranikuz/gophkeeper/pkg/repository"
	"github.com/andranikuz/gophkeeper/pkg/services"

	"github.com/go-chi/chi/v5"
)

// Handler инкапсулирует хранилище данных и настройки авторизации.
type Handler struct {
	DataItemRepo  repository.DataItemRepository
	UserRepo      repository.UserRepository
	Authenticator services.AuthenticatorInterface
}

// NewHandler создаёт новый Handler.
func NewHandler(
	dataItemRepo repository.DataItemRepository,
	userRepo repository.UserRepository,
	authenticator services.AuthenticatorInterface,
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
