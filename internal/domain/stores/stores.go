package stores

import (
	"context"

	api "github.com/comfforts/comff-stores/api/stores/v1"
)

type StoresRepo interface {
	AddStore(ctx context.Context, store *Store) (string, error)
	GetStore(ctx context.Context, idHex string) (*Store, error)
	DeleteStore(ctx context.Context, idHex string) error
	UpdateStore(ctx context.Context, idHex string, params *UpdateStoreQuery) error
	SearchStores(ctx context.Context, params *SearchStoreQuery) ([]*Store, error)
	Close(ctx context.Context) error
}

type StoresService interface {
	AddStore(ctx context.Context, st *AddStoreParams) (string, error)
	GetStore(ctx context.Context, id string) (*Store, error)
	DeleteStore(ctx context.Context, id string) error
	UpdateStore(ctx context.Context, id string, params *UpdateStoreParams) error
	SearchStores(ctx context.Context, params *SearchStoreParams) ([]*Store, error)
}

type Store struct {
	ID        string `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string `bson:"name" json:"name"`
	Org       string `bson:"org" json:"org"`
	AddressId string `bson:"address_id" json:"address_id"`
}

type AddStoreParams struct {
	Name      string
	Org       string
	AddressId string
}

type UpdateStoreParams struct {
	Name      string
	Org       string
	AddressId string
}

type UpdateStoreQuery struct {
	Name      string
	Org       string
	AddressId string
}

type SearchStoreParams struct {
	Org        string
	Name       string
	AddressId  string
	AddressStr string
	Latitude   float64
	Longitude  float64
	Distance   uint32
}

type SearchStoreQuery struct {
	Org       string
	Name      string
	AddressId string
}

func MapToAddStoreParams(st *api.AddStoreRequest) *AddStoreParams {
	if st == nil {
		return nil
	}
	return &AddStoreParams{
		Name:      st.GetName(),
		Org:       st.GetOrg(),
		AddressId: st.GetAddressId(),
	}
}

func MapToStoreProto(store *Store) *api.Store {
	if store == nil {
		return nil
	}
	return &api.Store{
		Id:        store.ID,
		Name:      store.Name,
		Org:       store.Org,
		AddressId: store.AddressId,
	}
}

func MapToUpdateStoreParams(st *api.UpdateStoreRequest) *UpdateStoreParams {
	if st == nil {
		return nil
	}
	return &UpdateStoreParams{
		Name:      st.GetName(),
		Org:       st.GetOrg(),
		AddressId: st.GetAddressId(),
	}
}

func MapToSearchStoreParams(st *api.SearchStoreRequest) *SearchStoreParams {
	if st == nil {
		return nil
	}
	return &SearchStoreParams{
		Org:        st.GetOrg(),
		Name:       st.GetName(),
		AddressId:  st.GetAddressId(),
		AddressStr: st.GetAddressStr(),
		Latitude:   st.GetLatitude(),
		Longitude:  st.GetLongitude(),
		Distance:   st.GetDistance(),
	}
}
