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

func TestAgentSingle(t *testing.T) {
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
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      serCertFilePath,
		KeyFile:       serKeyFilePath,
		CAFile:        caFilePath,
		Server:        true,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.RemoveAll(TEST_DIR))
	}()

	servPort := getPort(0)
	bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", servPort)
	rpcPort := getPort(1)

	var peerNodeAddrs []string
	agent, err := NewAgent(Config{
		NodeName:        fmt.Sprintf("%d", 0),
		Bootstrap:       true,
		DataDir:         filepath.Join(TEST_DIR, fmt.Sprintf("node-%d", 0)),
		PeerNodeAddrs:   peerNodeAddrs,
		BindAddr:        bindAddr,
		RPCPort:         rpcPort,
		ACLModelFile:    modelFilePath,
		ACLPolicyFile:   policyFilePath,
		ServerTLSConfig: serverTLSConfig,
		PeerTLSConfig:   peerTLSConfig,
		AppConfigFile:   "../../cmd/store/test-config.json",
	})
	require.NoError(t, err)

	defer func() {
		err := agent.Shutdown()
		require.NoError(t, err)
	}()

	// skip the cycle and wait for bit
	time.Sleep(1 * time.Second)

	// get leader node
	leaderClient := client(t, agent, peerTLSConfig)

	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()
	_, err = leaderClient.GetStats(ctx, &api.StatsRequest{})
	require.NoError(t, err)

	storeId, name, org, city, country := uint64(1), "Plaza Hollywood", "starbucks", "Hong Kong", "CN"
	lat, long := 22.3228702545166, 114.21343994140625
	asReq := testUtils.CreateAddStoreRequest(storeId, name, org, city, country, lat, long)

	// add store through leader node should succeed
	ctx, cancel1 := context.WithCancel(baseCtx)
	defer cancel1()
	asrResp, err := leaderClient.AddStore(ctx, asReq)
	require.NoError(t, err)

	// get store through leader node should succeed
	ctx, cancel2 := context.WithCancel(baseCtx)
	defer cancel2()
	gsReq := &api.GetStoreRequest{Id: asrResp.Store.Id}
	got, err := leaderClient.GetStore(ctx, gsReq)
	require.NoError(t, err)
	require.Equal(t, asrResp.Store.Id, got.Store.Id)
}

func getPort(id int) int {
	return 15001 + id
}

func TestAgent(t *testing.T) {
	caFilePath := filepath.Join("../../cmd/store", config.CertFile(config.CAFile))
	serCertFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ServerCertFile))
	serKeyFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ServerKeyFile))
	clCertFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ClientCertFile))
	clKeyFilePath := filepath.Join("../../cmd/store", config.CertFile(config.ClientKeyFile))
	modelFilePath := filepath.Join("../../cmd/store", config.PolicyFile(config.ACLModelFile))
	policyFilePath := filepath.Join("../../cmd/store", config.PolicyFile(config.ACLPolicyFile))

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      serCertFilePath,
		KeyFile:       serKeyFilePath,
		CAFile:        caFilePath,
		Server:        true,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	peerTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      clCertFilePath,
		KeyFile:       clKeyFilePath,
		CAFile:        caFilePath,
		Server:        false,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	var agents []*Agent
	for i := 0; i < 3; i++ {
		ports := dynaport.Get(2)
		bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
		rpcPort := ports[1]

		var peerNodeAddrs []string
		if i != 0 {
			peerNodeAddrs = append(
				peerNodeAddrs,
				agents[0].Config.BindAddr,
			)
		}

		agent, err := NewAgent(Config{
			NodeName:        fmt.Sprintf("%d", i),
			Bootstrap:       i == 0,
			DataDir:         filepath.Join(TEST_DIR, fmt.Sprintf("node-%d", i)),
			PeerNodeAddrs:   peerNodeAddrs,
			BindAddr:        bindAddr,
			RPCPort:         rpcPort,
			ACLModelFile:    modelFilePath,
			ACLPolicyFile:   policyFilePath,
			ServerTLSConfig: serverTLSConfig,
			PeerTLSConfig:   peerTLSConfig,
			AppConfigFile:   "../../cmd/store/test-config.json",
		})
		require.NoError(t, err)
		agents = append(agents, agent)
	}

	defer func() {
		for _, agent := range agents {
			err := agent.Shutdown()
			require.NoError(t, err)
			// require.NoError(t, os.RemoveAll(TEST_CACHE_DIR))
			require.NoError(t, os.RemoveAll(TEST_DIR))
		}
	}()

	// skip the cycle and wait for bit
	time.Sleep(50 * time.Millisecond)

	// get leader node
	leaderClient := client(t, agents[0], peerTLSConfig)

	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()
	_, err = leaderClient.GetStats(ctx, &api.StatsRequest{})
	require.NoError(t, err)

	storeId, name, org, city, country := uint64(1), "Plaza Hollywood", "starbucks", "Hong Kong", "CN"
	lat, long := 22.3228702545166, 114.21343994140625
	asReq := testUtils.CreateAddStoreRequest(storeId, name, org, city, country, lat, long)

	// add store through leader node should succeed
	ctx, cancel1 := context.WithCancel(baseCtx)
	defer cancel1()
	asrResp, err := leaderClient.AddStore(ctx, asReq)
	require.NoError(t, err)

	// get store through leader node should succeed
	ctx, cancel2 := context.WithCancel(baseCtx)
	defer cancel2()
	gsReq := &api.GetStoreRequest{Id: asrResp.Store.Id}
	got, err := leaderClient.GetStore(ctx, gsReq)
	require.NoError(t, err)
	require.Equal(t, asrResp.Store.Id, got.Store.Id)

	// skip the cycle and wait for bit
	time.Sleep(1 * time.Second)

	// get follower node
	followerClient := client(t, agents[1], peerTLSConfig)

	ctx, cancel3 := context.WithCancel(baseCtx)
	defer cancel3()
	_, err = followerClient.GetStats(ctx, &api.StatsRequest{})
	require.NoError(t, err)

	// get store through follower node should succeed
	ctx, cancel4 := context.WithCancel(baseCtx)
	defer cancel4()
	gsReq = &api.GetStoreRequest{Id: asrResp.Store.Id}
	got, err = followerClient.GetStore(ctx, gsReq)
	require.NoError(t, err)
	require.Equal(t, asrResp.Store.Id, got.Store.Id)
}

func client(
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
