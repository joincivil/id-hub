package idhubmain

import (
	"fmt"

	"github.com/pkg/errors"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/utils"
	"github.com/robfig/cron/v3"
)

// RootCron controls the root commit cron process
type RootCron struct {
	cr *cron.Cron
}

// CheckCron emits cron messages
func (s *RootCron) CheckCron() {
	if s.cr != nil {
		entries := s.cr.Entries()
		for _, entry := range entries {
			log.Infof("Proc run times: prev: %v, next: %v\n", entry.Prev, entry.Next)
		}
	}
}

// StartCron kicks off the cron process
func (s *RootCron) StartCron(config *utils.IDHubConfig) error {
	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm: %v", err)
	}

	persister := initRootClaimPersister(db)

	ethHelper, err := initETHHelper(config)
	if err != nil {
		log.Fatalf("error initializing eth helper: %v", err)
	}

	treeStore := initTreePersister(db)

	rootService, err := initRootService(config, ethHelper, treeStore, persister)
	if err != nil {
		log.Fatalf("error initializing root service: %v", err)
	}

	s.cr = cron.New()
	s.RunCronProcess(rootService)
	// Start up the cron to run it periodically
	log.Infof("Cron config: %v", config.CronConfig)
	_, err = s.cr.AddFunc(config.CronConfig, func() {
		s.RunCronProcess(rootService)
	})
	if err != nil {
		log.Errorf("Error starting: err: %v", err)
		return errors.WithMessagef(err, "error starting cron")
	}

	s.cr.Start()
	s.CheckCron()

	// Blocks here while the cron process runs
	select {}
}

// RunCronProcess executes the scheduled code
func (s *RootCron) RunCronProcess(rootService *claims.RootService) {
	latest, err := rootService.GetLatest()
	if err != nil {
		log.Errorf("couldn't retrieve latest committed root: err: %v", err)
	}
	current, err := rootService.GetCurrent()
	if err != nil {
		log.Errorf("couldn't get current root: err: %v", err)
		return
	}
	if latest != nil && latest.Root == current {
		log.Infof("root already committed")
		fmt.Println("root already committed")
	} else {
		err := rootService.CommitRoot()
		if err != nil {
			log.Errorf("Error with committing root: err: %v", err)
		}
	}

	s.CheckCron()
	log.Infof("Cron job complete")
}
