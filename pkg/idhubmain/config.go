package idhubmain

import (
	"flag"
	"os"

	log "github.com/golang/glog"

	"github.com/joincivil/id-hub/pkg/utils"
)

func populateConfig() *utils.IDHubConfig {
	config := &utils.IDHubConfig{}
	flag.Usage = func() {
		config.OutputUsage()
		os.Exit(0)
	}
	flag.Parse()
	err := config.PopulateFromEnv()
	if err != nil {
		config.OutputUsage()
		log.Errorf("Invalid idhub config: err: %v\n", err)
		os.Exit(2)
	}
	return config
}
