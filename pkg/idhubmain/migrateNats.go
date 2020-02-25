package idhubmain

import (
	"io/ioutil"

	log "github.com/golang/glog"
)

// RunMigration runs the nats server migration
func RunMigration() error {
	config := populateConfig()

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
		log.Errorf("error executing sql script: err: %v", err)
	}

	return nil
}
