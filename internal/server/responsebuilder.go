package server

import (
	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/pkg/services/store"
)

func MapStoreModelToResponse(store *store.Store) *api.Store {
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

func MapStoreListToResponse(sts []*store.StoreGeo) []*api.StoreGeo {
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
