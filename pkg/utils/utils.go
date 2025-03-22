package utils

import (
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"path/filepath"
)

const ClientDestDir = "./data/client_files"

// GetLocalFilePath определяется путь к локальному файлу.
func GetLocalFilePath(item *entity.DataItem) string {
	// Определяем расширение файла.
	ext := filepath.Ext(item.Content)
	return ClientDestDir + `/` + item.ID + ext
}
