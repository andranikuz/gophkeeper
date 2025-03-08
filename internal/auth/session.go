package auth

import (
	"encoding/json"
	"io/ioutil"
)

const sessionFile = "session.json"

// Token хранит JWT-токен.
type Token struct {
	Token string `json:"token"`
}

// Session управляет сессией пользователя.
type Session struct {
	token Token
}

func NewSession() *Session {
	return &Session{
		token: Token{
			Token: ReadSession(),
		},
	}
}

// SaveSession сохраняет токен в файл.
func (s *Session) SaveSession(token string) error {
	sess := Token{Token: token}
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(sessionFile, data, 0644)
}

// GetSessionToken возвращает токен сессии.
func (s *Session) GetSessionToken() string {
	return s.token.Token
}

// ReadSession считывает токен из файла.
func ReadSession() string {
	data, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		return ""
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return ""
	}
	return token.Token
}
