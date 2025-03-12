package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"github.com/andranikuz/gophkeeper/internal/auth"
	"github.com/andranikuz/gophkeeper/internal/config"
	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/internal/grpcserver"
	"github.com/andranikuz/gophkeeper/internal/handlers"
	"github.com/andranikuz/gophkeeper/internal/sqlite"
	"github.com/andranikuz/gophkeeper/pkg/logger"
)

// Server реализует сервер.
type Server struct {
	handler    *handlers.Handler
	grpcServer *grpc.Server
	cfg        *config.Config
	ctx        context.Context
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
	// Создаем gRPC сервер с интерсепторами авторизации.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.JwtUnaryInterceptor(authManager)),
		grpc.StreamInterceptor(grpcserver.JwtStreamInterceptor(authManager)),
	)
	fileSyncSvc := grpcserver.NewFileSyncServiceServer("./data/server_files", dataItemRepo, authManager)
	pb.RegisterFileSyncServiceServer(grpcServer, fileSyncSvc)

	return &Server{
		ctx:        ctx,
		cfg:        cfg,
		handler:    handler,
		grpcServer: grpcServer,
	}, nil
}

func (s Server) Run() error {
	// Формируем адрес для HTTP-сервера.
	httpAddr := s.cfg.Host + ":" + strconv.Itoa(s.cfg.Port)
	// Запускаем HTTP-сервер в горутине.
	go func() {
		logger.InfoLogger.Printf("HTTP server started on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, s.handler.RegisterRoutes()); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Формируем адрес для gRPC-сервера.
	grpcAddr := s.cfg.Host + ":" + strconv.Itoa(s.cfg.GrpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listener error: %w", err)
	}
	logger.InfoLogger.Printf("gRPC server started on %s", grpcAddr)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}
	return nil
}

// InitDB открывает базу SQLite по заданному пути.
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
