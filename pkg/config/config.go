package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/comfforts/errors"
	"github.com/comfforts/logger"

	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

const (
	CONFIG_FILE_NAME string = "config.json"
)

var (
	ErrMissingCloudConfig    = errors.NewAppError("missing cloud storage client config")
	ErrMissingGeoCoderConfig = errors.NewAppError("missing geocoder config")
)

type Configuration struct {
	ServicePort int            `json:"service_port"`
	Services    ServicesConfig `json:"services"`
	Jobs        JobsConfig     `json:"jobs"`
}

type CloudStorageClientConfig struct {
	CredsPath string `json:"creds_path"`
}

type GeoCodeServiceConfig struct {
	GeocoderKey string `json:"geocoder_key"`
	Host        string `json:"host"`
	Path        string `json:"path"`
	Cached      bool   `json:"cached"`
	DataDir     string `json:"data_dir"`
	BucketName  string `json:"bucket_name"`
}

type ServicesConfig struct {
	CloudStorageClientCfg CloudStorageClientConfig `json:"cloud_storage"`
	GeoCodeCfg            GeoCodeServiceConfig     `json:"geo_code"`
}

type StoreLoaderConfig struct {
	DataDir    string `json:"data_dir"`
	BucketName string `json:"bucket_name"`
}

type JobsConfig struct {
	StoreLoaderConfig StoreLoaderConfig `json:"store_loader"`
}

func GetAppConfig(filePath string, logger logger.AppLogger) (*Configuration, error) {
	if filePath == "" {
		filePath = CONFIG_FILE_NAME
	}

	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Error("file doesn't exist", zap.Error(err), zap.String("filePath", filePath))
			return nil, errors.WrapError(err, fileUtils.ERROR_NO_FILE, filePath)
		}
		logger.Error("error accessing file", zap.Error(err), zap.String("filePath", filePath))
		return nil, errors.WrapError(err, fileUtils.ERROR_FILE_INACCESSIBLE, filePath)
	}

	return getFromConfigJson(logger, filePath)
}

func getFromConfigJson(logger logger.AppLogger, filePath string) (*Configuration, error) {
	f, err := os.Open(filePath)
	if err != nil {
		logger.Error("error opening file", zap.Error(err), zap.String("filePath", filePath))
		return nil, errors.WrapError(err, fileUtils.ERROR_OPENING_FILE, filePath)
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
			return nil, errors.WrapError(err, "error decoding config json %s", filePath)
		}

		config.Services = cfg.Services
		if config.Services.CloudStorageClientCfg.CredsPath == "" {
			logger.Error("missing cloud storage client config", zap.String("filePath", filePath))
			return nil, ErrMissingCloudConfig
		}

		fmt.Printf("    config: %v\n", config)

		if config.Services.GeoCodeCfg.GeocoderKey == "" {
			logger.Error("missing geocoder config", zap.String("filePath", filePath))
			return nil, ErrMissingGeoCoderConfig
		}

		config.Jobs = cfg.Jobs
		if config.Jobs.StoreLoaderConfig.BucketName == "" {
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
