package hedgehog

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

const (
	// UsernameExistsKey is used for the key to reserve a username
	UsernameExistsKey = "exists"
)

var (
	// ErrorUsernameExists is thrown when trying to reserve a username that is already in use
	ErrorUsernameExists = errors.New("username exists")
	// ErrorKeyExists is thrown when attempting to save a key that already exists
	ErrorKeyExists = errors.New("key exists")
	// ErrorNotFound is thrown when attempting to retrieve a key that does not exist
	ErrorNotFound = errors.New("not found")
)

// EncryptedData contains fields needed to store encrypted data
type EncryptedData struct {
	Iv         string `json:"iv"`
	CipherText string `json:"cipherText"`
}

// DataVaultItem is the gorm model that stores encrypted data
type DataVaultItem struct {
	Namespace string `gorm:"PRIMARY_KEY"`
	Key       string `gorm:"PRIMARY_KEY"`
	EncryptedData
}

// NewService constructs a new Service instance
func NewService(db *gorm.DB) *Service {
	return &Service{db}
}

// Service provides methods to store and retrieve encrypted data
type Service struct {
	db *gorm.DB
}

// GetAuthData retrieves stored auth data
func (s *Service) GetAuthData(lookupKey string) (EncryptedData, error) {
	fmt.Printf("lookupKey: %s\n", lookupKey)

	return s.GetItem("auth", lookupKey)
}

// SetAuthData inserts iv, cipherText, and lookupKey into the data item table.
func (s *Service) SetAuthData(lookupKey string, encryptedData EncryptedData) error {

	err := s.persist("auth", lookupKey, encryptedData, false)
	if err != nil {
		return err
	}

	return nil
}

// ReserveUsername sets a data item to identity that a username is in use
func (s *Service) ReserveUsername(username string) error {
	err := s.persist(username, UsernameExistsKey, EncryptedData{}, false)
	if err == ErrorKeyExists {
		return ErrorUsernameExists
	} else if err != nil {
		return err
	}

	return nil
}

// StoreItem stores an encrypted data item
func (s *Service) StoreItem(username string, key string, data EncryptedData) error {
	return s.persist(username, key, data, true)
}

// StoreItem stores an encrypted data item
func (s *Service) persist(username string, key string, data EncryptedData, allowUpdate bool) error {
	normalizedUsername := normalizeUsername(username)
	item := DataVaultItem{
		Namespace:     normalizedUsername,
		Key:           key,
		EncryptedData: data,
	}

	keyExists := !s.db.First(&DataVaultItem{}, "namespace = ? and key = ?", normalizedUsername, key).RecordNotFound()
	if keyExists && allowUpdate {
		return s.db.Model(&item).Update(data).Error
	} else if keyExists {
		return ErrorKeyExists
	}

	return s.db.Create(&item).Error
}

// GetItem retrieves an encrypted data item
func (s *Service) GetItem(username string, key string) (EncryptedData, error) {

	var item DataVaultItem
	err := s.db.Where("namespace = ? and key = ?", username, key).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return EncryptedData{}, ErrorNotFound
		}
		return EncryptedData{}, err
	}

	return item.EncryptedData, nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(username)
}
