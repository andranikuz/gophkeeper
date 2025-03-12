package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/andranikuz/gophkeeper/internal/filesync"
	"github.com/andranikuz/gophkeeper/pkg/entity"
	"github.com/andranikuz/gophkeeper/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== Фиктивные реализации =====

// fakeLocalStorage реализует интерфейс LocalStorage для тестирования.
type fakeLocalStorage struct {
	saveItemFunc    func(item *entity.DataItem) error
	saveItemsFunc   func(items []entity.DataItem) error
	getAllItemsFunc func() ([]entity.DataItem, error)
	getByIDFunc     func(id string) (*entity.DataItem, error)
	deleteItemFunc  func(id string) error
}

func (f *fakeLocalStorage) SaveItem(item *entity.DataItem) error {
	if f.saveItemFunc != nil {
		return f.saveItemFunc(item)
	}
	return nil
}
func (f *fakeLocalStorage) SaveItems(items []entity.DataItem) error {
	if f.saveItemsFunc != nil {
		return f.saveItemsFunc(items)
	}
	return nil
}
func (f *fakeLocalStorage) GetAllItems() ([]entity.DataItem, error) {
	if f.getAllItemsFunc != nil {
		return f.getAllItemsFunc()
	}
	return nil, nil
}
func (f *fakeLocalStorage) GetByID(id string) (*entity.DataItem, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(id)
	}
	return nil, nil
}
func (f *fakeLocalStorage) DeleteItem(id string) error {
	if f.deleteItemFunc != nil {
		return f.deleteItemFunc(id)
	}
	return nil
}

// fakeSession реализует интерфейс SessionService.
type fakeSession struct {
	token  string
	userID string
	// Функция, которую можно подменить для перехвата вызова Save.
	saveFunc func(token Token) error
}

func (fs *fakeSession) Save(token Token) error {
	if fs.saveFunc != nil {
		return fs.saveFunc(token)
	}
	// По умолчанию просто сохраняем данные.
	fs.token = token.Token
	fs.userID = token.UserID
	return nil
}
func (fs *fakeSession) GetSessionToken() string { return fs.token }
func (fs *fakeSession) GetUserID() string       { return fs.userID }

// fakeGrpcClient реализует интерфейс pb.FileSyncServiceClient.
type fakeGrpcClient struct {
	syncRecordsFunc  func(ctx context.Context, in *pb.SyncRecordsRequest, opts ...grpc.CallOption) (*pb.SyncRecordsResponse, error)
	uploadFileFunc   func(ctx context.Context, opts ...grpc.CallOption) (pb.FileSyncService_UploadFileClient, error)
	downloadFileFunc func(ctx context.Context, in *pb.FileDownloadRequest, opts ...grpc.CallOption) (pb.FileSyncService_DownloadFileClient, error)
}

func (f *fakeGrpcClient) SyncRecords(ctx context.Context, in *pb.SyncRecordsRequest, opts ...grpc.CallOption) (*pb.SyncRecordsResponse, error) {
	return f.syncRecordsFunc(ctx, in, opts...)
}
func (f *fakeGrpcClient) UploadFile(ctx context.Context, opts ...grpc.CallOption) (pb.FileSyncService_UploadFileClient, error) {
	return f.uploadFileFunc(ctx, opts...)
}
func (f *fakeGrpcClient) DownloadFile(ctx context.Context, in *pb.FileDownloadRequest, opts ...grpc.CallOption) (pb.FileSyncService_DownloadFileClient, error) {
	return f.downloadFileFunc(ctx, in, opts...)
}

// fakeUploadStream – фиктивный стрим для uploadFileGRPC.
type fakeUploadStream struct {
	sentChunks []*pb.FileChunk
}

func (s *fakeUploadStream) Send(chunk *pb.FileChunk) error {
	s.sentChunks = append(s.sentChunks, chunk)
	return nil
}
func (s *fakeUploadStream) CloseAndRecv() (*pb.FileUploadResponse, error) {
	return &pb.FileUploadResponse{Success: true, Message: "ok"}, nil
}
func (s *fakeUploadStream) Header() (metadata.MD, error) { return nil, nil }
func (s *fakeUploadStream) Trailer() metadata.MD         { return nil }
func (s *fakeUploadStream) CloseSend() error             { return nil }
func (s *fakeUploadStream) Context() context.Context     { return context.Background() }
func (s *fakeUploadStream) RecvMsg(m interface{}) error  { return nil }
func (s *fakeUploadStream) SendMsg(m interface{}) error  { return nil }

