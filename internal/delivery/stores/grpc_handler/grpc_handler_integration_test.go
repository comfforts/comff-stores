package grpchandler_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	config "github.com/comfforts/comff-config"
	geo_v1 "github.com/comfforts/comff-geo/api/geo/v1"
	geocl "github.com/comfforts/comff-geo/clients/go"
	"github.com/comfforts/logger"

	api "github.com/comfforts/comff-stores/api/stores/v1"
	grpchandler "github.com/comfforts/comff-stores/internal/delivery/stores/grpc_handler"
	"github.com/comfforts/comff-stores/internal/infra/mongostore"
	"github.com/comfforts/comff-stores/internal/infra/observability"
	strepo "github.com/comfforts/comff-stores/internal/repo/stores"
	"github.com/comfforts/comff-stores/internal/usecase/services/stores"
	envutils "github.com/comfforts/comff-stores/pkg/utils/environ"
	testutils "github.com/comfforts/comff-stores/pkg/utils/test"
)

func TestGRPCHandler(t *testing.T) {
	l := logger.GetSlogLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	_, _, teardown := setupTest(t)
	defer teardown()
}

func TestGRPCHandler_Stores_Unauthenticated_Client(t *testing.T) {
	l := logger.GetSlogLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	_, nbClient, teardown := setupTest(t)
	defer teardown()

	_, err := nbClient.AddStore(ctx, &api.AddStoreRequest{
		Org:       "Test Org",
		Name:      "Test Store",
		AddressId: "dacdbddabcadccbdacac",
	})
	require.Error(t, err)

	stErr, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, stErr.Code())
	assert.Equal(t, grpchandler.ERR_UNAUTHORIZED_ADD_STORE, stErr.Message())
}

