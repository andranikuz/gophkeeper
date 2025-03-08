package sqlite

import (
	"database/sql"
	"time"

	"github.com/andranikuz/gophkeeper/pkg/entity"
)

// UserRepository реализует операции для работы с пользователями в базе SQLite.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository открывает (или создаёт) таблицу пользователей и возвращает репозиторий.
func NewUserRepository(db *sql.DB) (*UserRepository, error) {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE,
		password TEXT,
		created_at DATETIME
	);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return nil, err
	}
	return &UserRepository{db: db}, nil
}

// SaveUser сохраняет или обновляет пользователя в базе.
func (r *UserRepository) SaveUser(user entity.User) error {
	query := `
	INSERT OR REPLACE INTO users (id, username, password, created_at)
	VALUES (?, ?, ?, ?);
	`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Password, user.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return nil
}

// GetUserByUsername возвращает пользователя по имени.
func (r *UserRepository) GetUserByUsername(username string) (*entity.User, error) {
	query := `SELECT id, username, password, created_at FROM users WHERE username = ?;`
	row := r.db.QueryRow(query, username)

	var user entity.User
	var createdAtStr string
	err := row.Scan(&user.ID, &user.Username, &user.Password, &createdAtStr)
	if err != nil {
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		user.CreatedAt = time.Now()
	} else {
		user.CreatedAt = t
	}
	return &user, nil
}
