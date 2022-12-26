package test

import (
	"encoding/json"
	"fmt"
	"os"

	fileModels "github.com/comfforts/comff-stores/pkg/models/file"
)

func CreateJSONFile(dir, name string) (string, error) {
	fPath := fmt.Sprintf("%s.json", name)
	if dir != "" {
		fPath = fmt.Sprintf("%s/%s", dir, fPath)
	}
	items := []fileModels.JSONMapper{
		{
			"city":      "Hong Kong",
			"name":      "Plaza Hollywood",
			"country":   "CN",
			"longitude": 114.20169067382812,
			"latitude":  22.340700149536133,
			"store_id":  1,
		},
		{
			"city":      "Hong Kong",
			"name":      "Exchange Square",
			"country":   "CN",
			"longitude": 114.15818786621094,
			"latitude":  22.283939361572266,
			"store_id":  6,
		},
		{
			"city":      "Kowloon",
			"name":      "Telford Plaza",
			"country":   "CN",
			"longitude": 114.21343994140625,
			"latitude":  22.3228702545166,
			"store_id":  8,
		},
	}

	file, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
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
	item := fileModels.JSONMapper{
		"city":      "Hong Kong",
		"name":      "Plaza Hollywood",
		"country":   "CN",
		"longitude": 114.20169067382812,
		"latitude":  22.340700149536133,
		"store_id":  1,
	}

	file, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(item)
	if err != nil {
		return "", err
	}
	return fPath, nil
}
