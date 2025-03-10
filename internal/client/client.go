package client

import (
	"github.com/andranikuz/gophkeeper/internal/auth"
	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

// Client представляет клиента для работы с сервером.
type Client struct {
	ServerURL  string
	Session    *auth.Session
	LocalDB    *BboltStorage
	grpcClient pb.FileSyncServiceClient
	grpcConn   *grpc.ClientConn
}

// NewClient создаёт новый экземпляр Client.
func NewClient(serverURL string, serverGrpcURL string, session *auth.Session, localDB *BboltStorage) *Client {
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
