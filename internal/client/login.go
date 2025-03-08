package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LoginDTO представляет данные для логина.
type LoginDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login отправляет запрос на логин и, при успешном входе, сохраняет JWT-токен в структуре.
func (c *Client) Login(ctx context.Context, dto LoginDTO) error {
	url := c.ServerURL + "/login"
	data, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %s", string(body))
	}
	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	token, ok := res["token"].(string)
	if !ok {
		return fmt.Errorf("token not found in response")
	}
	if err := c.Session.SaveSession(token); err != nil {
		return err
	}

	return nil
}
