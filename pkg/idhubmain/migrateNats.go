package idhubmain

import (
	"io/ioutil"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/utils"
)

func runMigration(config *utils.IDHubConfig) error {
	// init GORM
	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm")
	}

	migration, err := ioutil.ReadFile("./scripts/natspostgremigration.sql")
	if err != nil {
		log.Fatalf("error reading sql file: err: %v", err)
	}

	migrationS := string(migration)
	if err := db.Exec(migrationS).Error; err != nil {
		log.Fatalf("error executing sql script: err: %v", err)
	}

	return nil
}
