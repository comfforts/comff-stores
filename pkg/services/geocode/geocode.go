package geocode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	geoModels "github.com/comfforts/comff-stores/pkg/models/geo"
	"github.com/comfforts/comff-stores/pkg/services/cache"
	"github.com/comfforts/comff-stores/pkg/services/filestorage"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
	"github.com/comfforts/comff-stores/pkg/utils/geohash"
	"go.uber.org/zap"
)

type GeoCodeService struct {
	config       config.GeoCodeServiceConfig
	logger       *logging.AppLogger
	cache        *cache.CacheService
	cloudStorage *filestorage.CloudStorageClient
}

func unmarshalLPoint(p interface{}) (interface{}, error) {
	var point geohash.Point
	body, err := json.Marshal(p)
	if err != nil {
		appErr := errors.WrapError(err, cache.ERROR_MARSHALLING_CACHE_OBJECT)
		return point, appErr
	}

	err = json.Unmarshal(body, &point)
	if err != nil {
		appErr := errors.WrapError(err, cache.ERROR_UNMARSHALLING_CACHE_JSON)
		return point, appErr
	}
	return point, nil
}

func NewGeoCodeService(cfg config.GeoCodeServiceConfig, csc *filestorage.CloudStorageClient, logger *logging.AppLogger) (*GeoCodeService, error) {
	if cfg.GeocoderKey == "" || logger == nil {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}

	if cfg.Host == "" {
		cfg.Host = "https://maps.googleapis.com"
	}

	if cfg.Path == "" {
		cfg.Path = "/maps/api/geocode/json"
	}

	gcSrv := GeoCodeService{
		config:       cfg,
		logger:       logger,
		cloudStorage: csc,
	}

	if cfg.Cached {
		dataPath := filepath.Join(cfg.DataPath, fmt.Sprintf("%s.json", cache.CACHE_FILE_NAME))
		if _, err := fileUtils.FileStats(dataPath); err != nil {
			if err := gcSrv.downloadCache(); err != nil {
				logger.Error("error getting cache from storage", zap.Error(err))
			}
		}

		c, err := cache.NewCacheService(cfg.DataPath, logger, unmarshalLPoint)
		if err != nil {
			logger.Error("error setting up cache service", zap.Error(err))
			return nil, err
		}
		gcSrv.cache = c
	}
	return &gcSrv, nil
}

func (g *GeoCodeService) Geocode(ctx context.Context, postalCode, countryCode string) (*geohash.Point, error) {
	if ctx == nil {
		g.logger.Error("context is nil", zap.Error(errors.ErrNilContext))
		return nil, errors.ErrNilContext
	}

	if g.config.Cached {
		point, exp, err := g.getFromCache(postalCode)
		if err == nil {
			g.logger.Debug("returning cached value", zap.String("postalCode", postalCode), zap.Any("exp", exp))
			return point, nil
		} else {
			g.logger.Error("geocoder cache get error", zap.Error(err), zap.String("postalCode", postalCode))
		}
	}

	if countryCode == "" {
		countryCode = "US"
	}

	url := g.postalCodeURL(countryCode, postalCode)
	r, err := http.Get(url)
	if err != nil {
		g.logger.Error("geocoder request error", zap.Error(err), zap.String("postalCode", postalCode))
		return nil, err
	}
	defer r.Body.Close()

	var results geoModels.GeocoderResults
	err = json.NewDecoder(r.Body).Decode(&results)
	if err != nil || &results == (*geoModels.GeocoderResults)(nil) || len(results.Results) < 1 {
		g.logger.Error(errors.ERROR_GEOCODING_POSTALCODE, zap.Error(err), zap.String("postalCode", postalCode))
		return nil, errors.NewAppError(errors.ERROR_GEOCODING_POSTALCODE)
	}
	lat, long := results.Results[0].Geometry.Location.Lat, results.Results[0].Geometry.Location.Lng

	geoPoint := geohash.Point{
		Latitude:  lat,
		Longitude: long,
	}

	if g.config.Cached {
		err = g.setInCache(postalCode, geoPoint)
		if err != nil {
			g.logger.Error("geocoder cache set error", zap.Error(err), zap.String("postalCode", postalCode))
		}
	}

	return &geoPoint, nil
}

func (g *GeoCodeService) Clear() {
	g.logger.Info("cleaning up geo code data structures")
	if g.config.Cached && g.cache.Updated() {
		err := g.cache.SaveFile()
		if err != nil {
			g.logger.Error("error saving geocoder cache", zap.Error(err))
		} else {
			err = g.uploadCache()
			if err != nil {
				g.logger.Error("error uploading geocoder cache", zap.Error(err))
			}
		}
	}
}

