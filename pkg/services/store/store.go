package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	storeModels "github.com/comfforts/comff-stores/pkg/models/store"
	"github.com/comfforts/comff-stores/pkg/utils/geohash"
	"go.uber.org/zap"

	"gitlab.com/xerra/common/vincenty"
)

type StoreService struct {
	mu      sync.RWMutex
	logger  *logging.AppLogger
	stores  map[string]*storeModels.Store
	hashMap map[string][]string
	count   int
	ready   bool
}

func NewStoreService(logger *logging.AppLogger) *StoreService {
	ss := &StoreService{
		logger:  logger,
		stores:  map[string]*storeModels.Store{},
		hashMap: map[string][]string{},
		count:   0,
		ready:   false,
	}

	return ss
}

func (ss *StoreService) AddStore(ctx context.Context, s *storeModels.Store) (*storeModels.Store, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	id, err := BuildId(s.Latitude, s.Longitude, s.Org)
	if err != nil {
		ss.logger.Error(errors.ERROR_ENCODING_ID, zap.String("org", s.Org), zap.Float64("latitude", s.Latitude), zap.Float64("longitude", s.Longitude))
		return nil, errors.WrapError(err, errors.ERROR_ENCODING_LAT_LONG)
	}

	ssLookup := ss.lookup(id)
	if ssLookup != nil {
		if s.StoreId != ssLookup.StoreId {
			id, err = BuildIdC(id, fmt.Sprintf("%d", s.StoreId), "")
			if err != nil {
				ss.logger.Error(errors.ERROR_ENCODING_ID, zap.Error(err), zap.Uint64("storeId", s.StoreId))
				return nil, errors.WrapError(err, errors.ERROR_ENCODING_ID)
			}
		} else if s.Name != ssLookup.Name {
			id, err = BuildIdC(id, "", s.Name)
			if err != nil {
				ss.logger.Error(errors.ERROR_ENCODING_ID, zap.Error(err), zap.String("name", s.Name))
				return nil, errors.WrapError(err, errors.ERROR_ENCODING_ID)
			}
		} else {
			ss.logger.Error(errors.ERROR_STORE_ID_ALREADY_EXISTS, zap.String("id", id))
			return nil, errors.NewAppError(errors.ERROR_STORE_ID_ALREADY_EXISTS)
		}
	}

	hashKey, err := geohash.Encode(s.Latitude, s.Longitude, 8)
	if err != nil {
		ss.logger.Error(errors.ERROR_ENCODING_LAT_LONG, zap.Float64("latitude", s.Latitude), zap.Float64("longitude", s.Longitude))
		return nil, errors.WrapError(err, errors.ERROR_ENCODING_LAT_LONG)
	}
	hashStoreIDs, ok := ss.hashMap[hashKey]
	if !ok {
		hashStoreIDs = []string{}
	}
	s.ID = id
	hashStoreIDs = append(hashStoreIDs, s.ID)
	ss.hashMap[hashKey] = hashStoreIDs

	ss.stores[s.ID] = s
	ss.count++

	return s, nil
}

func (ss *StoreService) GetStore(ctx context.Context, id string) (*storeModels.Store, error) {
	s := ss.lookup(id)
	if s == nil {
		ss.logger.Error(errors.ERROR_NO_STORE_FOUND_FOR_ID, zap.String("id", id))
		return nil, errors.WrapError(errors.ErrNotFound, errors.ERROR_NO_STORE_FOUND_FOR_ID)
	}
	return s, nil
}

