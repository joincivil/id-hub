package idhubmain

import (
	log "github.com/golang/glog"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/id-hub/pkg/utils"
)

// initETHHelper makes an eth helper from the config
func initETHHelper(config *utils.IDHubConfig) (*eth.Helper, error) {
	if config.EthAPIURL != "" {
		accounts := map[string]string{}
		if config.EthereumDefaultPrivateKey != "" {
			log.Infof("Initialized default Ethereum account\n")
			accounts["default"] = config.EthereumDefaultPrivateKey
		}
		ethHelper, err := eth.NewETHClientHelper(config.EthAPIURL, accounts)
		if err != nil {
			return nil, err
		}
		log.Infof("Connected to Ethereum using %v\n", config.EthAPIURL)
		return ethHelper, nil
	}

	ethHelper, err := eth.NewSimulatedBackendHelper()
	if err != nil {
		return nil, err
	}
	log.Infof("Connected to Ethereum using Simulated Backend\n")
	return ethHelper, nil
}
