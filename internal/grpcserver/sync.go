package grpcserver

import (
	"context"
	"fmt"
	"time"

	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/logger"
)

// SyncRecords принимает от клиента массив записей, объединяет их с данными из хранилища,
// определяет какие файлы нужно загрузить с клиента и какие скачать с сервера,
// сохраняет объединённый список в хранилище и возвращает его вместе с массивами для загрузки.
func (s *fileSyncServiceServer) SyncRecords(ctx context.Context, req *pb.SyncRecordsRequest) (*pb.SyncRecordsResponse, error) {
	// 1. Извлекаем записи от клиента и записи с сервера.
	clientItems, serverItems, err := s.extractClientAndServerItems(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to extract items: %w", err)
	}

	// 2. Объединяем записи.
	mergedItems := mergeDataItems(serverItems, clientItems)

	// 3. Вычисляем списки файлов для загрузки/скачивания.
	uploadList, downloadList := computeSyncLists(clientItems, serverItems)

	// 4. Сохраняем объединённый список в БД.
	if err := s.dataItemRepository.SaveItems(mergedItems); err != nil {
		return nil, fmt.Errorf("failed to save merged items: %w", err)
	}

	// 5. Формируем и возвращаем ответ.
	resp := &pb.SyncRecordsResponse{
		UploadList:    dataItemsToProto(uploadList),
		DownloadList:  dataItemsToProto(downloadList),
		MergedRecords: dataItemsToProto(mergedItems),
	}

	logger.InfoLogger.Printf("SyncRecords: merged %d items; upload: %d, download: %d",
		len(mergedItems), len(uploadList), len(downloadList))
	return resp, nil
}

// extractClientAndServerItems извлекает userID из контекста, преобразует записи, полученные от клиента,
// и получает записи пользователя из БД.
func (s *fileSyncServiceServer) extractClientAndServerItems(ctx context.Context, req *pb.SyncRecordsRequest) (clientItems, serverItems []entity.DataItem, err error) {
	// Извлекаем userID из контекста.
	userID, err := s.authenticator.GetUserIdFromCtx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get userID: %w", err)
	}

	// Преобразуем записи, полученные от клиента, в объекты entity.DataItem, устанавливая userID.
	clientItems = protoToDataItems(req.Items, userID)

	// Получаем серверные записи для этого пользователя.
	serverItems, err = s.dataItemRepository.GetUserItems(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get server items: %w", err)
	}
	return clientItems, serverItems, nil
}

// computeSyncLists вычисляет, какие файлы нужно загрузить с клиента (uploadList)
// и какие файлы нужно скачать с сервера (downloadList) на основе сравнений записей.
func computeSyncLists(clientItems, serverItems []entity.DataItem) (uploadList, downloadList []entity.DataItem) {
	// Создаем маппы для быстрого поиска.
	clientMap := make(map[string]entity.DataItem)
	serverMap := make(map[string]entity.DataItem)
	for _, item := range clientItems {
		clientMap[item.ID] = item
	}
	for _, item := range serverItems {
		serverMap[item.ID] = item
	}

	// Если запись есть у клиента и либо отсутствует на сервере, либо версия от клиента новее – upload.
	for id, cItem := range clientMap {
		if sItem, ok := serverMap[id]; !ok || cItem.UpdatedAt.After(sItem.UpdatedAt) {
			uploadList = append(uploadList, cItem)
		}
	}
	// Если запись есть на сервере и либо отсутствует у клиента, либо серверная версия новее – download.
	for id, sItem := range serverMap {
		if cItem, ok := clientMap[id]; !ok || sItem.UpdatedAt.After(cItem.UpdatedAt) {
			downloadList = append(downloadList, sItem)
		}
	}
	return uploadList, downloadList
}

// --- Вспомогательные функции для конвертации между protobuf и внутренней моделью ---

// Преобразует срез protobuf DataItem в срез storage.DataItem.
func protoToDataItems(pbItems []*pb.DataItem, userID string) []entity.DataItem {
	var items []entity.DataItem
	for _, pbItem := range pbItems {
		t, _ := time.Parse(time.RFC3339, pbItem.UpdatedAt)
		items = append(items, entity.DataItem{
			ID:        pbItem.Id,
			Type:      entity.DataType(pbItem.Type),
			Content:   pbItem.Content,
			UserID:    userID,
			UpdatedAt: t,
		})
	}
	return items
}

// Преобразует срез storage.DataItem в срез protobuf DataItem.
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

// mergeDataItems объединяет два среза записей по принципу "последнее обновление выигрывает".
func mergeDataItems(serverItems, clientItems []entity.DataItem) []entity.DataItem {
	mergedMap := make(map[string]entity.DataItem)
	// Сначала кладем все серверные записи.
	for _, item := range serverItems {
		mergedMap[item.ID] = item
	}
	// Затем перебираем клиентские записи.
	for _, item := range clientItems {
		if exist, ok := mergedMap[item.ID]; ok {
			if item.UpdatedAt.After(exist.UpdatedAt) {
				mergedMap[item.ID] = item
			}
		} else {
			mergedMap[item.ID] = item
		}
	}
	var merged []entity.DataItem
	for _, item := range mergedMap {
		merged = append(merged, item)
	}
	return merged
}
