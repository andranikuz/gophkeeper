package entity

import "time"

// User представляет пользователя системы.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // Хэшированный пароль
	CreatedAt time.Time `json:"created_at"`
}
