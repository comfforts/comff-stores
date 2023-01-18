package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/internal/config"
	testUtils "github.com/comfforts/comff-stores/pkg/utils/test"
)

const TEST_DIR = "test-data"

type securityConfig struct {
	servTLS *tls.Config
	peerTLS *tls.Config
	modPath string
	polPath string
}

type nodeConfig struct {
	name      string
	dataDir   string
	bootstrap bool
	peerAddrs []string
	cfgPath   string
	servAddr  string
	bindPort  int
	rpcPort   int
}

func TestSingleAgentRestart(t *testing.T) {
	servAddr := "127.0.0.1"

	securityCfg := setupSecurityConfig(t, servAddr)

	nodeCfg := setupNodeConfig(t, 0, servAddr, true)

	ports := dynaport.Get(2)
	nodeCfg.bindPort = ports[0]
	nodeCfg.rpcPort = ports[1]

	defer func() {
		require.NoError(t, os.RemoveAll(TEST_DIR))
	}()

	agent := setupAgent(t, securityCfg, nodeCfg)

	defer func() {
		err := agent.Shutdown()
		require.NoError(t, err)
	}()

	// skip the cycle and wait for bit
	time.Sleep(50 * time.Millisecond)

	// get leader node
	leaderClient := setupClient(t, agent, securityCfg.peerTLS)
	testGetServers(t, leaderClient)
	testGetStats(t, leaderClient)

	resps := testAddStores(t, leaderClient)
	ids := []string{}
	for _, v := range resps {
		ids = append(ids, v.Store.Id)
	}
	testGetStores(t, leaderClient, ids)

	// skip the cycle and wait for bit
	time.Sleep(20 * time.Microsecond)

	err := agent.Shutdown()
	require.NoError(t, err)

	ports = dynaport.Get(2)
	nodeCfg.bindPort = ports[0]
	nodeCfg.rpcPort = ports[1]

	agent = setupAgent(t, securityCfg, nodeCfg)

	// skip the cycle and wait for bit
	time.Sleep(50 * time.Millisecond)

	testGetStores(t, leaderClient, ids)
}

func TestSingleSetup(t *testing.T) {
	servAddr := "127.0.0.1"

	securityCfg := setupSecurityConfig(t, servAddr)
	nodeCfg := setupNodeConfig(t, 0, servAddr, true)

	ports := dynaport.Get(2)
	nodeCfg.bindPort = ports[0]
	nodeCfg.rpcPort = ports[1]

	defer func() {
		require.NoError(t, os.RemoveAll(TEST_DIR))
	}()

	agent := setupAgent(t, securityCfg, nodeCfg)

	defer func() {
		err := agent.Shutdown()
		require.NoError(t, err)
	}()

	// skip the cycle and wait for bit
	time.Sleep(1 * time.Second)

	// get leader node
	leaderClient := setupClient(t, agent, securityCfg.peerTLS)
	testGetServers(t, leaderClient)
	testGetStats(t, leaderClient)

	store := testUtils.CreateStoreModel()
	asReq := testUtils.CreateAddStoreRequest(store.StoreId, store.Name, store.Org, store.City, store.Country, store.Latitude, store.Longitude)
	addResp := testAddStore(t, leaderClient, asReq)
	testGetStore(t, leaderClient, addResp.Store.Id)
}

func TestMultiAgentSetup(t *testing.T) {
	servAddr := "127.0.0.1"

	securityCfg := setupSecurityConfig(t, servAddr)

	var agents []*Agent
	for i := 0; i < 3; i++ {
		nodeCfg := setupNodeConfig(t, i, servAddr, i == 0)
		if i != 0 {
			nodeCfg.peerAddrs = append(nodeCfg.peerAddrs, agents[0].Config.BindAddr)
		}
		ports := dynaport.Get(2)
		nodeCfg.bindPort = ports[0]
		nodeCfg.rpcPort = ports[1]
		agents = append(agents, setupAgent(t, securityCfg, nodeCfg))
	}

	defer func() {
		for _, agent := range agents {
			err := agent.Shutdown()
			require.NoError(t, err)
			require.NoError(t, os.RemoveAll(TEST_DIR))
		}
	}()

	// skip the cycle and wait for bit
	time.Sleep(50 * time.Millisecond)

	// connect client with leader node
	leaderClient := setupClient(t, agents[0], securityCfg.peerTLS)

	testGetServers(t, leaderClient)
	testGetStats(t, leaderClient)

	store := testUtils.CreateStoreModel()
	asReq := testUtils.CreateAddStoreRequest(store.StoreId, store.Name, store.Org, store.City, store.Country, store.Latitude, store.Longitude)
	addResp := testAddStore(t, leaderClient, asReq)
	testGetStore(t, leaderClient, addResp.Store.Id)

	// skip the cycle and wait for bit
	time.Sleep(1 * time.Second)

	// connect client with follower node
	followerClient := setupClient(t, agents[1], securityCfg.peerTLS)
	testGetServers(t, followerClient)
	testGetStats(t, followerClient)
	testGetStore(t, followerClient, addResp.Store.Id)
}

