package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/internal/config"
	"github.com/comfforts/comff-stores/pkg/logging"
)

const SERVICE_PORT = 50051
const SERVICE_DOMAIN = "192.168.68.100"

func main() {
	// initialize app logger instance
	logCfg := &logging.AppLoggerConfig{
		FilePath: "logs/client.log",
		Level:    zapcore.DebugLevel,
	}
	logger := logging.NewAppLogger(nil, logCfg)

	tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: config.CertFile(config.ClientCertFile),
		KeyFile:  config.CertFile(config.ClientKeyFile),
		CAFile:   config.CertFile(config.CAFile),
		Server:   false,
	})
	if err != nil {
		logger.Fatal("error setting client TLS", zap.Error(err))
	}
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}

	servicePort := fmt.Sprintf("%s:%d", SERVICE_DOMAIN, SERVICE_PORT)
	conn, err := grpc.Dial(servicePort, opts...)
	if err != nil {
		logger.Fatal("client failed to connect", zap.Error(err))
	}
	defer conn.Close()

	client := api.NewStoresClient(conn)

	ok := testStoreUpload(client, logger)
	if ok {
		testSearchStore(client, logger)

		id := testAddStore(client, logger)
		testGetStore(client, logger, id)
	}
}

func testStoreUpload(client api.StoresClient, logger *logging.AppLogger) bool {
	ctx := context.Background()

	req := &api.StoreUploadRequest{
		FileName: "starbucks.json",
	}

	resp, err := client.StoreUpload(ctx, req)
	if err != nil {
		logger.Fatal("error uploading stores", zap.Error(err))
		return false
	}
	logger.Info("storeUpload response", zap.Any("resp", resp))
	return resp.Ok
}

func testAddStore(client api.StoresClient, logger *logging.AppLogger) string {
	ctx := context.Background()

	storeId, name, org, city := uint64(1), "Plaza Hollywood", "subway", "Hong Kong"
	addStoreReq := &api.AddStoreRequest{
		Name:      name,
		Org:       org,
		City:      city,
		Country:   "CN",
		Longitude: 114.22169067382812,
		Latitude:  22.347200149536133,
		StoreId:   storeId,
	}

	addStoreRes, err := client.AddStore(ctx, addStoreReq)
	if err != nil {
		logger.Fatal("error when calling AddStore", zap.Error(err))
		return ""
	}
	logger.Info("AddStore response from store server", zap.Any("addStoreResp", addStoreRes))
	return addStoreRes.Store.Id
}

func testGetStore(client api.StoresClient, logger *logging.AppLogger, id string) {
	ctx := context.Background()

	getStoreReq := &api.GetStoreRequest{
		Id: id,
	}

	getStoreRes, err := client.GetStore(ctx, getStoreReq)
	if err != nil {
		logger.Fatal("error when calling GetStore", zap.Error(err))
	}
	logger.Info("GetStore response from store server", zap.Any("getStoreResp", getStoreRes))
}

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
