package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/pkg/constants"
	"github.com/comfforts/comff-stores/pkg/logging"
)

func main() {
	// initialize app logger instance
	logCfg := &logging.AppLoggerConfig{
		FilePath: "logs/client.log",
		Level:    zapcore.DebugLevel,
	}
	logger := logging.NewAppLogger(nil, logCfg)

	servicePort := fmt.Sprintf(":%d", constants.SERVICE_PORT)
	conn, err := grpc.Dial(servicePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("client failed to connect", zap.Error(err))
	}
	defer conn.Close()

	client := api.NewStoresClient(conn)

	// id := testAddStore(client, logger)
	// testGetStore(client, logger, id)

	testSearchStore(client, logger)
}

// func testAddStore(client api.StoresClient, logger *logging.AppLogger) string {
// 	ctx := context.Background()

// 	storeId, name, org, city := uint64(1), "Plaza Hollywood", "starbucks", "Hong Kong"
// 	addStoreReq := &api.AddStoreRequest{
// 		Name:      name,
// 		Org:       org,
// 		City:      city,
// 		Country:   "CN",
// 		Longitude: 114.20169067382812,
// 		Latitude:  22.340700149536133,
// 		StoreId:   storeId,
// 	}

// 	addStoreRes, err := client.AddStore(ctx, addStoreReq)
// 	if err != nil {
// 		logger.Fatal("error when calling AddStore", zap.Error(err))
// 		return ""
// 	}
// 	logger.Info("AddStore response from store server", zap.Any("addStoreResp", addStoreRes))
// 	return addStoreRes.Store.Id
// }

// func testGetStore(client api.StoresClient, logger *logging.AppLogger, id string) {
// 	ctx := context.Background()

// 	getStoreReq := &api.GetStoreRequest{
// 		Id: id,
// 	}

// 	getStoreRes, err := client.GetStore(ctx, getStoreReq)
// 	if err != nil {
// 		logger.Fatal("error when calling GetStore", zap.Error(err))
// 	}
// 	logger.Info("GetStore response from store server", zap.Any("getStoreResp", getStoreRes))
// }

func testSearchStore(client api.StoresClient, logger *logging.AppLogger) {
	ctx := context.Background()

	searchStoreReq := &api.SearchStoreRequest{
		PostalCode: "94952",
		Distance:   5,
	}

	searchStoreRes, err := client.SearchStore(ctx, searchStoreReq)
	if err != nil {
		logger.Fatal("error when calling SearchStore", zap.Error(err))
	}
	logger.Info("SearchStore response from store server", zap.Any("searchStoreRes", searchStoreRes))
}
