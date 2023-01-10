package test

import (
	"encoding/json"
	"fmt"
	"os"

	api "github.com/comfforts/comff-stores/api/v1"
	fileModels "github.com/comfforts/comff-stores/pkg/models/file"
	storeModels "github.com/comfforts/comff-stores/pkg/models/store"
)

func CreateJSONFile(dir, name string) (string, error) {
	fPath := fmt.Sprintf("%s.json", name)
	if dir != "" {
		fPath = fmt.Sprintf("%s/%s", dir, fPath)
	}
	items := CreateStoreJSONList()

	f, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(items)
	if err != nil {
		return "", err
	}
	return fPath, nil
}

func CreateSingleJSONFile(dir, name string) (string, error) {
	fPath := fmt.Sprintf("%s.json", name)
	if dir != "" {
		fPath = fmt.Sprintf("%s/%s", dir, fPath)
	}

	item := CreateStoreJSON(uint64(1), "Mustum Bugdum", "starbucks", "Hong Kong", "CN")

	f, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(item)
	if err != nil {
		return "", err
	}
	return fPath, nil
}

func CreateStoreJSON(storeId uint64, name, org, city, country string) fileModels.JSONMapper {
	s := fileModels.JSONMapper{
		"name":      name,
		"org":       org,
		"city":      city,
		"country":   country,
		"longitude": 114.74169067382812,
		"latitude":  21.340700149536133,
		"store_id":  storeId,
	}
	return s
}

func CreateStoreJSONList() []fileModels.JSONMapper {
	items := []fileModels.JSONMapper{
		{
			"city":      "Hong Kong",
			"org":       "starbucks",
			"name":      "Plaza Hollywood",
			"country":   "CN",
			"longitude": 114.20169067382812,
			"latitude":  22.340700149536133,
			"store_id":  1,
		},
		{
			"city":      "Hong Kong",
			"org":       "starbucks",
			"name":      "Exchange Square",
			"country":   "CN",
			"longitude": 114.15818786621094,
			"latitude":  22.283939361572266,
			"store_id":  6,
		},
		{
			"city":      "Kowloon",
			"org":       "starbucks",
			"name":      "Telford Plaza",
			"country":   "CN",
			"longitude": 114.21343994140625,
			"latitude":  22.3228702545166,
			"store_id":  8,
		},
	}
	return items
}

func CreateStoreModel() *storeModels.Store {
	item := CreateStoreJSON(uint64(5), "Mustum Bugdum", "starbucks", "Hong Kong", "CN")
	st, err := storeModels.MapResultToStore(item)
	if err != nil {
		return nil
	}
	return st
}

func CreateStoreModelList() []*storeModels.Store {
	items := CreateStoreJSONList()
	list := []*storeModels.Store{}
	for _, v := range items {
		st, err := storeModels.MapResultToStore(v)
		if err == nil {
			list = append(list, st)
		}
	}
	return list
}

func CreateAddStoreRequest(storeId uint64, name, org, city, country string, lat, long float64) *api.AddStoreRequest {
	s := &api.AddStoreRequest{
		Name:      name,
		Org:       org,
		City:      city,
		Country:   country,
		Longitude: float32(long),
		Latitude:  float32(lat),
		StoreId:   storeId,
	}
	return s
}