// fakeDownloadStream – фиктивный стрим для downloadFileGRPC.
type fakeDownloadStream struct {
	chunks []*pb.FileChunk
	index  int
}

func (s *fakeDownloadStream) Recv() (*pb.FileChunk, error) {
	if s.index >= len(s.chunks) {
		return nil, io.EOF
	}
	chunk := s.chunks[s.index]
	s.index++
	return chunk, nil
}
func (s *fakeDownloadStream) Header() (metadata.MD, error) { return nil, nil }
func (s *fakeDownloadStream) Trailer() metadata.MD         { return nil }
func (s *fakeDownloadStream) CloseSend() error             { return nil }
func (s *fakeDownloadStream) Context() context.Context     { return context.Background() }
func (s *fakeDownloadStream) RecvMsg(m interface{}) error  { return nil }
func (s *fakeDownloadStream) SendMsg(m interface{}) error  { return nil }

// ===== Тесты для CardDTO.Validate =====

func TestCardDTOValidate_Valid(t *testing.T) {
	// Используем формат MM/YYYY
	future := time.Now().AddDate(1, 0, 0)
	exp := future.Format("01/2006")
	dto := CardDTO{
		CardNumber:     "1234567890123", // 13 цифр
		ExpirationDate: exp,
		CVV:            "123",
		CardHolderName: "John Doe",
	}
	err := dto.Validate()
	assert.NoError(t, err)
}

func TestCardDTOValidate_InvalidCardNumber(t *testing.T) {
	future := time.Now().AddDate(1, 0, 0)
	exp := future.Format("01/2006")
	dto := CardDTO{
		CardNumber:     "123", // слишком короткий
		ExpirationDate: exp,
		CVV:            "123",
		CardHolderName: "John Doe",
	}
	err := dto.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid card number")
}

func TestCardDTOValidate_InvalidExpirationFormat(t *testing.T) {
	dto := CardDTO{
		CardNumber:     "1234567890123",
		ExpirationDate: "13-2025", // неверный формат
		CVV:            "123",
		CardHolderName: "John Doe",
	}
	err := dto.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expiration date format")
}

func TestCardDTOValidate_ExpiredCard(t *testing.T) {
	// Устанавливаем дату в прошлом
	past := time.Now().AddDate(-1, 0, 0)
	exp := past.Format("01/2006")
	dto := CardDTO{
		CardNumber:     "1234567890123",
		ExpirationDate: exp,
		CVV:            "123",
		CardHolderName: "John Doe",
	}
	err := dto.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card is expired")
}

func TestCardDTOValidate_InvalidCVV(t *testing.T) {
	future := time.Now().AddDate(1, 0, 0)
	exp := future.Format("01/2006")
	dto := CardDTO{
		CardNumber:     "1234567890123",
		ExpirationDate: exp,
		CVV:            "12a", // содержит не только цифры
		CardHolderName: "John Doe",
	}
	err := dto.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CVV")
}

func TestCardDTOValidate_EmptyCardHolder(t *testing.T) {
	future := time.Now().AddDate(1, 0, 0)
	exp := future.Format("01/2006")
	dto := CardDTO{
		CardNumber:     "1234567890123",
		ExpirationDate: exp,
		CVV:            "123",
		CardHolderName: "",
	}
	err := dto.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card holder name cannot be empty")
}

// ===== Тесты для SaveCard =====

