package config

import (
	"log"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/comfforts/comff-stores/pkg/logging"
)

func TestGetConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)
	config, _ := GetAppConfig(appLogger)
	log.Printf("Config: %v\n", config)
}
