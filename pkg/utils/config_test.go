package utils_test

import (
	"os"
	"testing"

	cconfig "github.com/joincivil/go-common/pkg/config"
	"github.com/joincivil/id-hub/pkg/utils"
)

func setEnvironmentVariables() {
	_ = os.Setenv(
		"IDHUB_GQLPORT",
		"8080",
	)
	_ = os.Setenv(
		"IDHUB_PERSISTER_TYPE_NAME",
		"postgresql",
	)
	_ = os.Setenv(
		"IDHUB_PERSISTER_POSTGRES_ADDRESS",
		"localhost",
	)
	_ = os.Setenv(
		"IDHUB_PERSISTER_POSTGRES_PORT",
		"5432",
	)
	_ = os.Setenv(
		"IDHUB_PERSISTER_POSTGRES_DBNAME",
		"docker",
	)
}

func TestIDHubConfig(t *testing.T) {
	setEnvironmentVariables()
	config := &utils.IDHubConfig{}
	err := config.PopulateFromEnv()
	if err != nil {
		t.Errorf("Failed to populate from environment: err: %v", err)
	}
	if config.GqlPort != 8080 {
		t.Error("Should have gotten 8080 for port")
	}
	if config.PersisterType != cconfig.PersisterTypePostgresql {
		t.Error("Should have gotten postgresql for persister type")
	}
	if config.PersisterPostgresAddress != "localhost" {
		t.Error("Should have gotten postgresql address")
	}
	if config.PersisterPostgresPort != 5432 {
		t.Error("Should have gotten postgresql port")
	}
	if config.PersisterPostgresDbname != "docker" {
		t.Error("Should have gotten postgresql dbname")
	}
}

func TestIDHubConfigUsage(t *testing.T) {
	config := &utils.IDHubConfig{}
	config.OutputUsage()
}
