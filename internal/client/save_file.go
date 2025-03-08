package client

import (
	"context"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/gofrs/uuid"
	"io"
	"os"
	"path/filepath"
	"time"
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

	// Определяем расширение исходного файла.
	ext := filepath.Ext(dto.FilePath)

	// Директория для хранения скопированных файлов.
	destDir := "./data/client_files"
	// Создаем директорию, если ее не существует.
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Формируем новый путь для копирования: destDir + "/" + id + ext.
	destPath := filepath.Join(destDir, id.String()+ext)

	// Открываем исходный файл.
	srcFile, err := os.Open(dto.FilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Создаем новый файл в директории назначения.
	dstFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Копируем содержимое файла.
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Получаем базовое имя исходного файла (без пути).
	baseName := filepath.Base(dto.FilePath)

	// Создаем объект DataItem, в котором в Content записано только базовое имя исходного файла.
	dataItem := entity.DataItem{
		ID:        id.String(),
		Type:      entity.DataTypeBinary,
		Content:   []byte(baseName),
		UpdatedAt: time.Now(),
	}

	return c.LocalDB.SaveItem(dataItem)
}
