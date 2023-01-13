package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

const CACHE_FILE_NAME = "cache"

const (
	ERROR_SET_CACHE                string = "error adding key/value to cache"
	ERROR_GET_CACHE                string = "error getting key/value from cache"
	ERROR_CREATING_CACHE_DIR       string = "error creating cache directory"
	ERROR_GETTING_CACHE_FILE       string = "error getting cache file"
	ERROR_SAVING_CACHE_FILE        string = "error saving cache file"
	ERROR_OPENING_CACHE_FILE       string = "error opening cache file"
	ERROR_LOADING_CACHE_FILE       string = "error loading cache file"
	ERROR_MARSHALLING_CACHE_OBJECT string = "error marshalling object to json"
	ERROR_UNMARSHALLING_CACHE_JSON string = "error unmarshalling json to struct"

	VALUE_ADDED         = "added value to cache"
	RETURNING_VALUE     = "returning value for given key"
	KEY_DELETED         = "deleted value with given key"
	DELETED_EXPIRED     = "deleted expired cache values"
	RETURNING_COUNT     = "returning item count"
	RETURNING_ALL_ITEMS = "returning all items"
	CACHE_FLUSHED       = "cache flushed"
)

var (
	ErrSetCache      = errors.NewAppError(ERROR_SET_CACHE)
	ErrGetCache      = errors.NewAppError(ERROR_GET_CACHE)
	ErrGetCacheFile  = errors.NewAppError(ERROR_GETTING_CACHE_FILE)
	ErrSaveCacheFile = errors.NewAppError(ERROR_SAVING_CACHE_FILE)
)

type MarshalFn func(p interface{}) (interface{}, error)

type CacheService struct {
	DataDir   string
	cache     *cache.Cache
	logger    *logging.AppLogger
	marshalFn MarshalFn
	loadedAt  int64
	updatedAt int64
}

func NewCacheService(dataDir string, logger *logging.AppLogger, marshalFn MarshalFn) (*CacheService, error) {
	if dataDir == "" || logger == nil {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}

	default_expiration := 5 * time.Minute
	cleanup_interval := 10 * time.Minute
	c := cache.New(default_expiration, cleanup_interval)

	cacheService := &CacheService{
		DataDir:   dataDir,
		cache:     c,
		logger:    logger,
		marshalFn: marshalFn,
	}

	err := cacheService.LoadFile()
	if err != nil {
		logger.Error(ERROR_LOADING_CACHE_FILE, zap.Error(err), zap.String("cacheDir", dataDir))
		logger.Info("starting with fresh cache")
	}

	return cacheService, nil
}

func (c *CacheService) SetLoadedAt(at int64) {
	c.loadedAt = at
	c.updatedAt = at
}

func (c *CacheService) Updated() bool {
	c.logger.Info("cache file", zap.Int64("loadedAt", c.loadedAt), zap.Int64("updatedAt", c.updatedAt))
	return c.updatedAt > c.loadedAt
}

func (c *CacheService) Set(key string, value interface{}, d time.Duration) error {
	err := c.cache.Add(key, value, d)
	if err != nil {
		c.logger.Error(ERROR_SET_CACHE, zap.Error(err), zap.String("cacheDir", c.DataDir))
		return errors.WrapError(err, ERROR_SET_CACHE)
	}
	c.updatedAt = time.Now().Unix()
	c.logger.Debug(VALUE_ADDED, zap.String("key", key), zap.String("cacheDir", c.DataDir))
	return nil
}

func (c *CacheService) Get(key string) (interface{}, time.Time, error) {
	val, exp, ok := c.cache.GetWithExpiration(key)
	if !ok {
		c.logger.Error(ERROR_GET_CACHE, zap.Error(ErrGetCache), zap.String("cacheDir", c.DataDir))
		return nil, time.Time{}, ErrGetCache
	}
	c.logger.Debug(RETURNING_VALUE, zap.String("key", key), zap.String("cacheDir", c.DataDir))
	return val, exp, nil
}

