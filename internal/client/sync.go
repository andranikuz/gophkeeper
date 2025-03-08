package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"io"
	"net/http"
	"time"
)

// Sync отправляет локальные данные на сервер для синхронизации,
// затем получает объединённый список данных и обновляет локальное хранилище.
func (c *Client) Sync(ctx context.Context) error {
	// 1. Получаем все локальные данные.
	localItems, err := c.LocalDB.GetAllItems()
	if err != nil {
		return fmt.Errorf("failed to get local items: %w", err)
	}

	// 2. Отправляем их на сервер.
	payload, err := json.Marshal(localItems)
	if err != nil {
		return fmt.Errorf("failed to marshal local items: %w", err)
	}
	url := c.ServerURL + "/sync"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Session.GetSessionToken())
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sync request error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sync failed: %s", string(body))
	}

	// 3. Читаем объединённый ответ.
	var response struct {
		User string            `json:"user"`
		Data []entity.DataItem `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode sync response: %w", err)
	}

	// 4. Обновляем локальное хранилище: сохраняем каждый элемент.
	if err := c.LocalDB.SaveItems(response.Data); err != nil {
		return err
	}

	return nil
}
