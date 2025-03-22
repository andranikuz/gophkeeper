package entity

import (
	"time"
)

// DataType определяет типы данных, сохраняемых в системе.
type DataType int

// UpdateContent обновляет содержимое DataItem и обновляет временную метку.
func (d DataType) String() string {
	switch d {
	case DataTypeCredential:
		return "Credential"
	case DataTypeText:
		return "Text"
	case DataTypeBinary:
		return "Binary"
	case DataTypeCard:
		return "Card"
	default:
		return "Unknown type"
	}
}

const (
	// DataTypeText используется для произвольного текстового контента.
	DataTypeText DataType = iota
	// DataTypeCredential используется для хранения пар логин/пароль.
	DataTypeCredential
	// DataTypeBinary используется для произвольных бинарных данных.
	DataTypeBinary
	// DataTypeCard используется для хранения данных банковских карт.
	DataTypeCard
)

// DataItem представляет единицу данных, которую можно хранить в системе.
type DataItem struct {
	ID        string    `json:"id"`         // Уникальный идентификатор данных (например, UUID)
	Type      DataType  `json:"type"`       // Тип данных
	Content   string    `json:"content"`    // Содержимое данных (например, зашифрованное)
	Meta      string    `json:"meta"`       // Произвольная метаинформация
	UserID    string    `json:"user_id"`    // Владелец записи
	UpdatedAt time.Time `json:"updated_at"` // Время последнего обновления (используется для синхронизации)
}

// NewDataItem создаёт новый экземпляр DataItem с заданными параметрами.
func NewDataItem(id string, dataType DataType, content string, meta string, userID string) *DataItem {
	return &DataItem{
		ID:        id,
		Type:      dataType,
		Content:   content,
		Meta:      meta,
		UserID:    userID,
		UpdatedAt: time.Now(),
	}
}
