package client

import (
	"context"
	"fmt"
	"os"

	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/logger"
	"github.com/andranikuz/gophkeeper/pkg/utils"
)

// DeleteItem удаляет запись по указанному id из локального хранилища.
// Если тип записи указывает на файл, также удаляет файл из файловой системы.
// Предполагается, что в поле Content записи хранится базовое имя файла.
func (c *Client) DeleteItem(ctx context.Context, id string) error {
	// Получаем запись из локального хранилища.
	item, err := c.LocalDB.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get item by ID: %w", err)
	}

	// Если запись относится к файлам, удаляем файл из файловой системы.
	if item.Type == entity.DataTypeBinary {
		// Предполагается, что файлы хранятся в директории, которую возвращает GetLocalFilePath.
		// Здесь, например, мы объединяем директорию хранения и базовое имя файла.
		filePath := utils.GetLocalFilePath(item)
		if err := os.Remove(filePath); err != nil {
			logger.ErrorLogger.Printf("failed to remove file from disk (%s): %s", filePath, err.Error())
		}
	}

	// Удаляем запись из локального хранилища.
	return c.LocalDB.DeleteItem(id)
}
