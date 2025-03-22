package grpcserver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/logger"
)

// DownloadFile открывает файл по заданному ID и отправляет его чанками.
func (s *fileSyncServiceServer) DownloadFile(req *pb.FileDownloadRequest, stream pb.FileSyncService_DownloadFileServer) error {
	userID, err := s.authenticator.GetUserIdFromCtx(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	uploadDir := s.uploadDir + `/` + userID
	filePath := filepath.Join(uploadDir, req.Id)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	buf := make([]byte, 32*1024) // 32 KB чанки
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 {
			break
		}
		chunk := &pb.FileChunk{
			Id:        req.Id,
			ChunkData: buf[:n],
		}
		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
	}
	logger.InfoLogger.Printf("DownloadFile: file %s sent successfully", req.Id)
	return nil
}
