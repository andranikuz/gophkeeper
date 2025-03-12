package client

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/entity"
)

// Client представляет клиента для работы с сервером.
type Client struct {
	ServerURL  string
	Session    SessionService
	LocalDB    LocalStorage
	grpcClient pb.FileSyncServiceClient
	grpcConn   *grpc.ClientConn
}

// NewClient создаёт новый экземпляр Client.
func NewClient(serverURL string, serverGrpcURL string, session SessionService, localDB LocalStorage) *Client {
	// Устанавливаем gRPC-соединение. Используем insecure для примера.
	conn, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server at %s: %v", serverGrpcURL, err)
	}
	grpcClient := pb.NewFileSyncServiceClient(conn)

	return &Client{
		ServerURL:  serverURL,
		Session:    session,
		LocalDB:    localDB,
		grpcClient: grpcClient,
		grpcConn:   conn,
	}
}

type LocalStorage interface {
	SaveItems(items []entity.DataItem) error
	SaveItem(item *entity.DataItem) error
	GetAllItems() ([]entity.DataItem, error)
	DeleteItem(id string) error
	GetByID(id string) (*entity.DataItem, error)
}

// Token хранит JWT-токен и идентификатор пользователя.
type Token struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

type SessionService interface {
	Save(token Token) error
	GetSessionToken() string
	GetUserID() string
}
