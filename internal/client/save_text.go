package client

import (
	"context"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/gofrs/uuid"
	"time"
)

// TextDTO представляет данные для типа "text".
type TextDTO struct {
	Text string `json:"text"`
}

// SaveText сохраняет данные типа "text" в локальное хранилище.
func (c *Client) SaveText(ctx context.Context, dto TextDTO) error {
	id, err := uuid.NewV6()
	if err != nil {
		return err
	}
	dataItem := entity.DataItem{
		ID:        id.String(),
		Type:      entity.DataTypeText,
		Content:   []byte(dto.Text),
		UpdatedAt: time.Now(),
	}
	return c.LocalDB.SaveItem(dataItem)
}
