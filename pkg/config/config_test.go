package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/comfforts/comff-stores/pkg/logging"
)

func TestGetConfig(t *testing.T) {
	err := createConfigFile()
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)

	config, err := GetAppConfig(appLogger, "")
	require.NoError(t, err)
	require.Equal(t, 50051, config.ServicePort)
	require.Equal(t, "mustum-data", config.Services.CloudStorageCfg.BucketName)
	require.Equal(t, "creds/mustum-420420-z335d4c3c763.json", config.Services.CloudStorageCfg.CredsPath)
	require.Equal(t, "APIKEY34df56APIKEY", config.Services.GeoCodeCfg.GeocoderKey)
	require.Equal(t, false, config.Services.GeoCodeCfg.IsCached)
	require.Equal(t, "data", config.Jobs.StoreLoaderConfig.DataPath)
	require.Equal(t, "mustum-data", config.Jobs.StoreLoaderConfig.CloudStorageCfg.BucketName)
	require.Equal(t, "creds/mustum-420420-z335d4c3c763.json", config.Jobs.StoreLoaderConfig.CloudStorageCfg.CredsPath)

	err = removeConfigFile()
	require.NoError(t, err)
}

func createConfigFile() error {
	cfgJSON := map[string]interface{}{
		"service_port": 50051,
		"services": map[string]interface{}{
			"cloud_storage": map[string]interface{}{
				"creds_path":  "creds/mustum-420420-z335d4c3c763.json",
				"bucket_name": "mustum-data",
			},
			"geo_code": map[string]interface{}{
				"geocoder_key": "APIKEY34df56APIKEY",
			},
		},
		"jobs": map[string]interface{}{
			"store_loader": map[string]interface{}{
				"cloud_storage": map[string]interface{}{
					"creds_path":  "creds/mustum-420420-z335d4c3c763.json",
					"bucket_name": "mustum-data",
				},
				"data_path": "data",
			},
		},
	}

	file, err := os.Create(CONFIG_FILE_NAME)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfgJSON)
	if err != nil {
		return err
	}
	return nil
}

func removeConfigFile() error {
	return os.Remove(CONFIG_FILE_NAME)
}