func TestSaveCard_Success(t *testing.T) {
	var savedItem *entity.DataItem
	fakeStore := &fakeLocalStorage{
		saveItemFunc: func(item *entity.DataItem) error {
			savedItem = item
			return nil
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}
	// Используем формат MM/YYYY
	future := time.Now().AddDate(1, 0, 0)
	exp := future.Format("01/2006")
	card := CardDTO{
		CardNumber:     "1234567890123456",
		ExpirationDate: exp,
		CVV:            "123",
		CardHolderName: "John Doe",
	}
	err := client.SaveCard(context.Background(), card)
	require.NoError(t, err)
	require.NotNil(t, savedItem)
	assert.Equal(t, entity.DataTypeCard, savedItem.Type)

	// Проверяем, что сохранённое содержимое корректно сериализовано
	var cardPayload CardDTO
	err = json.Unmarshal([]byte(savedItem.Content), &cardPayload)
	require.NoError(t, err)
	assert.Equal(t, card.CardNumber, cardPayload.CardNumber)
	assert.Equal(t, card.ExpirationDate, cardPayload.ExpirationDate)
	assert.Equal(t, card.CVV, cardPayload.CVV)
	assert.Equal(t, card.CardHolderName, cardPayload.CardHolderName)
}

func TestSaveCard_Invalid(t *testing.T) {
	fakeStore := &fakeLocalStorage{}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}
	card := CardDTO{
		CardNumber:     "123", // неверный номер
		ExpirationDate: "01/2006",
		CVV:            "12",
		CardHolderName: "",
	}
	err := client.SaveCard(context.Background(), card)
	assert.Error(t, err)
}

// ===== Тест для SaveCredential =====

