package session

import (
	"encoding/json"
	"github.com/andranikuz/gophkeeper/internal/client"
	"os"
	"testing"
)

// TestSaveAndReadToken проверяет сохранение токена в файл и его последующее считывание.
func TestSaveAndReadToken(t *testing.T) {
	// Создаём временную директорию
	dir, err := os.MkdirTemp("", "testsession")
	if err != nil {
		t.Fatalf("Не удалось создать временную директорию: %v", err)
	}
	defer os.RemoveAll(dir)

	// Меняем текущую рабочую директорию на временную
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Не удалось получить текущую директорию: %v", err)
	}
	defer os.Chdir(oldWD)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Не удалось перейти в временную директорию: %v", err)
	}

	// Создаём директорию "data", если она отсутствует
	dataDir := "data"
	if err := os.Mkdir(dataDir, 0755); err != nil && !os.IsExist(err) {
		t.Fatalf("Не удалось создать директорию %s: %v", dataDir, err)
	}

	// Создаём новую сессию и сохраняем токен
	session := NewSession()
	token := client.Token{
		Token:  "testToken",
		UserID: "userID123",
	}
	if err := session.Save(token); err != nil {
		t.Fatalf("Ошибка сохранения токена: %v", err)
	}

	// Проверяем, что файл был создан и содержит корректные данные
	data, err := os.ReadFile("data/session.json")
	if err != nil {
		t.Fatalf("Не удалось прочитать файл сессии: %v", err)
	}

	var savedToken client.Token
	if err := json.Unmarshal(data, &savedToken); err != nil {
		t.Fatalf("Ошибка при разборе JSON: %v", err)
	}
	if savedToken != token {
		t.Errorf("Ожидался токен %+v, получен %+v", token, savedToken)
	}

	// Создаём новую сессию, чтобы проверить, что токен считывается из файла
	newSession := NewSession()
	if newSession.GetSessionToken() != token.Token || newSession.GetUserID() != token.UserID {
		t.Errorf("Ожидался токен %+v, получена сессия %+v", token, newSession.Token)
	}
}
