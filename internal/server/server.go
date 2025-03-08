package server

import (
	"context"
	"database/sql"
	"github.com/andranikuz/gophkeeper/internal/auth"
	"github.com/andranikuz/gophkeeper/internal/config"
	"github.com/andranikuz/gophkeeper/internal/handlers"
	"github.com/andranikuz/gophkeeper/internal/sqlite"
	"net/http"
	"strconv"
	"time"
)

// Server реализует сервер.
type Server struct {
	handler *handlers.Handler
	cfg     config.Config
	ctx     context.Context
}

// NewServer создаёт новый экземпляр Server.
func NewServer(ctx context.Context) (*Server, error) {
	// Инициализируем конфиг.
	cfg := config.LoadConfig()
	// Инициализируем базу данных sqlite.
	db, err := InitDB(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	// Инициализируем репозитории.
	dataItemRepo, err := sqlite.NewDataItemRepository(db)
	if err != nil {
		return nil, err
	}
	userRepo, err := sqlite.NewUserRepository(db)
	if err != nil {
		return nil, err
	}
	// Инициализируем модуль аутентификации.
	authManager := auth.NewAuthenticator(cfg.TokenSecret, cfg.TokenExpiration)
	// Инициализируем http хендлеры.
	handler := handlers.NewHandler(dataItemRepo, userRepo, authManager)

	return &Server{
		ctx:     ctx,
		handler: handler,
	}, nil
}

func (s Server) Run() error {
	return http.ListenAndServe(s.cfg.Host+`:`+strconv.Itoa(s.cfg.Port), s.handler.RegisterRoutes())
}

// InitDB открывает базу SQLite по заданному пути и инициализирует схему.
func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// Устанавливаем параметры
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}
