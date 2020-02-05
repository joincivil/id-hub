package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"

	ccfg "github.com/joincivil/go-common/pkg/config"
)

const (
	envVarPrefixIDHub = "idhub"
)

// IDHubConfig is the master config for the ID Hub API derived from environment
// variables.
type IDHubConfig struct {
	GqlPort                   int    `required:"true" desc:"Sets the ID Hub GraphQL port"`
	RootCommitsAddress        string `split_words:"true" desc:"address where root commits are stored"`
	EthereumDefaultPrivateKey string `split_words:"true" desc:"Private key to use when sending Ethereum transactions"`
	EthAPIURL                 string `envconfig:"eth_api_url" desc:"Ethereum API address"`

	CronConfig string `envconfig:"cron_config" desc:"Cron config string * * * * *"`

	PersisterType             ccfg.PersisterType `ignored:"true"`
	PersisterTypeName         string             `split_words:"true" required:"true" desc:"Sets the persister type to use"`
	PersisterPostgresAddress  string             `split_words:"true" desc:"If persister type is Postgresql, sets the address"`
	PersisterPostgresPort     int                `split_words:"true" desc:"If persister type is Postgresql, sets the port"`
	PersisterPostgresDbname   string             `split_words:"true" desc:"If persister type is Postgresql, sets the database name"`
	PersisterPostgresUser     string             `split_words:"true" desc:"If persister type is Postgresql, sets the database user"`
	PersisterPostgresPw       string             `split_words:"true" desc:"If persister type is Postgresql, sets the database password"`
	PersisterPostgresMaxConns *int               `split_words:"true" desc:"If persister type is Postgresql, sets the max conns in pool"`
	PersisterPostgresMaxIdle  *int               `split_words:"true" desc:"If persister type is Postgresql, sets the max idle conns in pool"`
	PersisterPostgresConnLife *int               `split_words:"true" desc:"If persister type is Postgresql, sets the max conn lifetime in secs"`

	RedisHosts []string `split_words:"true" desc:"List of Redis host:port for caching and locking"`

	DidUniversalResolverHost *string `split_words:"true" desc:"Sets the host for the universal DID resolver"`
	DidUniversalResolverPort *int    `split_words:"true" desc:"Sets the port for the universal DID resolver"`
}

// OutputUsage prints the usage string to os.Stdout
func (c *IDHubConfig) OutputUsage() {
	ccfg.OutputUsage(c, envVarPrefixIDHub, envVarPrefixIDHub)
}

// PopulateFromEnv processes the environment vars, populates config
// with the respective values, and validates the values.
func (c *IDHubConfig) PopulateFromEnv() error {
	envEnvVar := fmt.Sprintf("%v_ENV", strings.ToUpper(envVarPrefixIDHub))
	err := ccfg.PopulateFromDotEnv(envEnvVar)
	if err != nil {
		return err
	}

	err = envconfig.Process(envVarPrefixIDHub, c)
	if err != nil {
		return err
	}

	err = c.populatePersisterType()
	if err != nil {
		return err
	}

	return c.validatePersister()
}

func (c *IDHubConfig) populatePersisterType() error {
	var err error
	c.PersisterType, err = ccfg.PersisterTypeFromName(c.PersisterTypeName)
	return err
}

func (c *IDHubConfig) validatePersister() error {
	var err error
	if c.PersisterType == ccfg.PersisterTypePostgresql {
		err = validatePostgresqlPersisterParams(
			c.PersisterPostgresAddress,
			c.PersisterPostgresPort,
			c.PersisterPostgresDbname,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func validatePostgresqlPersisterParams(address string, port int, dbname string) error {
	if address == "" {
		return errors.New("Postgresql address required")
	}
	if port == 0 {
		return errors.New("Postgresql port required")
	}
	if dbname == "" {
		return errors.New("Postgresql db name required")
	}
	return nil
}
