package store

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"

	"github.com/comfforts/comff-stores/pkg/logging"
	testUtils "github.com/comfforts/comff-stores/pkg/utils/test"
)

const TEST_DIR = "test-data"

func TestMultipleNodes(t *testing.T) {
	var stores []*DistributedStores
	nodeCount := 3
	ports := dynaport.Get(nodeCount)

	// remove data directory after test ends
	defer func() {
		err := os.RemoveAll(TEST_DIR)
		require.NoError(t, err)
	}()

	getConfig := func(ln net.Listener, dataDir string, id int) Config {
		config := Config{}
		config.Raft.StreamLayer = NewStreamLayer(ln, nil, nil)
		config.Raft.LocalID = raft.ServerID(fmt.Sprintf("%d", id))
		config.Raft.HeartbeatTimeout = 500 * time.Millisecond
		config.Raft.ElectionTimeout = 500 * time.Millisecond
		config.Raft.LeaderLeaseTimeout = 500 * time.Millisecond
		config.Raft.CommitTimeout = 5 * time.Millisecond
		config.Segment.MaxIndexSize = 10
		config.Segment.InitialOffset = 1
		config.Raft.BindAddr = ln.Addr().String()

		config.Logger = logging.NewTestAppLogger(dataDir)

		if id == 0 {
			config.Raft.Bootstrap = true
		}
		return config
	}

	// configure distributed store nodes
	for i := 0; i < nodeCount; i++ {
		// setup data directory
		dataDir := filepath.Join(TEST_DIR, fmt.Sprintf("node-%d/", i))
		err := os.MkdirAll(dataDir, os.ModePerm)
		require.NoError(t, err)

		ln, err := net.Listen(
			"tcp",
			fmt.Sprintf("127.0.0.1:%d", ports[i]),
		)
		require.NoError(t, err)

		config := getConfig(ln, dataDir, i)

		ds, err := NewDistributedStores(dataDir, config)
		require.NoError(t, err)

		if i != 0 {
			// join node with cluster
			err = stores[0].Join(fmt.Sprintf("%d", i), ln.Addr().String())
			require.NoError(t, err)
		} else {
			// bootstrap cluster with first node
			err = ds.WaitForLeader(3 * time.Second)
			require.NoError(t, err)
		}

		stores = append(stores, ds)
	}
	// skip the cycle and wait for bit
	time.Sleep(50 * time.Millisecond)

	ctx := context.Background()
	servers, err := stores[0].GetServers(ctx)
	require.NoError(t, err)
	require.Equal(t, len(servers), 3)
	require.Equal(t, servers[0].IsLeader, true)
	require.Equal(t, servers[1].IsLeader, false)
	require.Equal(t, servers[2].IsLeader, false)

	recs := testUtils.CreateStoreModelList()

	for _, v := range recs {
		// add a store through leader node
		ctx := context.Background()
		st, err := stores[0].AddStore(ctx, v)
		require.NoError(t, err)

		// skip the cycle and wait for bit
		time.Sleep(50 * time.Millisecond)

		// verify all nodes nodes have data synced
		for j := 0; j < len(stores); j++ {
			ctx := context.Background()
			got, err := stores[j].GetStore(ctx, st.ID)
			require.NoError(t, err)
			require.Equal(t, st.ID, got.ID)
			require.Equal(t, st.Name, got.Name)
		}
	}

	// remove a follower node
	err = stores[0].Leave(stores[1].ServerId())
	require.NoError(t, err)

	// skip the cycle and wait for bit
	time.Sleep(50 * time.Millisecond)

	ctx = context.Background()
	servers, err = stores[0].GetServers(ctx)
	require.NoError(t, err)
	require.Equal(t, len(servers), 2)
	require.Equal(t, servers[0].IsLeader, true)
	require.Equal(t, servers[1].IsLeader, false)

	// skip the cycle and wait for bit
	time.Sleep(50 * time.Millisecond)

	v := testUtils.CreateStoreModel()

	// adding a store through follower node should fail
	ctx = context.Background()
	_, err = stores[2].AddStore(ctx, v)
	require.NotNil(t, err)

	// add a new store through leader node
	stn, err := stores[0].AddStore(ctx, v)
	require.NoError(t, err)

	// getting new store through removed follower node should fail
	gt, err := stores[1].GetStore(ctx, stn.ID)
	require.NotNil(t, err)
	require.Nil(t, gt)

	time.Sleep(50 * time.Millisecond)

	// getting new store through active follower node should succeed
	got, err := stores[2].GetStore(ctx, stn.ID)
	require.NoError(t, err)
	require.Equal(t, stn.ID, got.ID)

	// join a new to cluster
	dataDir := filepath.Join(TEST_DIR, fmt.Sprintf("node-%d/", 3))
	err = os.MkdirAll(dataDir, os.ModePerm)
	require.NoError(t, err)

	newPorts := dynaport.Get(1)
	ln, err := net.Listen(
		"tcp",
		fmt.Sprintf("127.0.0.1:%d", newPorts[0]),
	)
	require.NoError(t, err)

	config := getConfig(ln, dataDir, 3)
	ds, err := NewDistributedStores(dataDir, config)
	require.NoError(t, err)

	err = stores[0].Join(fmt.Sprintf("%d", 3), ln.Addr().String())
	require.NoError(t, err)

	stores = append(stores, ds)

	time.Sleep(50 * time.Millisecond)

	// TODO getting new store through new follower node should succeed
	got, err = stores[3].GetStore(ctx, stn.ID)
	require.NoError(t, err)
	require.Equal(t, stn.ID, got.ID)

	// shutdown all nodes
	for j := 0; j < len(stores); j++ {
		err := stores[j].Close()
		require.NoError(t, err)
	}
}
