package sqlite

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/andranikuz/gophkeeper/pkg/entity"
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
		content text,
		meta text,
		user_id text,
		updated_at DATETIME
	);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return &DataItemRepository{db: db}, nil
}

// SaveItems сохраняет срез DataItem атомарно (в транзакции).
func (s *DataItemRepository) SaveItems(items []entity.DataItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
	INSERT OR REPLACE INTO data_items (id, type, content, meta, user_id, updated_at)
	VALUES (?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(item.ID, int(item.Type), item.Content, item.Meta, item.UserID, item.UpdatedAt.Format(time.RFC3339))
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetUserItems извлекает все объекты пользователя DataItem из базы.
func (s *DataItemRepository) GetUserItems(userID string) ([]entity.DataItem, error) {
	query := `
	SELECT id, type, content, meta, user_id, updated_at 
	FROM data_items
	WHERE user_id = ?;
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.DataItem
	for rows.Next() {
		var item entity.DataItem
		var updatedAtStr string
		var typeInt int
		err := rows.Scan(&item.ID, &typeInt, &item.Content, &item.Meta, &item.UserID, &updatedAtStr)
		if err != nil {
			return nil, err
		}
		item.Type = entity.DataType(typeInt)
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