func TestSaveCredential_Success(t *testing.T) {
	var savedItem *entity.DataItem
	fakeStore := &fakeLocalStorage{
		saveItemFunc: func(item *entity.DataItem) error {
			savedItem = item
			return nil
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}
	cred := CredentialDTO{
		Login:    "user@example.com",
		Password: "securepassword",
	}
	err := client.SaveCredential(context.Background(), cred)
	require.NoError(t, err)
	require.NotNil(t, savedItem)
	assert.Equal(t, entity.DataTypeCredential, savedItem.Type)

	var credPayload CredentialDTO
	err = json.Unmarshal([]byte(savedItem.Content), &credPayload)
	require.NoError(t, err)
	assert.Equal(t, cred.Login, credPayload.Login)
	assert.Equal(t, cred.Password, credPayload.Password)
}

// ===== Тест для SaveText =====

func TestSaveText_Success(t *testing.T) {
	var savedItem *entity.DataItem
	fakeStore := &fakeLocalStorage{
		saveItemFunc: func(item *entity.DataItem) error {
			savedItem = item
			return nil
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}
	textDTO := TextDTO{
		Text: "sample text",
	}
	err := client.SaveText(context.Background(), textDTO)
	require.NoError(t, err)
	require.NotNil(t, savedItem)
	assert.Equal(t, entity.DataTypeText, savedItem.Type)
	assert.Equal(t, textDTO.Text, savedItem.Content)
}

// ===== Тест для SaveFile =====

func TestSaveFile_Success(t *testing.T) {
	// Создаем временный исходный файл
	srcFile, err := os.CreateTemp("", "srcfile")
	require.NoError(t, err)
	content := []byte("file content for testing")
	_, err = srcFile.Write(content)
	require.NoError(t, err)
	srcFile.Close()
	defer os.Remove(srcFile.Name())

	// Сохраняем текущую рабочую директорию и создаем временную рабочую директорию
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	tempWd, err := os.MkdirTemp("", "tempWd")
	require.NoError(t, err)
	defer func() {
		// Возвращаем рабочую директорию и удаляем временную
		os.Chdir(oldWd)
		os.RemoveAll(tempWd)
	}()

	// Меняем текущую директорию на временную.
	require.NoError(t, os.Chdir(tempWd))

	var savedItem *entity.DataItem
	fakeStore := &fakeLocalStorage{
		saveItemFunc: func(item *entity.DataItem) error {
			savedItem = item
			return nil
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}
	dto := FileDTO{FilePath: srcFile.Name()}
	err = client.SaveFile(context.Background(), dto)
	require.NoError(t, err)
	require.NotNil(t, savedItem)

	// Определяем путь, куда файл должен был быть сохранён. Он вычисляется как "./data/client_files/<id><ext>"
	destPath := utils.GetLocalFilePath(savedItem)
	copiedContent, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, content, copiedContent)
	// Проверяем, что в DataItem в поле Content записано базовое имя исходного файла.
	assert.Equal(t, filepath.Base(srcFile.Name()), savedItem.Content)
}

// ===== Тест для SyncGRPC =====

func TestSyncGRPC_Success(t *testing.T) {
	// Фиктивное локальное хранилище с одной записью.
	localItems := []entity.DataItem{
		{ID: "1", Type: entity.DataTypeText, Content: "text1", UpdatedAt: time.Now()},
	}
	var savedItems []entity.DataItem
	fakeStore := &fakeLocalStorage{
		getAllItemsFunc: func() ([]entity.DataItem, error) {
			return localItems, nil
		},
		saveItemsFunc: func(items []entity.DataItem) error {
			savedItems = items
			return nil
		},
	}
	fakeSess := &fakeSession{userID: "user123", token: "testtoken"}

	// Подготавливаем фиктивный ответ SyncRecords:
	mergedTime := time.Now().Format(time.RFC3339)
	fakeResp := &pb.SyncRecordsResponse{
		MergedRecords: []*pb.DataItem{
			{Id: "merged1", Type: int32(entity.DataTypeText), Content: "merged text", UpdatedAt: mergedTime},
		},
		UploadList:   nil,
		DownloadList: nil,
	}
	fakeGrpc := &fakeGrpcClient{
		syncRecordsFunc: func(ctx context.Context, req *pb.SyncRecordsRequest, opts ...grpc.CallOption) (*pb.SyncRecordsResponse, error) {
			return fakeResp, nil
		},
		uploadFileFunc: func(ctx context.Context, opts ...grpc.CallOption) (pb.FileSyncService_UploadFileClient, error) {
			return &fakeUploadStream{}, nil
		},
		downloadFileFunc: func(ctx context.Context, req *pb.FileDownloadRequest, opts ...grpc.CallOption) (pb.FileSyncService_DownloadFileClient, error) {
			return &fakeDownloadStream{}, nil
		},
	}

	client := &Client{
		LocalDB:    fakeStore,
		Session:    fakeSess,
		grpcClient: fakeGrpc,
	}

	err := client.SyncGRPC(context.Background())
	require.NoError(t, err)
	// Проверяем, что метод SaveItems был вызван с данными, преобразованными из mergedRecords.
	require.Len(t, savedItems, 1)
	assert.Equal(t, "merged1", savedItems[0].ID)
	assert.Equal(t, entity.DataTypeText, savedItems[0].Type)
	assert.Equal(t, "merged text", savedItems[0].Content)
}

// ===== Тесты для Register =====

func TestRegister_Success(t *testing.T) {
	// Сервер возвращает статус 201 Created.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	fakeStore := &fakeLocalStorage{}
	fakeSess := &fakeSession{}
	client := &Client{
		ServerURL: ts.URL,
		LocalDB:   fakeStore,
		Session:   fakeSess,
	}
	dto := RegisterDTO{
		Username: "newuser",
		Password: "newpassword",
	}
	err := client.Register(context.Background(), dto)
	require.NoError(t, err)
}

func TestRegister_Failure(t *testing.T) {
	// Сервер возвращает ошибку регистрации.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "username already exists", http.StatusBadRequest)
	}))
	defer ts.Close()

	fakeStore := &fakeLocalStorage{}
	fakeSess := &fakeSession{}
	client := &Client{
		ServerURL: ts.URL,
		LocalDB:   fakeStore,
		Session:   fakeSess,
	}
	dto := RegisterDTO{
		Username: "existinguser",
		Password: "password",
	}
	err := client.Register(context.Background(), dto)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed")
}

// ===== Тесты для Login =====

func TestLogin_Success(t *testing.T) {
	// Создаем тестовый HTTP-сервер, который симулирует эндпоинт /login.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]string{
			"token":   "testtoken",
			"user_id": "user123",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Фиктивное хранилище.
	fakeStore := &fakeLocalStorage{}
	// Создаем фиктивную сессию.
	var savedToken Token
	fakeSess := &fakeSession{userID: ""}
	// Задаем функцию Save отдельно, чтобы можно было ссылаться на fakeSess.
	fakeSess.saveFunc = func(token Token) error {
		savedToken = token
		fakeSess.userID = token.UserID
		return nil
	}

	client := &Client{
		ServerURL: ts.URL,
		LocalDB:   fakeStore,
		Session:   fakeSess,
	}
	dto := LoginDTO{
		Username: "testuser",
		Password: "testpassword",
	}
	err := client.Login(context.Background(), dto)
	require.NoError(t, err)
	assert.Equal(t, "testtoken", savedToken.Token)
	assert.Equal(t, "user123", savedToken.UserID)
}