func (c *CacheService) Delete(key string) {
	c.cache.Delete(key)
	c.updatedAt = time.Now().Unix()
	c.logger.Debug(KEY_DELETED, zap.String("key", key), zap.String("cacheDir", c.DataDir))
}

func (c *CacheService) DeleteExpired() {
	c.cache.DeleteExpired()
	c.logger.Debug(DELETED_EXPIRED, zap.String("cacheDir", c.DataDir))
}

func (c *CacheService) ItemCount() int {
	count := c.cache.ItemCount()
	c.logger.Info(RETURNING_COUNT, zap.String("cacheDir", c.DataDir))
	return count
}

func (c *CacheService) Items() map[string]cache.Item {
	items := c.cache.Items()
	c.logger.Info(RETURNING_ALL_ITEMS, zap.String("cacheDir", c.DataDir))
	return items
}

func (c *CacheService) Clear() {
	c.cache.Flush()
	c.logger.Info(CACHE_FLUSHED, zap.String("cacheDir", c.DataDir))
}

func (c *CacheService) SaveFile() error {
	filePath := filepath.Join(c.DataDir, fmt.Sprintf("%s.json", CACHE_FILE_NAME))
	c.logger.Info("saving cache file", zap.String("filePath", filePath))

	err := fileUtils.CreateDirectory(filePath)
	if err != nil {
		c.logger.Error(ERROR_CREATING_CACHE_DIR, zap.Error(err), zap.String("filePath", filePath))
		return ErrSaveCacheFile
	}

	file, err := os.Create(filePath)
	if err != nil {
		c.logger.Error(ERROR_GETTING_CACHE_FILE, zap.Error(err), zap.String("filePath", filePath))
		return ErrGetCacheFile
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	items := c.cache.Items()
	err = encoder.Encode(items)
	if err != nil {
		c.logger.Error(ERROR_SAVING_CACHE_FILE, zap.Error(err), zap.String("filePath", filePath))
		return ErrSaveCacheFile
	}
	c.logger.Info("cache file saved", zap.String("filePath", filePath))
	return nil
}

func (c *CacheService) LoadFile() error {
	filePath := filepath.Join(c.DataDir, fmt.Sprintf("%s.json", CACHE_FILE_NAME))
	c.logger.Info("loading cache file", zap.String("filePath", filePath))
	file, err := os.Open(filePath)
	if err != nil {
		c.logger.Error(ERROR_OPENING_CACHE_FILE, zap.Error(err))
		return err
	}

	err = c.Load(file)
	if err != nil {
		c.logger.Error(ERROR_LOADING_CACHE_FILE, zap.Error(err), zap.String("filePath", filePath))
		return err
	}
	return nil
}

func (c *CacheService) Load(r io.Reader) error {
	dec := json.NewDecoder(r)
	items := map[string]cache.Item{}
	err := dec.Decode(&items)
	if err == nil {
		for k, v := range items {
			if !v.Expired() {
				obj, err := c.marshalFn(v.Object)
				if err != nil {
					c.logger.Error("error marshalling file object", zap.Error(err), zap.String("cacheDir", c.DataDir))
				} else {
					err = c.Set(k, obj, 5*time.Hour)
					if err != nil {
						c.logger.Error(ERROR_SET_CACHE, zap.Error(err), zap.String("cacheDir", c.DataDir))
					} else {
						c.logger.Debug("cache item loaded", zap.String("cacheDir", c.DataDir), zap.String("key", k), zap.Any("value", obj), zap.Any("exp", v.Expiration))
					}
				}
			}
		}
	}
	c.SetLoadedAt(time.Now().Unix())
	c.logger.Info("cache file loaded", zap.Int64("loadedAt", c.loadedAt), zap.Int64("updatedAt", c.updatedAt))
	return err
}
