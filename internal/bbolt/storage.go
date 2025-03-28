package bbolt

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/andranikuz/gophkeeper/pkg/entity"
)

const bucketName = "data"

// BboltStorage оборачивает BoltDB для хранения данных.
type BboltStorage struct {
	db *bolt.DB
}

// OpenLocalStorage открывает (или создаёт) базу BoltDB по указанному пути.
func OpenLocalStorage(path string) (BboltStorage, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return BboltStorage{}, err
	}
	// Создаём бакет, если он отсутствует.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		db.Close()
		return BboltStorage{}, err
	}
	return BboltStorage{db: db}, nil
}

// Close закрывает базу.
func (ls *BboltStorage) Close() error {
	return ls.db.Close()
}

// SaveItems сохраняет объект storage.DataItem в бакете.
func (ls BboltStorage) SaveItems(items []entity.DataItem) error {
	return ls.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		for _, item := range items {
			data, err := json.Marshal(item)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(item.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveItem сохраняет объект storage.DataItem в бакете.
func (ls BboltStorage) SaveItem(item *entity.DataItem) error {
	return ls.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(item.ID), data)
	})
}

// GetAllItems возвращает все объекты storage.DataItem из бакета.
func (ls BboltStorage) GetAllItems() ([]entity.DataItem, error) {
	var items []entity.DataItem
	err := ls.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return bucket.ForEach(func(k, v []byte) error {
			var item entity.DataItem
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteItem удаляет элемент из локального хранилища по его ID.
func (ls BboltStorage) DeleteItem(id string) error {
	return ls.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return bucket.Delete([]byte(id))
	})
}

// GetByID извлекает один объект entity.DataItem из локального хранилища по его ID.
func (ls BboltStorage) GetByID(id string) (*entity.DataItem, error) {
	var item *entity.DataItem
	err := ls.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		data := bucket.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("item with id %s not found", id)
		}
		if err := json.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("failed to unmarshal item: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}