func TestLogin_Failure(t *testing.T) {
	// Сервер возвращает ошибку.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	}))
	defer ts.Close()

	fakeStore := &fakeLocalStorage{}
	fakeSess := &fakeSession{}
	client := &Client{
		ServerURL: ts.URL,
		LocalDB:   fakeStore,
		Session:   fakeSess,
	}
	dto := LoginDTO{
		Username: "testuser",
		Password: "wrongpassword",
	}
	err := client.Login(context.Background(), dto)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
}

// ===== Тесты для DeleteItem =====

// DeleteItem удаляет запись из LocalDB, а если тип записи — файл, дополнительно удаляет файл с диска.
// Предполагается, что путь к файлу определяется функцией utils.GetLocalFilePath.
func TestDeleteItem_FileDeletion_Success(t *testing.T) {
	// Создаем временную директорию для теста.
	tempDir, err := os.MkdirTemp("", "testdelete")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Формируем ID и базовое имя файла.
	itemID := "test-id"
	baseFileName := "file.txt"

	// Создаем файл по пути, который вернет utils.GetLocalFilePath.
	item := &entity.DataItem{
		ID:      itemID,
		Type:    entity.DataTypeBinary,
		Content: baseFileName,
	}
	expectedPath := utils.GetLocalFilePath(item)
	// Для теста создаем директорию, если нужно.
	require.NoError(t, os.MkdirAll(filepath.Dir(expectedPath), 0755))
	// Записываем временное содержимое, чтобы файл существовал.
	err = os.WriteFile(expectedPath, []byte("dummy content"), 0644)
	require.NoError(t, err)

	// Фиктивное хранилище возвращает элемент по ID.
	fakeStore := &fakeLocalStorage{
		getByIDFunc: func(id string) (*entity.DataItem, error) {
			if id == itemID {
				return item, nil
			}
			return nil, errors.New("not found")
		},
		deleteItemFunc: func(id string) error {
			return nil
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}

	err = client.DeleteItem(context.Background(), itemID)
	require.NoError(t, err)

	// Проверяем, что файл удален.
	_, err = os.Stat(expectedPath)
	assert.True(t, os.IsNotExist(err), "Файл должен быть удален с диска")
}

func TestDeleteItem_GetByIDError(t *testing.T) {
	fakeStore := &fakeLocalStorage{
		getByIDFunc: func(id string) (*entity.DataItem, error) {
			return nil, errors.New("item not found")
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}

	err := client.DeleteItem(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get item by ID")
}

func TestDeleteItem_NonBinary(t *testing.T) {
	deleteCalled := false
	fakeStore := &fakeLocalStorage{
		getByIDFunc: func(id string) (*entity.DataItem, error) {
			return &entity.DataItem{
				ID:      id,
				Type:    entity.DataTypeText, // не бинарный тип
				Content: "some text",
			}, nil
		},
		deleteItemFunc: func(id string) error {
			deleteCalled = true
			return nil
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}
	err := client.DeleteItem(context.Background(), "test-id")
	require.NoError(t, err)
	assert.True(t, deleteCalled, "Метод DeleteItem в LocalDB должен быть вызван")
}

// ===== Тест для GetItems =====

func TestGetItems(t *testing.T) {
	expectedItems := []entity.DataItem{
		{ID: "1", Type: entity.DataTypeText, Content: "content1"},
		{ID: "2", Type: entity.DataTypeCard, Content: "content2"},
	}
	fakeStore := &fakeLocalStorage{
		getAllItemsFunc: func() ([]entity.DataItem, error) {
			return expectedItems, nil
		},
	}
	fakeSess := &fakeSession{userID: "user123"}
	client := &Client{
		LocalDB: fakeStore,
		Session: fakeSess,
	}
	items, err := client.GetItems(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expectedItems, items)
}
