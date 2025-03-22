package repository

import "github.com/andranikuz/gophkeeper/pkg/entity"

// UserRepository определяет операции для работы с пользователями в базе SQLite.
type UserRepository interface {
	// SaveUser сохраняет или обновляет пользователя в базе.
	SaveUser(user entity.User) error
	// GetUserByUsername возвращает пользователя по имени.
	GetUserByUsername(username string) (*entity.User, error)
}
