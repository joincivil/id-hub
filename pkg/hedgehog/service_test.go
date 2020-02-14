package hedgehog_test

import (
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/hedgehog"
	"github.com/joincivil/id-hub/pkg/testutils"
	"strings"
	"testing"
)

const user1 = "test-user1"
const user2 = "test-user2"
const key1 = "test-key1"

func setupDB() *gorm.DB {
	db, _ := testutils.GetTestDBConnection()
	db.AutoMigrate(&hedgehog.DataVaultItem{})
	_ = db.Where("namespace LIKE 'test%'").Delete(&hedgehog.DataVaultItem{}).Error
	_ = db.Where("key LIKE 'test%'").Delete(&hedgehog.DataVaultItem{}).Error
	return db
}

func TestService(t *testing.T) {
	db := setupDB()
	service := hedgehog.NewService(db)

	// SetAuthData
	t.Run("SetAuthData+GetAuthData", func(t *testing.T) {
		t.Parallel()

		const lookupKey = "test-lookup-key"
		data := hedgehog.EncryptedData{Iv: "foo", CipherText: "bar"}

		err := service.SetAuthData(lookupKey, data)
		if err != nil {
			t.Errorf("error setting auth data: %v", err)
		}

		returnedData, err := service.GetAuthData(lookupKey)
		if err != nil {
			t.Errorf("GetAuthData error: %v", err)
		}

		if returnedData.Iv != data.Iv && returnedData.CipherText != data.CipherText {
			t.Error("returned data not matching")
		}

		_, err = service.GetAuthData("nonexistent-lookup-key")
		if err != hedgehog.ErrorNotFound {
			t.Errorf("expecting ErrorNotFound: %v", err)
		}

	})
	// ReserveUsername
	t.Run("ReserveUsername", func(t *testing.T) {
		t.Parallel()
		// reserve a username
		err := service.ReserveUsername(user1)
		if err != nil {
			t.Errorf("error reserving username: %v", err)
		}

		// reserving the same should fail
		err = service.ReserveUsername(strings.ToUpper(user1))
		if err == nil {
			t.Error("expecting reserving a duplicate username to fail")
		}
		if err != hedgehog.ErrorUsernameExists {
			t.Errorf("expecting ErrorUsernameExists error but got error: %v", err)
		}

		// reserve another username
		err = service.ReserveUsername(user2)
		if err != nil {
			t.Errorf("error reserving another username: %v", err)
		}
	})

	// StoreItem / GetItem
	t.Run("StoreItem+GetItem", func(t *testing.T) {
		t.Parallel()
		requestData := hedgehog.EncryptedData{Iv: "foo", CipherText: "bar"}
		err := service.StoreItem(user1, key1, requestData)
		if err != nil {
			t.Errorf("error storing item: %v", err)
		}

		retrievedData, err := service.GetItem(user1, key1)
		if err != nil {
			t.Errorf("error getting item: %v", err)
		}

		if requestData.Iv != retrievedData.Iv {
			t.Errorf("retrieve data Iv doesn't match. request: %v | retrieved %v", requestData.Iv, retrievedData.Iv)
		}
		if requestData.CipherText != retrievedData.CipherText {
			t.Errorf("retrieve data CipherText doesn't match. request: %v | retrieved %v", requestData.CipherText, retrievedData.CipherText)
		}
	})

}
