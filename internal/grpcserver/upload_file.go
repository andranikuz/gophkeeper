package grpcserver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/logger"
)

// UploadFile принимает поток чанков файла от клиента и сохраняет файл.
func (s *fileSyncServiceServer) UploadFile(stream pb.FileSyncService_UploadFileServer) error {
	userID, err := s.authenticator.GetUserIdFromCtx(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	var fileID string
	// Создаем временный файл.
	tmpFile, err := os.CreateTemp(s.uploadDir, "upload-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	logger.InfoLogger.Printf("UploadFile: receiving file into %s", tmpFile.Name())
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// При первом чанке сохраняем fileID.
		if fileID == "" {
			fileID = chunk.Id
		}
		if _, err := tmpFile.Write(chunk.ChunkData); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
	}
	// Закрываем файл перед переименованием.
	if err := tmpFile.Close(); err != nil {
		return err
	}

	uploadDir := s.uploadDir + `/` + userID
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.ErrorLogger.Printf("Failed to create upload directory: %v", err)
	}

	// Переименовываем файл в окончательное имя.
	finalPath := filepath.Join(uploadDir, fileID)
	if err := os.Rename(tmpFile.Name(), finalPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	logger.InfoLogger.Printf("UploadFile: file %s saved as %s", fileID, finalPath)

	resp := &pb.FileUploadResponse{
		Id:      fileID,
		Success: true,
		Message: fmt.Sprintf("File uploaded successfully, saved at %s", finalPath),
	}
	return stream.SendAndClose(resp)
}
