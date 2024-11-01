package main

import (
	"context"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/comfforts/logger"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/internal/config"
)

const SERVICE_PORT = 50051
const SERVICE_DOMAIN = "127.0.0.1"

func main() {
	// initialize app logger instance
	logCfg := &logger.AppLoggerConfig{
		FilePath: filepath.Join("logs", "client.log"),
		Level:    zapcore.DebugLevel,
		Name:     "comff-stores-client",
	}
	logger := logger.NewAppLogger(logCfg)

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
	testGetServers(client, logger)

	// ok := testStoreUpload(client, logger)
	// if ok {
	// 	testSearchStore(client, logger)

	// 	id := testAddStore(client, logger)
	// 	testGetStore(client, logger, id)
	// }
}

func testGetServers(client api.StoresClient, logger logger.AppLogger) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req := &api.GetServersRequest{}
	resp, err := client.GetServers(ctx, req)
	if err != nil {
		logger.Fatal("error getting servers", zap.Error(err))
	}
	for k, v := range resp.Servers {
		logger.Info("Server response", zap.Any("num", k), zap.Any("server", v))
	}
}

func testStoreUpload(client api.StoresClient, logger logger.AppLogger) bool {
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

func testAddStore(client api.StoresClient, logger logger.AppLogger) string {
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

func testGetStore(client api.StoresClient, logger logger.AppLogger, id string) {
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

func testSearchStore(client api.StoresClient, logger logger.AppLogger) {
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
