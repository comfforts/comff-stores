package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/comfforts/logger"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	ors := Overrides{}
	err := createConfigFile(ors)
	require.NoError(t, err)

	appLogger := logger.NewTestAppLogger("")
	config, err := GetAppConfig("", appLogger)
	require.NoError(t, err)
	require.Equal(t, 50051, config.ServicePort)

	require.Equal(t, "creds/mustum-420420-z335d4c3c763.json", config.Services.CloudStorageClientCfg.CredsPath)

	require.Equal(t, "APIKEY34df56APIKEY", config.Services.GeoCodeCfg.GeocoderKey)
	require.Equal(t, "mustum-geo", config.Services.GeoCodeCfg.BucketName)
	require.Equal(t, false, config.Services.GeoCodeCfg.Cached)

	require.Equal(t, "data", config.Jobs.StoreLoaderConfig.DataDir)
	require.Equal(t, "mustum-store", config.Jobs.StoreLoaderConfig.BucketName)

	err = removeConfigFile()
	require.NoError(t, err)
}

type GeoOverrides struct {
	Cached bool
	Bucket string
	Path   string
}

type Overrides struct {
	Geo GeoOverrides
}

func createConfigFile(ors Overrides) error {

	cfgJSON := map[string]interface{}{
		"service_port": 50051,
		"services": map[string]interface{}{
			"cloud_storage": map[string]interface{}{
				"creds_path": "creds/mustum-420420-z335d4c3c763.json",
			},
			"geo_code": map[string]interface{}{
				"geocoder_key": "APIKEY34df56APIKEY",
				"bucket_name":  "mustum-geo",
				"data_dir":     "geocode",
			},
		},
		"jobs": map[string]interface{}{
			"store_loader": map[string]interface{}{
				"bucket_name": "mustum-store",
				"data_dir":    "data",
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
