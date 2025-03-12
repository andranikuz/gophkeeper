package client

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/andranikuz/gophkeeper/pkg/entity"
)

// TextDTO представляет данные для типа "text".
type TextDTO struct {
	Text string `json:"text"`
	Meta string `json:"meta"`
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
		dto.Meta,
		c.Session.GetUserID(),
	))
}
