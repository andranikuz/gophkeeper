package client

import (
	"context"
	"encoding/json"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/gofrs/uuid"
)

// CredentialDTO представляет данные для типа "credential".
type CredentialDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// SaveCredential сохраняет данные типа "credential" в локальное хранилище.
func (c *Client) SaveCredential(ctx context.Context, dto CredentialDTO) error {
	payload, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	id, err := uuid.NewV6()
	if err != nil {
		return err
	}
	return c.LocalDB.SaveItem(entity.NewDataItem(
		id.String(),
		entity.DataTypeCredential,
		string(payload),
		c.Session.GetUserID(),
	))
}
