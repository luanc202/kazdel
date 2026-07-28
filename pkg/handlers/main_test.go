package handlers

import (
	"os"
	"testing"

	"kazdel/pkg/infra/config"
)

func TestMain(m *testing.M) {
	config.SetEnvConfigForTest(&config.EnvConfig{})
	os.Exit(m.Run())
}
