package testutils

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did/ethuri"

	// load postgres specific dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// DBCreds is a struct to contain the creds for the test db
type DBCreds struct {
	Port     int
	Dbname   string
	User     string
	Password string
	Host     string
}

// GetTestDBCreds returns the credentials for the local docker instance
// dependent on env vars.
func GetTestDBCreds() DBCreds {
	var creds DBCreds
	if os.Getenv("CI") == "true" {
		creds = DBCreds{
			Port:     5432,
			Dbname:   "circle_test",
			User:     "root",
			Password: "root",
			Host:     "localhost",
		}
	} else {
		creds = DBCreds{
			Port:     5432,
			Dbname:   "development",
			User:     "docker",
			Password: "docker",
			Host:     "localhost",
		}
	}
	return creds
}

var db *gorm.DB

// GetTestDBConnection returns a new gorm Database connection for the local docker instance
func GetTestDBConnection() (*gorm.DB, error) {
	if db == nil {
		creds := GetTestDBCreds()

		connStr := fmt.Sprintf(
			"host=%v port=%v user=%v dbname=%v password=%v sslmode=disable",
			creds.Host, creds.Port, creds.User, creds.Dbname, creds.Password)

		fmt.Printf("Connecting to database: %v\n", connStr)
		dbConn, err := gorm.Open("postgres", connStr)
		if err != nil {
			fmt.Printf("Error opening database connection:: err: %v", err)
			return nil, err
		}

		db = dbConn
		db.DB().SetMaxIdleConns(1)
		db.DB().SetMaxOpenConns(1)

		db.LogMode(true)
	}

	return db, nil
}

// SetupConnection returns a db instance
func SetupConnection() (*gorm.DB, error) {
	db, err := GetTestDBConnection()
	if err != nil {
		return nil, err
	}
	db.DropTable(&ethuri.PostgresDocument{}, &claimsstore.RootCommit{}, &claimsstore.Node{})
	err = db.AutoMigrate(&ethuri.PostgresDocument{}, &claimsstore.SignedClaimPostgres{}, &claimsstore.Node{}, &claimsstore.RootCommit{}, &claimsstore.JWTClaimPostgres{}).Error
	if err != nil {
		return nil, err
	}
	return db, nil
}
