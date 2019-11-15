package idhubmain

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	log "github.com/golang/glog"

	"github.com/joincivil/id-hub/pkg/utils"

	"github.com/jinzhu/gorm"
)

const (
	defaultMaxIdleConns    = 5
	defaultMaxOpenConns    = 5
	defaultConnMaxLifetime = time.Second * 180 // 3 min
)

func initGorm(config *utils.IDHubConfig) (*gorm.DB, error) {
	return NewGormPostgres(GormPostgresConfig{
		Host:            config.PersisterPostgresAddress,
		Port:            config.PersisterPostgresPort,
		Dbname:          config.PersisterPostgresDbname,
		User:            config.PersisterPostgresUser,
		Password:        config.PersisterPostgresPw,
		MaxIdleConns:    config.PersisterPostgresMaxIdle,
		MaxOpenConns:    config.PersisterPostgresMaxConns,
		ConnMaxLifetime: config.PersisterPostgresConnLife,
	})
}

// GormPostgresConfig is the config struct for initializing a GORM object for
// Postgres.
type GormPostgresConfig struct {
	Host            string
	Port            int
	Dbname          string
	User            string
	Password        string
	MaxIdleConns    *int
	MaxOpenConns    *int
	ConnMaxLifetime *int
	LogMode         bool
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

	log.Infof("Connecting to database: %v, %v\n", creds.Host, creds.Port)

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

	if creds.ConnMaxLifetime != nil {
		db.DB().SetConnMaxLifetime(time.Second * time.Duration(*creds.ConnMaxLifetime))
	} else {
		db.DB().SetConnMaxLifetime(defaultConnMaxLifetime)
	}

	db.LogMode(creds.LogMode)
	return db, nil
}
