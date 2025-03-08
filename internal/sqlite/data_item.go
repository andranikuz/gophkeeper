package sqlite

import (
	"database/sql"
	"encoding/json"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DataItemRepository реализует хранилище данных с использованием SQLite.
type DataItemRepository struct {
	db *sql.DB
}

// NewDataItemRepository открывает (или создаёт) базу SQLite по заданному пути и инициализирует схему.
func NewDataItemRepository(db *sql.DB) (*DataItemRepository, error) {
	// Создаем таблицу, если ее еще нет.
	schema := `
	CREATE TABLE IF NOT EXISTS data_items (
		id TEXT PRIMARY KEY,
		type INTEGER,
		content BLOB,
		metadata TEXT,
		updated_at DATETIME
	);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return &DataItemRepository{db: db}, nil
}

// SaveItem сохраняет (или обновляет) один объект DataItem.
func (s *DataItemRepository) SaveItem(item entity.DataItem) error {
	// Сериализуем метаданные как JSON.
	meta, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}
	query := `
	INSERT OR REPLACE INTO data_items (id, type, content, metadata, updated_at)
	VALUES (?, ?, ?, ?, ?);
	`
	_, err = s.db.Exec(query, item.ID, int(item.Type), item.Content, string(meta), item.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return nil
}

// SaveItems сохраняет срез DataItem атомарно (в транзакции).
func (s *DataItemRepository) SaveItems(items []entity.DataItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
	INSERT OR REPLACE INTO data_items (id, type, content, metadata, updated_at)
	VALUES (?, ?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		meta, err := json.Marshal(item.Metadata)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = stmt.Exec(item.ID, int(item.Type), item.Content, string(meta), item.UpdatedAt.Format(time.RFC3339))
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetAllItems извлекает все объекты DataItem из базы.
func (s *DataItemRepository) GetAllItems() ([]entity.DataItem, error) {
	query := `
	SELECT id, type, content, metadata, updated_at FROM data_items;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.DataItem
	for rows.Next() {
		var item entity.DataItem
		var metaStr string
		var updatedAtStr string
		var typeInt int
		err := rows.Scan(&item.ID, &typeInt, &item.Content, &metaStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}
		item.Type = entity.DataType(typeInt)
		// Десериализуем метаданные.
		if err := json.Unmarshal([]byte(metaStr), &item.Metadata); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			item.UpdatedAt = time.Now()
		} else {
			item.UpdatedAt = t
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
