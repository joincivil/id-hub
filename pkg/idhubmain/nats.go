package idhubmain

import (
	"fmt"

	"github.com/joincivil/id-hub/pkg/nats"
	"github.com/joincivil/id-hub/pkg/utils"
	stand "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats-streaming-server/stores"
	_ "github.com/nats-io/nats-streaming-server/stores/pqdeadlines"
	stan "github.com/nats-io/stan.go"
)

const (
	clientID = "id-hub-1"
)

func initializeNats(config *utils.IDHubConfig) (*nats.Service, error) {
	opts := stand.GetDefaultOptions()
	opts.StoreType = stores.TypeSQL
	sourceString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", config.PersisterPostgresAddress,
		config.PersisterPostgresPort, config.PersisterPostgresUser, config.PersisterPostgresPw,
		config.PersisterPostgresDbname)

	opts.SQLStoreOpts = stores.SQLStoreOptions{
		Driver: config.NatsPersisterDriver,
		Source: sourceString,
	}

	opts.ID = config.NatsID

	_, err := stand.RunServerWithOpts(opts, nil)
	if err != nil {
		return nil, err
	}

	sc, err := stan.Connect(config.NatsID, clientID)
	if err != nil {
		return nil, err
	}

	return nats.NewService(sc, config.NatsPrefix), nil
}
