package server

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/comfforts/logger"

	api "github.com/comfforts/comff-stores/api/v1"
	"github.com/comfforts/comff-stores/internal/auth"
	"github.com/comfforts/comff-stores/internal/config"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/examples/exporter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
	testUtils "github.com/comfforts/comff-stores/pkg/utils/test"
)

const TEST_DIR = "test-data"

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client api.StoresClient,
		nbClient api.StoresClient,
		config *Config,
	){
		"add and get store succeeds": testAddAndGetStore,
		"test unauthorized client":   testUnauthorizedAddStore,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, nbClient, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, nbClient, config)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config)) (
	client api.StoresClient,
	nbClient api.StoresClient,
	cfg *Config,
	teardown func(),
) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:50055")
	require.NoError(t, err)

	caFilePath := filepath.Join("../../cmd/store", config.CertFile(config.CAFile))

	// grpc client
	newClient := func(crtPath, keyPath string) (*grpc.ClientConn, api.StoresClient, []grpc.DialOption) {
		certFilePath := filepath.Join("../../cmd/store", config.CertFile(crtPath))
		keyFilePath := filepath.Join("../../cmd/store", config.CertFile(keyPath))

		tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
			CertFile: certFilePath,
			KeyFile:  keyFilePath,
			CAFile:   caFilePath,
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
	nbcc, nbClient, _ := newClient(config.NobodyClientCertFile, config.NobodyClientKeyFile)

	appLogger := logger.NewTestAppLogger(TEST_DIR)
	css, err := store.NewStoreService(appLogger)
	require.NoError(t, err)

	modelFilePath := filepath.Join("../../cmd/store", config.PolicyFile(config.ACLModelFile))
	policyFilePath := filepath.Join("../../cmd/store", config.PolicyFile(config.ACLPolicyFile))

	authorizer, err := auth.NewAuthorizer(modelFilePath, policyFilePath, appLogger)
	require.NoError(t, err)

	mfPath := filepath.Join(TEST_DIR, "logs", "metrics.log")
	err = fileUtils.CreateDirectory(mfPath)
	require.NoError(t, err)

	metricsLogFile, err := os.Create(mfPath)
	require.NoError(t, err)

	tfPath := filepath.Join(TEST_DIR, "logs", "traces.log")
	tracesLogFile, err := os.Create(tfPath)
	require.NoError(t, err)

	telemetryExporter, err := exporter.NewLogExporter(exporter.Options{
		MetricsLogFile:    metricsLogFile.Name(),
		TracesLogFile:     tracesLogFile.Name(),
		ReportingInterval: time.Second,
	})
	require.NoError(t, err)

	err = telemetryExporter.Start()
	require.NoError(t, err)

	cfg = &Config{
		StoreService: css,
		Authorizer:   authorizer,
		Logger:       appLogger,
	}
	if fn != nil {
		fn(cfg)
	}

	// TLS config
	serCertFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ServerCertFile))
	serKeyFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ServerKeyFile))
	srvTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      serCertFilePath,
		KeyFile:       serKeyFilePath,
		CAFile:        caFilePath,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})
	require.NoError(t, err)
	srvCreds := credentials.NewTLS(srvTLSConfig)

	// grpc server
	server, err := NewGRPCServer(cfg, grpc.Creds(srvCreds))
	require.NoError(t, err)

	go func() {
		err := server.Serve(l)
		require.NoError(t, err)
	}()

	client = api.NewStoresClient(cc)

	return client, nbClient, cfg, func() {
		server.Stop()
		cc.Close()
		nbcc.Close()
		l.Close()
		css.Close()

		if telemetryExporter != nil {
			time.Sleep(1500 * time.Millisecond)
			telemetryExporter.Stop()
			telemetryExporter.Close()
		}

		err := os.RemoveAll(TEST_DIR)
		require.NoError(t, err)
	}
}

func testAddAndGetStore(t *testing.T, client, nbClient api.StoresClient, config *Config) {
	t.Helper()
	storeId, name, org, city, country := uint64(1), "Plaza Hollywood", "starbucks", "Hong Kong", "CN"
	lat, long := 22.3228702545166, 114.21343994140625
	addStoreReq := testUtils.CreateAddStoreRequest(storeId, name, org, city, country, lat, long)

	ctx := context.Background()
	addStoreRes, err := client.AddStore(ctx, addStoreReq)
	require.NoError(t, err)
	require.Equal(t, addStoreRes.Ok, true)

	getStoreReq := &api.GetStoreRequest{Id: addStoreRes.Store.Id}
	ctx = context.Background()
	getStoreRes, err := client.GetStore(ctx, getStoreReq)
	require.NoError(t, err)
	require.Equal(t, getStoreRes.Store.Id, addStoreRes.Store.Id)
}

func testUnauthorizedAddStore(t *testing.T, client, nbClient api.StoresClient, config *Config) {
	t.Helper()
	storeId, name, org, city, country := uint64(1), "Plaza Hollywood", "starbucks", "Hong Kong", "CN"
	lat, long := 22.3228702545166, 114.21343994140625
	addStoreReq := testUtils.CreateAddStoreRequest(storeId, name, org, city, country, lat, long)

	ctx := context.Background()
	addStoreRes, err := nbClient.AddStore(ctx, addStoreReq)
	require.Error(t, err)
	assert.Equal(t, addStoreRes, (*api.AddStoreResponse)(nil), "adding new store by an unauthorized client should fail")
}
