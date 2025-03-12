package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"github.com/andranikuz/gophkeeper/pkg/utils"
)

// SyncGRPC выполняет синхронизацию метаданных и файлов с сервером через gRPC.
func (c *Client) SyncGRPC(ctx context.Context) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.Session.GetSessionToken())
	// 1. Получаем локальные записи.
	localItems, err := c.LocalDB.GetAllItems()
	if err != nil {
		return fmt.Errorf("failed to get local items: %w", err)
	}

	// 2. Преобразуем записи в protobuf-формат.
	pbItems := dataItemsToProto(localItems)

	// 3. Формируем запрос на синхронизацию.
	syncReq := &pb.SyncRecordsRequest{Items: pbItems}
	resp, err := c.grpcClient.SyncRecords(ctx, syncReq)
	if err != nil {
		return fmt.Errorf("sync records error: %w", err)
	}

	// 4. Преобразуем объединённый список из ответа в []entity.DataItem и обновляем локальное хранилище.
	mergedItems := protoToDataItems(resp.MergedRecords)
	if err := c.LocalDB.SaveItems(mergedItems); err != nil {
		return fmt.Errorf("failed to update local DB: %w", err)
	}

	// 5. Обрабатываем списки для передачи файлов.
	uploadList := protoToDataItems(resp.UploadList)
	downloadList := protoToDataItems(resp.DownloadList)

	// Для загрузки файлов с клиента на сервер.
	var wg sync.WaitGroup
	for _, item := range uploadList {
		if item.Type == entity.DataTypeBinary {
			wg.Add(1)
			go func(fileID string) {
				defer wg.Done()
				// Получаем локальный путь к файлу по его ID
				localFilePath := utils.GetLocalFilePath(&item)
				if err := c.uploadFileGRPC(ctx, fileID, localFilePath); err != nil {
					logger.ErrorLogger.Printf("Error uploading file %s: %v", fileID, err)
				} else {
					logger.InfoLogger.Printf("File %s uploaded successfully", fileID)
				}
			}(item.ID)
		}
	}

	// Для скачивания файлов с сервера на клиента.
	for _, item := range downloadList {
		if item.Type == entity.DataTypeBinary {
			wg.Add(1)
			go func(item entity.DataItem) {
				defer wg.Done()
				// Скачиваем файл и сохраняем его локально.
				localFilePath, err := c.downloadFileGRPC(ctx, item)
				if err != nil {
					logger.ErrorLogger.Printf("Error downloading file %s: %v", item.ID, err)
					return
				}
				logger.InfoLogger.Printf("File %s downloaded successfully, saved at %s", item.ID, localFilePath)
				// При необходимости можно обновить запись в локальном хранилище с новым путем.
			}(item)
		}
	}
	wg.Wait()

	logger.InfoLogger.Printf("SyncGRPC: synchronization completed successfully")
	return nil
}

// uploadFileGRPC выполняет загрузку файла с клиента на сервер с использованием стриминга.
func (c *Client) uploadFileGRPC(ctx context.Context, fileID, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	stream, err := c.grpcClient.UploadFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to start upload stream: %w", err)
	}

	buf := make([]byte, 32*1024) // Чанк размером 32KB.
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 {
			break
		}
		chunk := &pb.FileChunk{
			Id:        fileID,
			ChunkData: buf[:n],
		}
		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close upload stream: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("upload failed: %s", resp.Message)
	}
	return nil
}

// downloadFileGRPC скачивает файл с сервера по ID с использованием стриминга и сохраняет его локально.
// Возвращает путь к сохраненному файлу.
func (c *Client) downloadFileGRPC(ctx context.Context, item entity.DataItem) (string, error) {
	req := &pb.FileDownloadRequest{Id: item.ID}
	stream, err := c.grpcClient.DownloadFile(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to start download stream: %w", err)
	}

	// Определяем локальный путь для сохранения файла.
	localFilePath := utils.GetLocalFilePath(&item)
	f, err := os.Create(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", localFilePath, err)
	}
	defer f.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to receive chunk: %w", err)
		}
		if _, err := f.Write(chunk.ChunkData); err != nil {
			return "", fmt.Errorf("failed to write chunk: %w", err)
		}
	}
	return localFilePath, nil
}

// dataItemsToProto преобразует срез entity.DataItem в срез pb.DataItem.
func dataItemsToProto(items []entity.DataItem) []*pb.DataItem {
	var pbItems []*pb.DataItem
	for _, item := range items {
		pbItems = append(pbItems, &pb.DataItem{
			Id:        item.ID,
			Type:      int32(item.Type),
			Content:   item.Content,
			UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
		})
	}
	return pbItems
}

// protoToDataItems преобразует срез pb.DataItem в срез entity.DataItem.
func protoToDataItems(pbItems []*pb.DataItem) []entity.DataItem {
	var items []entity.DataItem
	for _, pbItem := range pbItems {
		t, _ := time.Parse(time.RFC3339, pbItem.UpdatedAt)
		items = append(items, entity.DataItem{
			ID:        pbItem.Id,
			Type:      entity.DataType(pbItem.Type),
			Content:   pbItem.Content,
			UpdatedAt: t,
		})
	}
	return items
}
