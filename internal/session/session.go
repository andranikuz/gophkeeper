package session

import (
	"encoding/json"
	"os"

	"github.com/andranikuz/gophkeeper/internal/client"
)

// BboltStorage must implements LocalStorage interface
var _ client.SessionService = &Session{}

const sessionFile = "data/session.json"

// Session управляет сессией пользователя.
type Session struct {
	Token client.Token
}

// NewSession создаёт новую сессию, пытаясь прочитать токен из файла.
// Если файла нет, возвращается сессия с пустым токеном.
func NewSession() Session {
	return Session{Token: readToken()}
}

// Save сохраняет переданный токен в сессию и записывает его в файл.
func (s Session) Save(token client.Token) error {
	s.Token = token
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(sessionFile, data, 0644)
}

// GetSessionToken возвращает строку токена сессии.
func (s Session) GetSessionToken() string {
	return s.Token.Token
}

// GetUserID возвращает userID из сессии.
func (s Session) GetUserID() string {
	return s.Token.UserID
}

// readToken считывает токен из файла и возвращает его.
// Если чтение или парсинг не удался, возвращается пустой Token.
func readToken() client.Token {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return client.Token{}
	}
	var token client.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return client.Token{}
	}
	return token
}
