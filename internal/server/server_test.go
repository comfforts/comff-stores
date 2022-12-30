package server

import (
	"context"
	"net"
	"testing"

	api "github.com/comfforts/comff-stores/api/v1"
	"github.com/comfforts/comff-stores/internal/config"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client api.StoresClient,
		config *Config,
	){
		"add and get store by id succeeds": testAddAndGetStore,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, config)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config)) (
	client api.StoresClient,
	cfg *Config,
	teardown func(),
) {
	t.Helper()

	l, err := net.Listen("tcp", "192.168.68.100:50055")
	require.NoError(t, err)

	// grpc client
	newClient := func(crtPath, keyPath string) (*grpc.ClientConn, api.StoresClient, []grpc.DialOption) {
		tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
			CertFile: crtPath,
			KeyFile:  keyPath,
			CAFile:   config.CAFile,
			Server:   false,
		})
		require.NoError(t, err)

		tlsCreds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
		addr := l.Addr().String()
		conn, err := grpc.Dial(addr, opts...)
		require.NoError(t, err)
		client = api.NewStoresClient(conn)
		return conn, client, opts
	}

	cc, client, _ := newClient(config.ClientCertFile, config.ClientKeyFile)

	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)
	css := store.NewStoreService(appLogger)

	cfg = &Config{
		StoreService: css,
	}
	if fn != nil {
		fn(cfg)
	}

	// grpc server
	srvTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})
	require.NoError(t, err)
	srvCreds := credentials.NewTLS(srvTLSConfig)

	server, err := NewGRPCServer(cfg, grpc.Creds(srvCreds))
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	client = api.NewStoresClient(cc)

	return client, cfg, func() {
		server.Stop()
		cc.Close()
		l.Close()
		css.Clear()
	}
}

func defaulAddStoreRequest(storeId uint64, name, org, city string) *api.AddStoreRequest {
	s := &api.AddStoreRequest{
		Name:      name,
		Org:       org,
		City:      city,
		Country:   "CN",
		Longitude: 114.20169067382812,
		Latitude:  22.340700149536133,
		StoreId:   storeId,
	}
	return s
}

func testAddAndGetStore(t *testing.T, client api.StoresClient, config *Config) {
	t.Helper()
	storeId, name, org, city := uint64(1), "Plaza Hollywood", "starbucks", "Hong Kong"
	addStoreReq := defaulAddStoreRequest(storeId, name, org, city)

	ctx := context.Background()
	addStoreRes, err := client.AddStore(ctx, addStoreReq)
	require.NoError(t, err)
	require.Equal(t, addStoreRes.Ok, true)

	id, err := store.BuildId(float64(addStoreReq.Latitude), float64(addStoreReq.Longitude), org)
	require.NoError(t, err)
	require.Equal(t, addStoreRes.Store.Id, id)

	getStoreReq := &api.GetStoreRequest{Id: id}
	ctx = context.Background()
	getStoreRes, err := client.GetStore(ctx, getStoreReq)
	require.NoError(t, err)
	require.Equal(t, getStoreRes.Store.Id, id)
}
