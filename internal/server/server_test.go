package server

import (
	"context"
	"net"
	"testing"

	api "github.com/comfforts/comff-stores/api/v1"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client api.StoresClient,
		config *Config,
	){
		"add and get store by id suceeds": testAddAndGetStore,
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

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	css := store.New(logger)

	cfg = &Config{
		StoreService: css,
	}
	if fn != nil {
		fn(cfg)
	}
	server, err := NewGrpcServer(cfg)
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

func defaulAddStoreRequest(id uint32, name, city string) *api.AddStoreRequest {
	s := &api.AddStoreRequest{
		City:      city,
		Name:      name,
		Country:   "CN",
		Longitude: 114.20169067382812,
		Latitude:  22.340700149536133,
		StoreId:   id,
	}
	return s
}

func testAddAndGetStore(t *testing.T, client api.StoresClient, config *Config) {
	t.Helper()
	id, name, city := 1, "Plaza Hollywood", "Hong Kong"
	addStoreReq := defaulAddStoreRequest(uint32(id), name, city)

	ctx := context.Background()
	addStoreRes, err := client.AddStore(ctx, addStoreReq)
	require.NoError(t, err)
	assert.Equal(t, addStoreRes.Ok, true, "adding new store should be success")
}
