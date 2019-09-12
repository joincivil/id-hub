package idhubmain

import (
	"fmt"

	"github.com/pkg/errors"

	log "github.com/golang/glog"

	"github.com/joincivil/id-hub/pkg/utils"

	"github.com/jinzhu/gorm"
)

const (
	defaultMaxIdleConns = 5
	defaultMaxOpenConns = 10
)

func initGorm(config *utils.IDHubConfig) (*gorm.DB, error) {
	return NewGormPostgres(GormPostgresConfig{
		Host:     config.PersisterPostgresAddress,
		Port:     config.PersisterPostgresPort,
		Dbname:   config.PersisterPostgresDbname,
		User:     config.PersisterPostgresUser,
		Password: config.PersisterPostgresPw,
	})
}

// GormPostgresConfig is the config struct for initializing a GORM object for
// Postgres.
type GormPostgresConfig struct {
	Host         string
	Port         int
	Dbname       string
	User         string
	Password     string
	MaxIdleConns *int
	MaxOpenConns *int
	LogMode      bool
}

// NewGormPostgres initializes a new GORM object for Postgres
func NewGormPostgres(creds GormPostgresConfig) (*gorm.DB, error) {
	if creds.Host == "" || creds.Port == 0 || creds.User == "" ||
		creds.Password == "" || creds.Dbname == "" {
		return nil, errors.New("all db creds required")
	}

	connStr := fmt.Sprintf(
		"host=%v port=%v user=%v dbname=%v password=%v sslmode=disable",
		creds.Host, creds.Port, creds.User, creds.Dbname, creds.Password)

	log.Infof("Connecting to database: %v\n", connStr)

	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		log.Errorf("Error opening database connection:: err: %v", err)
		return nil, errors.Wrap(err, "error opening gorm connection")
	}

	if creds.MaxIdleConns != nil {
		db.DB().SetMaxIdleConns(*creds.MaxIdleConns)
	} else {
		db.DB().SetMaxIdleConns(defaultMaxIdleConns)
	}

	if creds.MaxOpenConns != nil {
		db.DB().SetMaxOpenConns(*creds.MaxOpenConns)
	} else {
		db.DB().SetMaxOpenConns(defaultMaxOpenConns)
	}

	db.LogMode(creds.LogMode)
	return db, nil
}