func TestGRPCHandler_Stores_CRUD(t *testing.T) {
	l := logger.GetSlogLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	client, _, teardown := setupTest(t)
	defer teardown()

	asResp, err := client.AddStore(ctx, &api.AddStoreRequest{
		Org:       "Test Org",
		Name:      "Test Store",
		AddressId: "dacdbddabcadccbdacac",
	})
	require.NoError(t, err)

	store, err := client.GetStore(ctx, &api.GetStoreRequest{
		Id: asResp.GetId(),
	})
	require.NoError(t, err)
	assert.Equal(t, "Test Store", store.GetStore().Name)

	usResp, err := client.UpdateStore(ctx, &api.UpdateStoreRequest{
		Id:   asResp.GetId(),
		Name: "Updated Test Store",
	})
	require.NoError(t, err)
	require.True(t, usResp.Ok)

	store, err = client.GetStore(ctx, &api.GetStoreRequest{
		Id: asResp.GetId(),
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Store", store.GetStore().Name)

	delResp, err := client.DeleteStore(ctx, &api.DeleteStoreRequest{
		Id: asResp.GetId(),
	})
	require.NoError(t, err)
	require.True(t, delResp.Ok)

	_, err = client.GetStore(ctx, &api.GetStoreRequest{
		Id: asResp.GetId(),
	})
	require.Error(t, err)
}

func TestGRPCHandler_Stores_Search(t *testing.T) {
	l := logger.GetSlogLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	client, _, teardown := setupTest(t)
	defer teardown()

	// Initialize geo client options
	clientOpts := geocl.NewDefaultClientOption()
	clientOpts.Caller = "stores-server-geo-client-test"

	// Initialize geo client
	gc, err := geocl.NewClient(ctx, clientOpts)
	require.NoError(t, err)
	defer func() {
		err := gc.Close(ctx)
		require.NoError(t, err)
	}()

	addrIdMap := map[string]*geo_v1.Point{}
	addrIds := []string{}

	dests := testutils.BuildPetalumaSet1()
	for i, dest := range dests {
		l.Debug("Testing GeoLocate for destination", "index", i, "destination", dest)
		resp, err := gc.GeoLocate(ctx, dest)
		require.NoError(t, err)
		require.NotNil(t, resp)
		addrIdMap[resp.Point.Hash] = resp.Point
		addrIds = append(addrIds, resp.Point.Hash)
	}

	dests = testutils.BuildPetalumaSet2()
	for i, dest := range dests {
		l.Debug("Testing GeoLocate for destination", "index", i, "destination", dest)
		resp, err := gc.GeoLocate(ctx, dest)
		require.NoError(t, err)
		require.NotNil(t, resp)
		addrIdMap[resp.Point.Hash] = resp.Point
		addrIds = append(addrIds, resp.Point.Hash)
	}

	stIds := []string{}
	for i, addrId := range addrIds {
		asResp, err := client.AddStore(ctx, &api.AddStoreRequest{
			Org:       fmt.Sprintf("Test Org %d", i%2), // Alternate between 2 orgs
			Name:      fmt.Sprintf("Test Store %d", i),
			AddressId: addrId,
		})
		require.NoError(t, err)
		require.NotEmpty(t, asResp.GetId())
		l.Debug("Added store for address", "index", i, "addressId", addrId, "storeId", asResp.GetId())
		stIds = append(stIds, asResp.GetId())
	}
	defer func() {
		for i, stId := range stIds {
			dResp, err := client.DeleteStore(ctx, &api.DeleteStoreRequest{
				Id: stId,
			})
			require.NoError(t, err)
			require.True(t, dResp.GetOk())
			l.Debug("Deleted store", "index", i, "storeId", stId)
		}
	}()

	// org search
	ssResp, err := client.SearchStore(ctx, &api.SearchStoreRequest{
		Org: "Test Org 0",
	})
	require.NoError(t, err)
	require.NotNil(t, ssResp)
	require.GreaterOrEqual(t, len(ssResp.GetStores()), 1)
	l.Debug("SearchStores returned stores", "count", len(ssResp.GetStores()))

	// name search
	ssResp, err = client.SearchStore(ctx, &api.SearchStoreRequest{
		Name: "Test Store 0",
	})
	require.NoError(t, err)
	require.NotNil(t, ssResp)
	require.GreaterOrEqual(t, len(ssResp.GetStores()), 1)
	l.Debug("SearchStores returned stores", "count", len(ssResp.GetStores()))

	// address id search
	ssResp, err = client.SearchStore(ctx, &api.SearchStoreRequest{
		AddressId: addrIds[0],
	})
	require.NoError(t, err)
	require.NotNil(t, ssResp)
	require.Equal(t, 1, len(ssResp.GetStores()))
	l.Debug("SearchStores returned stores", "count", len(ssResp.GetStores()))

	// lat/long search
	ssResp, err = client.SearchStore(ctx, &api.SearchStoreRequest{
		Latitude:  addrIdMap[addrIds[1]].Latitude,
		Longitude: addrIdMap[addrIds[1]].Longitude,
		Distance:  10000, // meters
	})
	require.NoError(t, err)
	require.NotNil(t, ssResp)
	require.GreaterOrEqual(t, len(ssResp.GetStores()), 1)
	l.Debug("SearchStores returned stores", "count", len(ssResp.GetStores()))

	// address string search
	ssResp, err = client.SearchStore(ctx, &api.SearchStoreRequest{
		AddressStr: addrIdMap[addrIds[1]].FormattedAddress,
		Distance:   50000, // meters
	})
	require.NoError(t, err)
	require.NotNil(t, ssResp)
	require.GreaterOrEqual(t, len(ssResp.GetStores()), 1)
	l.Debug("SearchStores returned stores", "count", len(ssResp.GetStores()))
}

func setupTest(t *testing.T) (client, nbClient api.StoresClient, teardown func()) {
	t.Helper()

	l := logger.GetSlogLogger()

	lis, err := net.Listen("tcp", "127.0.0.1:62151")
	require.NoError(t, err)

	// grpc clients
	cc, client, _, err := newClient(config.CLIENT, lis)
	require.NoError(t, err)
	nbcc, nbClient, _, err := newClient(config.NOBODY_CLIENT, lis)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	server, closeFn, err := newServer(ctx, lis.Addr().String())
	require.NoError(t, err)

	go func() {
		err := server.Serve(lis)
		require.NoError(t, err)
	}()

	return client, nbClient, func() {
		err := cc.Close()
		require.NoError(t, err)

		err = nbcc.Close()
		require.NoError(t, err)

		err = closeFn()
		require.NoError(t, err)

		server.GracefulStop()

		// err = os.RemoveAll(TEST_DIR)
		// require.NoError(t, err)
	}
}

func newClient(target config.ConfigurationTarget, lis net.Listener) (*grpc.ClientConn, api.StoresClient, []grpc.DialOption, error) {
	clTLSCfg := envutils.BuildClientTLSConfig()
	if target == config.NOBODY_CLIENT {
		clTLSCfg = envutils.BuildNobodyClientTLSConfig()
	}

	// Client TLS config
	tlsConfig, err := config.SetupTLSConfig(&config.ConfigOpts{
		Target: target,
		Opts: &config.CustomOpts{
			CAFilePath:   clTLSCfg.CAFilePath,
			CertFilePath: clTLSCfg.CertFilePath,
			KeyFilePath:  clTLSCfg.KeyFilePath,
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}

	addr := lis.Addr().String()
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, nil, nil, err
	}
	client := api.NewStoresClient(conn)
	return conn, client, opts, nil
}

func newServer(ctx context.Context, addr string) (*grpc.Server, func() error, error) {
	l := logger.GetSlogLogger()
	l.Debug("TestGRPCHandler started")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	metrics, err := observability.NewMetrics()
	if err != nil {
		return nil, nil, err
	}

	// Initialize MongoDB store
	nmCfg := envutils.BuildMongoStoreConfig(true)
	ms, err := mongostore.NewMongoStore(ctx, nmCfg)
	if err != nil {
		return nil, nil, err
	}

	// Initialize stores repository
	sr, err := strepo.NewStoresRepo(ctx, ms, metrics)
	if err != nil {
		return nil, nil, err
	}

	// Initialize geo client options
	clientOpts := geocl.NewDefaultClientOption()
	clientOpts.Caller = "geo-service-geo-client-test"

	// Initialize geo client
	gc, err := geocl.NewClient(ctx, clientOpts)
	if err != nil {
		return nil, nil, err
	}

	closeFn := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		ctx = logger.WithLogger(ctx, logger.GetSlogLogger())

		err := sr.Close(ctx)
		if err != nil {
			logger.GetSlogLogger().Error("error closing stores repository", "error", err.Error())
		}

		if rErr := gc.Close(ctx); rErr != nil {
			logger.GetSlogLogger().Error("error closing geo client", "error", rErr.Error())
			err = errors.Join(err, rErr)
		}
		return err
	}

	// Initialize stores service
	ss, err := stores.NewStoresService(ctx, sr, gc, metrics)
	if err != nil {
		return nil, closeFn, err
	}

	// Build gRPC server config
	cfg, err := grpchandler.BuildServerConfig(ctx, ss)
	if err != nil {
		return nil, closeFn, err
	}

	srvTLSCfg := envutils.BuildServerTLSConfig()

	// Server TLS config
	srvTLSConfig, err := config.SetupTLSConfig(&config.ConfigOpts{
		Target: config.SERVER,
		Addr:   addr,
		Opts: &config.CustomOpts{
			CAFilePath:   srvTLSCfg.CAFilePath,
			CertFilePath: srvTLSCfg.CertFilePath,
			KeyFilePath:  srvTLSCfg.KeyFilePath,
		},
	})
	if err != nil {
		return nil, closeFn, err
	}
	srvCreds := credentials.NewTLS(srvTLSConfig)

	// grpc server
	server, err := grpchandler.NewGRPCServer(cfg, grpc.Creds(srvCreds))
	if err != nil {
		return nil, closeFn, err
	}
	return server, closeFn, nil
}
