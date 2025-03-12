package grpcserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/metadata"

	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------------------
// Фиктивный авторизатор
// -------------------------
type fakeAuthenticator struct {
	userID string
	err    error
}

// GetUserIdFromCtx возвращает заданный userID или ошибку.
func (fa *fakeAuthenticator) GetUserIdFromCtx(ctx context.Context) (string, error) {
	return fa.userID, fa.err
}

// HashPassword возвращает фиктивный хэш, добавляя префикс "hashed:" к паролю.
func (fa *fakeAuthenticator) HashPassword(password string) (string, error) {
	return "hashed:" + password, nil
}

// CheckPasswordHash возвращает true, если хэш совпадает с "hashed:"+password.
func (fa *fakeAuthenticator) CheckPasswordHash(password, hash string) bool {
	return hash == "hashed:"+password
}

// GenerateToken возвращает фиктивный токен в формате "token:<userID>:<username>".
func (fa *fakeAuthenticator) GenerateToken(userID, username string) (string, error) {
	return fmt.Sprintf("token:%s:%s", userID, username), nil
}

// ValidateToken разбирает фиктивный токен и возвращает Claims, если токен корректный.
func (fa *fakeAuthenticator) ValidateToken(tokenStr string) (*services.Claims, error) {
	parts := strings.Split(tokenStr, ":")
	if len(parts) != 3 || parts[0] != "token" {
		return nil, errors.New("invalid token")
	}
	return &services.Claims{
		UserID:   parts[1],
		Username: parts[2],
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "fake",
		},
	}, nil
}

// -------------------------
// Фиктивный репозиторий (для SyncRecords)
// -------------------------
type fakeRepository struct {
	getUserItemsFunc func(userID string) ([]entity.DataItem, error)
	saveItemsFunc    func(items []entity.DataItem) error
}

func (fr *fakeRepository) GetUserItems(userID string) ([]entity.DataItem, error) {
	if fr.getUserItemsFunc != nil {
		return fr.getUserItemsFunc(userID)
	}
	return nil, nil
}

func (fr *fakeRepository) SaveItems(items []entity.DataItem) error {
	if fr.saveItemsFunc != nil {
		return fr.saveItemsFunc(items)
	}
	return nil
}

// -------------------------
// Фиктивный grpc‑стрим для DownloadFile
// -------------------------
type fakeDownloadStream struct {
	chunks  []*pb.FileChunk
	ctx     context.Context
	sendErr error
}

func (s *fakeDownloadStream) Send(chunk *pb.FileChunk) error {
	if s.sendErr != nil {
		return s.sendErr
	}
	s.chunks = append(s.chunks, chunk)
	return nil
}

func (s *fakeDownloadStream) Context() context.Context {
	return s.ctx
}

func (s *fakeDownloadStream) SetHeader(md metadata.MD) error {
	return nil
}

func (s *fakeDownloadStream) SendHeader(md metadata.MD) error {
	return nil
}

func (s *fakeDownloadStream) SetTrailer(md metadata.MD) {
	// no-op
}

func (s *fakeDownloadStream) SendMsg(m interface{}) error {
	// no-op
	return nil
}

func (s *fakeDownloadStream) RecvMsg(m interface{}) error {
	// no-op
	return nil
}

// Если интерфейс требует метод Trailer, реализуем его:
func (s *fakeDownloadStream) Trailer() metadata.MD {
	return metadata.MD{}
}

// -------------------------
// Фиктивный grpc‑стрим для UploadFile
// -------------------------
type fakeUploadStream struct {
	chunks []*pb.FileChunk // Чанки, которые вернёт метод Recv.
	index  int
	ctx    context.Context
	resp   *pb.FileUploadResponse
}

func (s *fakeUploadStream) Recv() (*pb.FileChunk, error) {
	if s.index >= len(s.chunks) {
		return nil, io.EOF
	}
	chunk := s.chunks[s.index]
	s.index++
	return chunk, nil
}

func (s *fakeUploadStream) SendAndClose(resp *pb.FileUploadResponse) error {
	s.resp = resp
	return nil
}

func (s *fakeUploadStream) Context() context.Context {
	return s.ctx
}

func (s *fakeUploadStream) SetHeader(md metadata.MD) error {
	return nil
}

func (s *fakeUploadStream) SendHeader(md metadata.MD) error {
	return nil
}

// Теперь метод SetTrailer не возвращает ошибку.
func (s *fakeUploadStream) SetTrailer(md metadata.MD) {
	// no-op
}

func (s *fakeUploadStream) RecvMsg(m interface{}) error {
	return nil
}

func (s *fakeUploadStream) SendMsg(m interface{}) error {
	return nil
}