func setupSecurityConfig(t *testing.T, servAddr string) securityConfig {
	t.Helper()

	caFilePath := filepath.Join("../../cmd/store", config.CertFile(config.CAFile))
	serCertFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ServerCertFile))
	serKeyFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ServerKeyFile))
	clCertFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ClientCertFile))
	clKeyFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ClientKeyFile))
	modelFilePath := filepath.Join("../../cmd/store", config.PolicyFile(config.ACLModelFile))
	policyFilePath := filepath.Join("../../cmd/store", config.PolicyFile(config.ACLPolicyFile))

	peerTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      clCertFilePath,
		KeyFile:       clKeyFilePath,
		CAFile:        caFilePath,
		Server:        false,
		ServerAddress: servAddr,
	})
	require.NoError(t, err)

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      serCertFilePath,
		KeyFile:       serKeyFilePath,
		CAFile:        caFilePath,
		Server:        true,
		ServerAddress: servAddr,
	})
	require.NoError(t, err)
	return securityConfig{
		serverTLSConfig,
		peerTLSConfig,
		modelFilePath,
		policyFilePath,
	}
}

func setupNodeConfig(t *testing.T, id int, servAddr string, bootstrap bool) nodeConfig {
	return nodeConfig{
		name:      fmt.Sprintf("%d", id),
		bootstrap: bootstrap,
		dataDir:   filepath.Join(TEST_DIR, fmt.Sprintf("node-%d", id)),
		cfgPath:   "../../cmd/store/test-config.json",
		peerAddrs: []string{},
		servAddr:  servAddr,
	}
}

func setupAgent(t *testing.T, securityCfg securityConfig, nodeCfg nodeConfig) *Agent {
	bindAddr := fmt.Sprintf("%s:%d", nodeCfg.servAddr, nodeCfg.bindPort)

	agent, err := NewAgent(Config{
		NodeName:        nodeCfg.name,
		Bootstrap:       nodeCfg.bootstrap,
		DataDir:         nodeCfg.dataDir,
		PeerNodeAddrs:   nodeCfg.peerAddrs,
		BindAddr:        bindAddr,
		RPCPort:         nodeCfg.rpcPort,
		ACLModelFile:    securityCfg.modPath,
		ACLPolicyFile:   securityCfg.polPath,
		ServerTLSConfig: securityCfg.servTLS,
		PeerTLSConfig:   securityCfg.peerTLS,
		AppConfigFile:   nodeCfg.cfgPath,
		MaxIndexSize:    3,
	})
	require.NoError(t, err)
	return agent
}

func setupClient(
	t *testing.T,
	agent *Agent,
	tlsConfig *tls.Config,
) api.StoresClient {
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
	rpcAddr, err := agent.Config.RPCAddr()
	require.NoError(t, err)
	conn, err := grpc.Dial(rpcAddr, opts...)
	require.NoError(t, err)
	client := api.NewStoresClient(conn)
	return client
}

func testGetServers(t *testing.T, client api.StoresClient) {
	t.Helper()

	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()
	resp, err := client.GetServers(ctx, &api.GetServersRequest{})
	t.Logf("GetServers response: %v", resp)
	require.NoError(t, err)
}

func testGetStats(t *testing.T, client api.StoresClient) {
	t.Helper()

	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()
	resp, err := client.GetStats(ctx, &api.StatsRequest{})
	t.Logf("GetStats response: %v", resp)
	require.NoError(t, err)
}

func testAddStore(t *testing.T, client api.StoresClient, asReq *api.AddStoreRequest) *api.AddStoreResponse {
	t.Helper()

	baseCtx := context.Background()

	// add store through leader node should succeed
	ctx, cancel1 := context.WithCancel(baseCtx)
	defer cancel1()
	resp, err := client.AddStore(ctx, asReq)
	require.NoError(t, err)
	return resp
}

func testAddStores(t *testing.T, client api.StoresClient) []*api.AddStoreResponse {
	t.Helper()

	stores := testUtils.CreateStoreModelList()
	resps := []*api.AddStoreResponse{}
	for _, v := range stores {
		asReq := testUtils.CreateAddStoreRequest(v.StoreId, v.Name, v.Org, v.City, v.Country, v.Latitude, v.Longitude)
		addResp := testAddStore(t, client, asReq)
		resps = append(resps, addResp)
	}

	require.Equal(t, len(stores), len(resps))
	return resps
}

func testGetStores(t *testing.T, client api.StoresClient, ids []string) {
	t.Helper()

	for _, id := range ids {
		testGetStore(t, client, id)
	}
}

func testGetStore(t *testing.T, client api.StoresClient, id string) {
	t.Helper()

	baseCtx := context.Background()
	ctx, cancel2 := context.WithCancel(baseCtx)
	defer cancel2()
	gsReq := &api.GetStoreRequest{Id: id}
	got, err := client.GetStore(ctx, gsReq)
	require.NoError(t, err)
	require.Equal(t, id, got.Store.Id)
}
