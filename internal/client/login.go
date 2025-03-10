package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/andranikuz/gophkeeper/internal/auth"
	"io"
	"net/http"
	"time"
)

// LoginDTO представляет данные для логина.
type LoginDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login отправляет HTTP-запрос на логин и, при успешном ответе, сохраняет JWT-токен и userID в сессии.
func (c *Client) Login(ctx context.Context, dto LoginDTO) error {
	// Формируем URL для запроса логина.
	url := c.ServerURL + "/login"
	// Сериализуем DTO в JSON.
	data, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	// Создаем HTTP-запрос с заданным контекстом.
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Отправляем запрос с таймаутом.
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Если статус не OK, читаем тело ответа и возвращаем ошибку.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %s", string(body))
	}

	// Декодируем ответ в map[string]string.
	var res map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	// Извлекаем токен и userID из ответа.
	token, ok := res["token"]
	if !ok {
		return fmt.Errorf("token not found in response")
	}
	userID, ok := res["user_id"]
	if !ok {
		return fmt.Errorf("user_id not found in response")
	}

	// Сохраняем токен и userID в сессии.
	// Метод SaveSession принимает объект типа Token (с полями Token и UserID).
	if err := c.Session.Save(auth.Token{
		Token:  token,
		UserID: userID,
	}); err != nil {
		return err
	}

	return nil
}
