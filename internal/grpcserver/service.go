package grpcserver

import (
	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"github.com/andranikuz/gophkeeper/pkg/repository"
	"github.com/andranikuz/gophkeeper/pkg/services"
	"os"
)

// fileSyncServiceServer реализует pb.FileSyncServiceServer.
type fileSyncServiceServer struct {
	pb.UnimplementedFileSyncServiceServer
	uploadDir          string                          // Директория для хранения файлов
	dataItemRepository repository.DataItemRepository   // Репозиторий data_item
	authenticator      services.AuthenticatorInterface // Сервис авторизации
}

// NewFileSyncServiceServer создаёт новый экземпляр сервиса.
func NewFileSyncServiceServer(
	uploadDir string,
	dataItemRepository repository.DataItemRepository,
	authenticator services.AuthenticatorInterface,
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
