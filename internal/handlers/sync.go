package handlers

import (
	"encoding/json"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"net/http"
)

// Sync объединяет данные от клиента с данными сервера и сохраняет результат.
func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	var clientItems []entity.DataItem
	if err := json.NewDecoder(r.Body).Decode(&clientItems); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	serverItems, err := h.DataItemRepo.GetAllItems()
	if err != nil {
		http.Error(w, "Failed to get server data", http.StatusInternalServerError)
		return
	}

	mergedItems := h.mergeDataItems(serverItems, clientItems)
	if err := h.DataItemRepo.SaveItems(mergedItems); err != nil {
		http.Error(w, "Failed to save merged data", http.StatusInternalServerError)
		return
	}
	logger.InfoLogger.Println("Sync: data synchronized")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"merged": mergedItems,
	})
}

// mergeDataItems объединяет записи по принципу "последнее обновление выигрывает".
func (h *Handler) mergeDataItems(serverItems, clientItems []entity.DataItem) []entity.DataItem {
	mergedMap := make(map[string]entity.DataItem)
	for _, item := range serverItems {
		mergedMap[item.ID] = item
	}
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
