package repository

import "github.com/andranikuz/gophkeeper/pkg/entity"

// DataItemRepository описывает набор методов для работы с данными типа DataItem.
type DataItemRepository interface {
	// SaveItems сохраняет срез DataItem атомарно (в транзакции).
	SaveItems(items []entity.DataItem) error
	// GetUserItems извлекает все объекты DataItem для заданного пользователя.
	GetUserItems(userID string) ([]entity.DataItem, error)
}
