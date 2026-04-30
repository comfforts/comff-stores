package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	config "github.com/comfforts/comff-config"
	"github.com/comfforts/logger"

	api "github.com/comfforts/comff-stores/api/stores/v1"
)

const SERVICE_PORT = 62151
const SERVICE_HOST = "127.0.0.1"

func main() {

	l := logger.GetSlogLogger()

	// TLS certificate file paths
	// Uncomment & adjust for local certs path
	// caFilePath := "certs/local-certs/ca.pem"
	// certFilePath := "certs/local-certs/client.pem"
	// keyFilePath := "certs/local-certs/client-key.pem"

	tlsConfig, err := config.SetupTLSConfig(&config.ConfigOpts{
		Target: config.CLIENT,
		// Uncomment & adjust for local certs path
		// Opts: &config.CustomOpts{
		// 	CAFilePath:   caFilePath,
		// 	CertFilePath: certFilePath,
		// 	KeyFilePath:  keyFilePath,
		// },
	})
	if err != nil {
		l.Error("error setting client TLS", "error", err.Error())
		return
	}
	tlsConfig.InsecureSkipVerify = true
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}

	serviceAddr := fmt.Sprintf("%s:%d", SERVICE_HOST, SERVICE_PORT)
	conn, err := grpc.NewClient(serviceAddr, opts...)
	if err != nil {
		l.Error("client failed to connect", "error", err.Error())
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			l.Error("error closing connection", "error", err.Error())
		}
	}()

	client := api.NewStoresClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	storeCRUD(ctx, client)

	l.Info("stores client testing done")
}

func storeCRUD(ctx context.Context, client api.StoresClient) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	asResp, err := client.AddStore(ctx, &api.AddStoreRequest{
		Org:       "Test Org",
		Name:      "Test Store",
		AddressId: "dacdbddabcadccbdacac",
	})
	if err != nil {
		l.Error("error adding store", "error", err.Error())
		return
	}

	store, err := client.GetStore(ctx, &api.GetStoreRequest{
		Id: asResp.GetId(),
	})
	if err != nil {
		l.Error("error getting store", "error", err.Error())
		return
	}

	usResp, err := client.UpdateStore(ctx, &api.UpdateStoreRequest{
		Id:   asResp.GetId(),
		Name: "Updated Test Store",
	})
	if err != nil {
		l.Error("error updating store", "error", err.Error())
		return
	}
	if !usResp.Ok {
		l.Error("update store response not ok")
		return
	}

	store, err = client.GetStore(ctx, &api.GetStoreRequest{
		Id: asResp.GetId(),
	})
	if err != nil {
		l.Error("error getting store", "error", err.Error())
		return
	}
	if store.GetStore().Name != "Updated Test Store" {
		l.Error("store name mismatch", "expected", "Updated Test Store", "actual", store.GetStore().Name)
		return
	}

	delResp, err := client.DeleteStore(ctx, &api.DeleteStoreRequest{
		Id: asResp.GetId(),
	})
	if err != nil {
		l.Error("error deleting store", "error", err.Error())
		return
	}
	if !delResp.Ok {
		l.Error("delete store response not ok")
		return
	}

	_, err = client.GetStore(ctx, &api.GetStoreRequest{
		Id: asResp.GetId(),
	})
	if err == nil {
		l.Error("expected error when getting deleted store, but got none")
		return
	}
}
