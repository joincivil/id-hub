package idhubmain

import (
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/did"
)

func initDidPersister(db *gorm.DB) did.Persister {
	persister := did.NewPostgresPersister(db)
	db.AutoMigrate(
		did.PostgresDocument{},
	)
	return persister
}
