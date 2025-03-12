package client

import (
	"context"

	"github.com/andranikuz/gophkeeper/pkg/entity"
)

// GetItems возвращает данные из локального хранилища.
func (c *Client) GetItems(ctx context.Context) ([]entity.DataItem, error) {
	items, err := c.LocalDB.GetAllItems()
	if err != nil {
		return nil, err
	}
	return items, nil
}
