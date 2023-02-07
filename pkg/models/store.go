package models

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/comfforts/errors"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/pkg/constants"
)

type Stores interface {
	AddStore(ctx context.Context, s *Store) (*Store, error)
	GetStore(ctx context.Context, id string) (*Store, error)
	GetStoresForGeoPoint(ctx context.Context, lat, long float64, dist int) ([]*StoreGeo, error)
	Reader(ctx context.Context, dataDir string) (*os.File, error)
	GetStoreStats() StoreStats
	SetReady(ctx context.Context, ready bool)
	Close() error
}

type Store struct {
	ID        string    `json:"id,omitempty"`
	StoreId   uint64    `json:"store_id"`
	Name      string    `json:"name"`
	Org       string    `json:"org"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Created   time.Time `json:"created,omitempty"`
}

type StoreGeo struct {
	Store    *Store
	Distance float64
}

type StoreStats struct {
	Count     int
	HashCount int
	Ready     bool
}

func MapResultToStore(r JSONMapper) (*Store, error) {
	storeJson, err := json.Marshal(r)
	if err != nil {
		return nil, errors.WrapError(err, constants.ERROR_MARSHALLING_RESULT)
	}

	var s Store
	err = json.Unmarshal(storeJson, &s)
	if err != nil {
		return nil, errors.WrapError(err, constants.ERROR_UNMARSHALLING_STORE_JSON)
	}
	return &s, nil
}

func MapStoreToJSON(s *Store) JSONMapper {
	return JSONMapper{
		"id":        s.ID,
		"name":      s.Name,
		"org":       s.Org,
		"city":      s.City,
		"country":   s.Country,
		"longitude": s.Longitude,
		"latitude":  s.Latitude,
		"store_id":  s.StoreId,
	}
}

func MapProtoToStore(s *api.Store) *Store {
	return &Store{
		ID:        s.Id,
		StoreId:   s.StoreId,
		Name:      s.Name,
		Org:       s.Org,
		Longitude: float64(s.Longitude),
		Latitude:  float64(s.Latitude),
		City:      s.City,
		Country:   s.Country,
	}
}

func MapStoreRequestToStore(s *api.AddStoreRequest) *Store {
	return &Store{
		StoreId:   s.StoreId,
		Name:      s.Name,
		Org:       s.Org,
		Longitude: float64(s.Longitude),
		Latitude:  float64(s.Latitude),
		City:      s.City,
		Country:   s.Country,
	}
}

func MapStoreModelToResponse(store *Store) *api.Store {
	return &api.Store{
		Id:        store.ID,
		Name:      store.Name,
		Org:       store.Org,
		Longitude: float32(store.Longitude),
		Latitude:  float32(store.Latitude),
		City:      store.City,
		Country:   store.Country,
		StoreId:   store.StoreId,
	}
}

func MapStoreListToResponse(sts []*StoreGeo) []*api.StoreGeo {
	stores := []*api.StoreGeo{}

	for _, st := range sts {
		stGeo := &api.StoreGeo{
			Store:    MapStoreModelToResponse(st.Store),
			Distance: float32(st.Distance),
		}
		stores = append(stores, stGeo)
	}

	return stores
}