// -------------------------
// Тест для DownloadFile
// -------------------------
func TestDownloadFile_Success(t *testing.T) {
	// Создаем временную директорию для загрузок
	tempUploadDir, err := os.MkdirTemp("", "uploadDir")
	require.NoError(t, err)
	defer os.RemoveAll(tempUploadDir)

	// Определяем userID и файл, который будем отдавать.
	userID := "testuser"
	fileID := "file123"
	fileContent := "This is a test file content."
	// Путь, по которому метод будет искать файл: uploadDir + "/" + userID + "/" + fileID
	userUploadDir := filepath.Join(tempUploadDir, userID)
	require.NoError(t, os.MkdirAll(userUploadDir, 0755))
	filePath := filepath.Join(userUploadDir, fileID)
	err = os.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err)

	// Фиктивный авторизатор, возвращающий userID
	auth := &fakeAuthenticator{userID: userID}

	// Создаем сервер с нужной uploadDir и фиктивным авторизатором.
	srv := &fileSyncServiceServer{
		uploadDir:     tempUploadDir,
		authenticator: auth,
		// Репозиторий не используется в DownloadFile, можно оставить nil.
	}

	// Создаем fakeDownloadStream с контекстом.
	fakeStream := &fakeDownloadStream{
		ctx: context.Background(),
	}

	// Формируем запрос с fileID.
	req := &pb.FileDownloadRequest{
		Id: fileID,
	}

	// Вызываем DownloadFile.
	err = srv.DownloadFile(req, fakeStream)
	require.NoError(t, err)

	// Собираем полученные чанки и сверяем с исходным содержимым.
	var receivedContent bytes.Buffer
	for _, chunk := range fakeStream.chunks {
		receivedContent.Write(chunk.ChunkData)
	}
	assert.Equal(t, fileContent, receivedContent.String())
}

// -------------------------
// Тест для SyncRecords
// -------------------------
func TestSyncRecords_Success(t *testing.T) {
	// Фиктивный авторизатор, возвращающий фиксированный userID.
	userID := "user123"
	auth := &fakeAuthenticator{userID: userID}

	// Фиктивный репозиторий.
	var savedItems []entity.DataItem
	repo := &fakeRepository{
		getUserItemsFunc: func(u string) ([]entity.DataItem, error) {
			// Для теста сервер не имеет записей.
			return []entity.DataItem{}, nil
		},
		saveItemsFunc: func(items []entity.DataItem) error {
			savedItems = items
			return nil
		},
	}

	// Создаем экземпляр сервера.
	srv := &fileSyncServiceServer{
		uploadDir:          "dummy", // не используется в SyncRecords
		authenticator:      auth,
		dataItemRepository: repo,
	}

	// Формируем запрос с одним клиентским элементом.
	now := time.Now().UTC()
	pbItem := &pb.DataItem{
		Id:        "item1",
		Type:      int32(entity.DataTypeText),
		Content:   "client content",
		UpdatedAt: now.Format(time.RFC3339),
	}
	req := &pb.SyncRecordsRequest{
		Items: []*pb.DataItem{pbItem},
	}

	// Вызываем SyncRecords.
	resp, err := srv.SyncRecords(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Так как сервер не имел записей, mergedItems должны совпадать с клиентскими.
	assert.Len(t, resp.MergedRecords, 1)
	assert.Len(t, resp.UploadList, 1)
	assert.Len(t, resp.DownloadList, 0)
	// Проверяем, что сохраненные записи совпадают.
	require.Len(t, savedItems, 1)
	assert.Equal(t, "item1", savedItems[0].ID)
	assert.Equal(t, "client content", savedItems[0].Content)
	assert.Equal(t, userID, savedItems[0].UserID)
}

// -------------------------
// Тест для UploadFile
// -------------------------
func TestUploadFile_Success(t *testing.T) {
	// Создаем временную директорию для загрузок.
	tempUploadDir, err := os.MkdirTemp("", "uploadDir")
	require.NoError(t, err)
	defer os.RemoveAll(tempUploadDir)

	// Фиктивный авторизатор, возвращающий userID.
	userID := "testuser"
	auth := &fakeAuthenticator{userID: userID}

	// Создаем экземпляр сервера.
	srv := &fileSyncServiceServer{
		uploadDir:     tempUploadDir,
		authenticator: auth,
	}

	// Формируем набор чанков для эмуляции загрузки файла.
	// Пусть файл состоит из двух чанков.
	fileID := "file456"
	chunk1 := &pb.FileChunk{
		Id:        fileID,
		ChunkData: []byte("Hello "),
	}
	chunk2 := &pb.FileChunk{
		Id:        fileID,
		ChunkData: []byte("World!"),
	}
	chunks := []*pb.FileChunk{chunk1, chunk2}

	// Создаем фиктивный uploadStream с подготовленным списком чанков.
	fakeStream := &fakeUploadStream{
		chunks: chunks,
		ctx:    context.Background(),
	}

	// Вызываем UploadFile.
	err = srv.UploadFile(fakeStream)
	require.NoError(t, err)
	require.NotNil(t, fakeStream.resp)
	// Проверяем, что в ответе указано fileID и успех.
	assert.Equal(t, fileID, fakeStream.resp.Id)
	assert.True(t, fakeStream.resp.Success)
	// Проверяем, что файл перемещен в директорию uploadDir/userID с именем fileID.
	finalPath := filepath.Join(tempUploadDir, userID, fileID)
	// Убеждаемся, что файл существует.
	_, err = os.Stat(finalPath)
	require.NoError(t, err, "Файл должен быть перемещен в итоговую директорию")
	// Считываем содержимое файла и сверяем.
	data, err := os.ReadFile(finalPath)
	require.NoError(t, err)
	expectedContent := "Hello World!"
	// Удаляем возможные лишние пробелы или символы переноса строк.
	assert.Equal(t, expectedContent, strings.TrimSpace(string(data)))
}
