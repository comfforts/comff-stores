package config

import (
	"bufio"
	"encoding/json"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/comfforts/comff-stores/pkg/logging"
)

const (
	CONFIG_FILE_NAME string = "config.json"
)

type Configuration struct {
	ServicePort int            `json:"service_port"`
	Services    ServicesConfig `json:"services"`
	Jobs        JobsConfig     `json:"jobs"`
}

type CloudStorageClientConfig struct {
	BucketName string `json:"bucket_name"`
	CredsPath  string `json:"creds_path"`
}

type GeoCodeServiceConfig struct {
	GeocoderKey string `json:"geocoder_key"`
	IsCached    bool   `json:"cached"`
}

type ServicesConfig struct {
	CloudStorageCfg CloudStorageClientConfig `json:"cloud_storage"`
	GeoCodeCfg      GeoCodeServiceConfig     `json:"geo_code"`
}

type StoreLoaderConfig struct {
	CloudStorageCfg CloudStorageClientConfig `json:"cloud_storage"`
	DataPath        string                   `json:"data_path"`
}

type JobsConfig struct {
	StoreLoaderConfig StoreLoaderConfig `json:"store_loader"`
}

func GetAppConfig(logger *logging.AppLogger, fileName string) (*Configuration, error) {
	if fileName == "" {
		fileName = CONFIG_FILE_NAME
	}

	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Error("file doesn't exist", zap.Error(err), zap.String("filePath", fileName))
		} else {
			logger.Error("error accessing file", zap.Error(err), zap.String("filePath", fileName))
		}
		return nil, err
	}

	return getFromConfigJson(logger, fileName)
}

func getFromConfigJson(logger *logging.AppLogger, filePath string) (*Configuration, error) {
	f, err := os.Open(filePath)
	if err != nil {
		logger.Error("error opening file", zap.Error(err), zap.String("filePath", filePath))
		return nil, err
	}
	defer func() {
		if err = f.Close(); err != nil {
			logger.Error("error closing file", zap.Error(err), zap.String("filePath", filePath))
		}
	}()

	r := bufio.NewReader(f)
	dec := json.NewDecoder(r)

	var config = new(Configuration)

	for {
		var cfg Configuration
		if err := dec.Decode(&cfg); err == io.EOF {
			break
		} else if err != nil {
			logger.Error("error decoding config json", zap.Error(err), zap.String("filePath", filePath))
			return nil, err
		}

		config.Services = cfg.Services
		if config.Services.CloudStorageCfg.BucketName == "" || config.Services.CloudStorageCfg.CredsPath == "" {
			logger.Error("missing cloud storage config", zap.Error(err), zap.String("filePath", filePath))
			return nil, err
		}

		if config.Services.GeoCodeCfg.GeocoderKey == "" {
			logger.Error("missing geocoder config", zap.Error(err), zap.String("filePath", filePath))
			return nil, err
		}

		config.Jobs = cfg.Jobs
		if config.Jobs.StoreLoaderConfig.CloudStorageCfg.BucketName == "" || config.Jobs.StoreLoaderConfig.CloudStorageCfg.CredsPath == "" {
			logger.Error("missing store loader cloud storage config", zap.Error(err), zap.String("filePath", filePath))
			return nil, err
		}

		config.ServicePort = cfg.ServicePort
		if config.ServicePort == 0 {
			config.ServicePort = 8080
		}
	}
	return config, nil
}
