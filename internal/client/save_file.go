package client

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"

	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/utils"
)

// FileDTO представляет данные для отправки файла.
type FileDTO struct {
	FilePath string `json:"file_path"`
}

// SaveFile копирует исходный файл в директорию ./data/client_files с новым именем,
// а в объект DataItem сохраняет только базовое имя исходного файла в поле Content.
func (c *Client) SaveFile(ctx context.Context, dto FileDTO) error {
	// Генерируем новый UUID.
	id, err := uuid.NewV6()
	if err != nil {
		return err
	}
	// Создаем директорию, если ее не существует.
	if err := os.MkdirAll(utils.ClientDestDir, 0755); err != nil {
		return err
	}
	// Открываем исходный файл.
	srcFile, err := os.Open(dto.FilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Получаем базовое имя исходного файла (без пути).
	baseName := filepath.Base(dto.FilePath)
	item := entity.NewDataItem(
		id.String(),
		entity.DataTypeBinary,
		baseName,
		c.Session.GetUserID(),
	)

	// Создаем новый файл в директории назначения.
	dstFile, err := os.Create(utils.GetLocalFilePath(item))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Копируем содержимое файла.
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return c.LocalDB.SaveItem(item)
}
