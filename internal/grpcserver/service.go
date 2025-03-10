package grpcserver

import (
	"github.com/andranikuz/gophkeeper/internal/auth"
	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/internal/sqlite"
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"os"
)

// fileSyncServiceServer реализует pb.FileSyncServiceServer.
type fileSyncServiceServer struct {
	pb.UnimplementedFileSyncServiceServer
	uploadDir          string                     // Директория для хранения файлов
	dataItemRepository *sqlite.DataItemRepository // Репозиторий data_item
	authenticator      *auth.Authenticator        // Сервис авторизации
}

// NewFileSyncServiceServer создаёт новый экземпляр сервиса.
func NewFileSyncServiceServer(
	uploadDir string,
	dataItemRepository *sqlite.DataItemRepository,
	authenticator *auth.Authenticator,
) pb.FileSyncServiceServer {
	// Создаем директорию для загрузок, если её нет.
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.ErrorLogger.Printf("Failed to create upload directory: %v", err)
	}
	return &fileSyncServiceServer{
		uploadDir:          uploadDir,
		dataItemRepository: dataItemRepository,
		authenticator:      authenticator,
	}
}
