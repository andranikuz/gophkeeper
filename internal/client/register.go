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

// RegisterDTO представляет данные для регистрации.
type RegisterDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register отправляет запрос на регистрацию пользователя.
func (c *Client) Register(ctx context.Context, dto RegisterDTO) error {
	url := c.ServerURL + "/register"
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
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(body))
	}
	return nil
}
