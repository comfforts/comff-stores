package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
)

const (
	ERROR_SET_CACHE                string = "error adding key/value to cache"
	ERROR_GET_CACHE                string = "error getting key/value from cache"
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

type CacheService struct {
	name      string
	cache     *cache.Cache
	config    *config.Configuration
	logger    *logging.AppLogger
	marshalFn MarshalFn
}

type MarshalFn func(p interface{}) (interface{}, error)

func NewCacheService(name string, config *config.Configuration, logger *logging.AppLogger, marshalFn MarshalFn) (*CacheService, error) {
	if name == "" || config == nil || logger == nil {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}
	default_expiration := 5 * time.Minute
	cleanup_interval := 10 * time.Minute
	c := cache.New(default_expiration, cleanup_interval)

	cacheService := &CacheService{
		name:      name,
		cache:     c,
		config:    config,
		logger:    logger,
		marshalFn: marshalFn,
	}

	err := cacheService.LoadFile()
	if err != nil {
		logger.Error(ERROR_LOADING_CACHE_FILE, zap.Error(err), zap.String("cacheName", name))
	}

	return cacheService, nil
}

func (c *CacheService) Set(key string, value interface{}, d time.Duration) error {
	err := c.cache.Add(key, value, d)
	if err != nil {
		c.logger.Error(ERROR_SET_CACHE, zap.Error(err), zap.String("cacheName", c.name))
		return errors.WrapError(err, ERROR_SET_CACHE)
	}
	c.logger.Info(VALUE_ADDED, zap.String("key", key), zap.String("cacheName", c.name))
	return nil
}

func (c *CacheService) Get(key string) (interface{}, time.Time, error) {
	val, exp, ok := c.cache.GetWithExpiration(key)
	if !ok {
		c.logger.Error(ERROR_GET_CACHE, zap.Error(ErrGetCache), zap.String("cacheName", c.name))
		return nil, time.Time{}, ErrGetCache
	}
	c.logger.Info(RETURNING_VALUE, zap.String("key", key), zap.String("cacheName", c.name))
	return val, exp, nil
}

func (c *CacheService) Delete(key string) {
	c.cache.Delete(key)
	c.logger.Info(KEY_DELETED, zap.String("key", key), zap.String("cacheName", c.name))
}

func (c *CacheService) DeleteExpired() {
	c.cache.DeleteExpired()
	c.logger.Info(DELETED_EXPIRED, zap.String("cacheName", c.name))
}

func (c *CacheService) ItemCount() int {
	count := c.cache.ItemCount()
	c.logger.Info(RETURNING_COUNT, zap.String("cacheName", c.name))
	return count
}

func (c *CacheService) Items() map[string]cache.Item {
	items := c.cache.Items()
	c.logger.Info(RETURNING_ALL_ITEMS, zap.String("cacheName", c.name))
	return items
}

func (c *CacheService) Clear() {
	c.cache.Flush()
	c.logger.Info(CACHE_FLUSHED, zap.String("cacheName", c.name))
}

func (c *CacheService) SaveFile() error {
	fileName := fmt.Sprintf("%s.json", strings.ToLower(c.name))
	c.logger.Info("saving cache file", zap.String("cacheName", c.name))
	file, err := os.Create(fileName)
	if err != nil {
		c.logger.Error(ERROR_GETTING_CACHE_FILE, zap.Error(err), zap.String("cacheName", c.name))
		return ErrGetCacheFile
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	items := c.cache.Items()
	err = encoder.Encode(items)
	if err != nil {
		c.logger.Error(ERROR_SAVING_CACHE_FILE, zap.Error(err), zap.String("cacheName", c.name))
		return ErrSaveCacheFile
	}
	c.logger.Info("cache file saved", zap.String("cacheName", c.name))
	return nil
}

func (c *CacheService) LoadFile() error {
	fileName := fmt.Sprintf("%s.json", strings.ToLower(c.name))
	c.logger.Info("loading cache file", zap.String("cacheName", c.name))
	file, err := os.Open(fileName)
	if err != nil {
		c.logger.Error(ERROR_OPENING_CACHE_FILE, zap.Error(err), zap.String("cacheName", c.name))
		return err
	}

	err = c.Load(file)
	if err != nil {
		c.logger.Error(ERROR_LOADING_CACHE_FILE, zap.Error(err), zap.String("cacheName", c.name))
		return err
	}
	c.logger.Info("cache file loaded", zap.String("cacheName", c.name))
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
					c.logger.Error("error marshalling file object", zap.Error(err), zap.String("cacheName", c.name))
				} else {
					err = c.Set(k, obj, 5*time.Hour)
					if err != nil {
						c.logger.Error(ERROR_SET_CACHE, zap.Error(err), zap.String("cacheName", c.name))
					} else {
						c.logger.Info("cache item loaded", zap.String("cacheName", c.name), zap.String("key", k), zap.Any("value", obj), zap.Any("exp", v.Expiration))
					}
				}
			}
		}
	}
	return err
}