func (g *GeoCodeService) postalCodeURL(countryCode, postalCode string) string {
	return fmt.Sprintf("%s%s?components=country:%s|postal_code:%s&sensor=false&key=%s", g.config.Host, g.config.Path, countryCode, postalCode, g.config.GeocoderKey)
}

func (g *GeoCodeService) getFromCache(postalCode string) (*geohash.Point, time.Time, error) {
	val, exp, err := g.cache.Get(postalCode)
	if err != nil {
		g.logger.Error(cache.ERROR_GET_CACHE, zap.Error(err), zap.String("postalCode", postalCode))
		return nil, exp, err
	}
	point, ok := val.(geohash.Point)
	if !ok {
		g.logger.Error("error getting cache point value", zap.Error(cache.ErrGetCache), zap.String("postalCode", postalCode))
		return nil, exp, cache.ErrGetCache
	}
	return &point, exp, nil
}

func (g *GeoCodeService) setInCache(postalCode string, point geohash.Point) error {
	err := g.cache.Set(postalCode, point, geoModels.OneYear)
	if err != nil {
		g.logger.Error(cache.ERROR_SET_CACHE, zap.Error(err), zap.String("postalCode", postalCode))
		return err
	}
	return nil
}

func (g *GeoCodeService) uploadCache() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataPath := filepath.Join(g.config.DataPath, fmt.Sprintf("%s.json", cache.CACHE_FILE_NAME))
	fStats, err := fileUtils.FileStats(dataPath)
	if err != nil {
		g.logger.Error("error accessing file", zap.Error(err), zap.String("filepath", dataPath))
		return errors.WrapError(err, fileUtils.ERROR_NO_FILE, dataPath)
	}
	fmod := fStats.ModTime().Unix()
	g.logger.Info("file mod time", zap.Int64("modtime", fmod), zap.String("filepath", dataPath))

	file, err := os.Open(dataPath)
	if err != nil {
		g.logger.Error("error accessing file", zap.Error(err), zap.String("filepath", dataPath))
		return errors.WrapError(err, fileUtils.ERROR_NO_FILE, dataPath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			g.logger.Error("error closing file", zap.Error(err), zap.String("filepath", dataPath))
		}
	}()

	cfr, err := filestorage.NewCloudFileRequest(g.config.BucketName, filepath.Base(dataPath), filepath.Dir(dataPath), fmod)
	if err != nil {
		g.logger.Error("error creating request", zap.Error(err), zap.String("filepath", dataPath))
		return err
	}

	n, err := g.cloudStorage.UploadFile(ctx, file, cfr)
	if err != nil {
		g.logger.Error("error uploading file", zap.Error(err))
		return err
	}
	g.logger.Info("uploaded file", zap.String("file", filepath.Base(dataPath)), zap.String("path", filepath.Dir(dataPath)), zap.Int64("bytes", n))
	return nil
}

func (g *GeoCodeService) downloadCache() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataPath := filepath.Join(g.config.DataPath, fmt.Sprintf("%s.json", cache.CACHE_FILE_NAME))
	var fmod int64
	fStats, err := fileUtils.FileStats(dataPath)
	if err != nil {
		g.logger.Error("error accessing file", zap.Error(err), zap.String("filepath", dataPath))
	} else {
		fmod := fStats.ModTime().Unix()
		g.logger.Info("file mod time", zap.Int64("modtime", fmod), zap.String("filepath", dataPath))
	}

	err = fileUtils.CreateDirectory(dataPath)
	if err != nil {
		g.logger.Error("error creating data directory", zap.Error(err), zap.String("filepath", dataPath))
		return err
	}

	f, err := os.Create(dataPath)
	if err != nil {
		g.logger.Error("error creating file", zap.Error(err), zap.String("filepath", dataPath))
		return errors.WrapError(err, fileUtils.ERROR_CREATING_FILE, dataPath)
	}
	defer func() {
		if err := f.Close(); err != nil {
			g.logger.Error("error closing file", zap.Error(err), zap.String("filepath", dataPath))
		}
	}()

	cfr, err := filestorage.NewCloudFileRequest(g.config.BucketName, filepath.Base(dataPath), filepath.Dir(dataPath), fmod)
	if err != nil {
		g.logger.Error("error creating request", zap.Error(err), zap.String("filepath", dataPath))
		return err
	}

	n, err := g.cloudStorage.DownloadFile(ctx, f, cfr)
	if err != nil {
		g.logger.Error("error downloading file", zap.Error(err), zap.String("filepath", dataPath))
		return err
	}
	g.logger.Info("downloaded file", zap.String("file", filepath.Base(dataPath)), zap.String("path", filepath.Dir(dataPath)), zap.Int64("bytes", n))
	return nil
}