func (ss *StoreService) GetStoresForGeoPoint(ctx context.Context, lat, long float64, dist int) ([]*storeModels.StoreGeo, error) {
	ss.logger.Debug("getting stores for geopoint", zap.Float64("latitude", lat), zap.Float64("longitude", long), zap.Int("distance", dist))
	ids, err := ss.getStoreIdsForLatLong(ctx, lat, long)
	if err != nil {
		return nil, err
	}
	ss.logger.Debug("found stores", zap.Int("numOfStores", len(ids)), zap.Float64("latitude", lat), zap.Float64("longitude", long))

	stores := []*storeModels.StoreGeo{}
	// origin := haversine.Point{Lat: lat, Lon: long}
	origin := vincenty.LatLng{Latitude: lat, Longitude: long}
	for _, v := range ids {
		store, err := ss.GetStore(ctx, v)
		if err != nil {
			ss.logger.Error(errors.ERROR_NO_STORE_FOUND_FOR_ID, zap.String("id", v))
		}
		// pos := haversine.Point{Lat: store.Latitude, Lon: store.Longitude}
		pos := vincenty.LatLng{Latitude: store.Latitude, Longitude: store.Longitude}
		// d := haversine.Distance(origin, pos)
		d := vincenty.Inverse(origin, pos)
		// if float64(d) <= dist*1000 {
		if d.Kilometers() <= float64(dist) {
			stGeo := &storeModels.StoreGeo{
				Store:    store,
				Distance: d.Kilometers(),
			}
			stores = append(stores, stGeo)
		}
	}
	ss.logger.Debug("returning stores", zap.Int("numOfStores", len(stores)), zap.Float64("latitude", lat), zap.Float64("longitude", long), zap.Int("distance", dist))
	return stores, nil
}

func (ss *StoreService) GetStoreStats() storeModels.StoreStats {
	return storeModels.StoreStats{
		Ready:     ss.ready,
		Count:     ss.count,
		HashCount: len(ss.hashMap),
	}
}

func (ss *StoreService) GetAllStores() []*storeModels.Store {
	stores := []*storeModels.Store{}
	for _, v := range ss.stores {
		stores = append(stores, v)
	}
	return stores
}

func (ss *StoreService) SetReady(ctx context.Context, ready bool) {
	ss.ready = ready
}

func (ss *StoreService) Clear() {
	ss.logger.Info("cleaning up store data structures")
	ss.count = 0
	ss.stores = map[string]*storeModels.Store{}
	ss.hashMap = map[string][]string{}
	ss.ready = false
}

func MapResultToStore(r map[string]interface{}) (*storeModels.Store, error) {
	storeJson, err := json.Marshal(r)
	if err != nil {
		return nil, errors.WrapError(err, errors.ERROR_MARSHALLING_RESULT)
	}

	var s storeModels.Store
	err = json.Unmarshal(storeJson, &s)
	if err != nil {
		return nil, errors.WrapError(err, errors.ERROR_UNMARSHALLING_STORE_JSON)
	}
	return &s, nil
}

func (ss *StoreService) getStoreIdsForLatLong(ctx context.Context, lat, long float64) ([]string, error) {
	hashKey, err := geohash.Encode(lat, long, 8)
	if err != nil {
		ss.logger.Error(errors.ERROR_ENCODING_LAT_LONG, zap.Float64("latitude", lat), zap.Float64("longitude", long))
		return nil, errors.WrapError(err, errors.ERROR_ENCODING_LAT_LONG)
	}
	ss.logger.Debug("created hash key", zap.String("hashKey", hashKey), zap.Float64("latitude", lat), zap.Float64("longitude", long))
	ids, ok := ss.hashMap[hashKey]
	if !ok || len(ids) < 1 {
		ss.logger.Error(errors.ERROR_NO_STORE_FOUND, zap.Float64("latitude", lat), zap.Float64("longitude", long))
		return nil, errors.WrapError(err, errors.ERROR_NO_STORE_FOUND)
	}
	return ids, nil
}

func (ss *StoreService) lookup(k string) *storeModels.Store {
	v, ok := ss.stores[k]
	if !ok {
		return nil
	}
	return v
}

func BuildId(lat, long float64, org string) (string, error) {
	hPart, err := geohash.Encode(lat, long, 12)
	if err != nil {
		return "", errors.WrapError(err, errors.ERROR_ENCODING_LAT_LONG)
	}
	oPart := org
	if len(org) > 6 {
		oPart = org[0:6]
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", oPart, hPart)))
	return encoded, nil
}

func BuildIdC(id string, storeId, name string) (string, error) {
	if storeId == "" && name == "" {
		return "", errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}
	if storeId != "" {
		return fmt.Sprintf("%s%s", id, storeId), nil
	}
	if len(name) > 6 {
		name = name[0:6]
	}
	return fmt.Sprintf("%s%s", id, name), nil
}
