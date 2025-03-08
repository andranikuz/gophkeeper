package client

import (
	"github.com/andranikuz/gophkeeper/internal/auth"
)

// Client представляет клиента для работы с сервером.
type Client struct {
	ServerURL string
	Session   *auth.Session
	LocalDB   *BboltStorage
}

// NewClient создаёт новый экземпляр Client.
func NewClient(serverURL string, session *auth.Session, localDB *BboltStorage) *Client {
	return &Client{
		ServerURL: serverURL,
		Session:   session,
		LocalDB:   localDB,
	}
}
