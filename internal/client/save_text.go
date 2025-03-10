package client

import (
	"context"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/gofrs/uuid"
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
	return c.LocalDB.SaveItem(entity.NewDataItem(
		id.String(),
		entity.DataTypeText,
		dto.Text,
		c.Session.GetUserID(),
	))
}
