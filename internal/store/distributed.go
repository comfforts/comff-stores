package store

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"time"

	raftboltdb "github.com/hashicorp/raft-boltdb"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/hashicorp/raft"

	"github.com/comfforts/errors"
	"github.com/comfforts/logger"
	"github.com/comfforts/recorder"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/pkg/models"
	"github.com/comfforts/comff-stores/pkg/services/store"
)

const (
	ERROR_CREATING_LOG_STORE_DIR  string = "error creating raft log store folder"
	ERROR_CREATING_STABLE_STORE   string = "error creating raft stable store"
	ERROR_CREATING_SNAPSHOT_STORE string = "error creating raft snapshot store"
	ERROR_CREATING_RAFT           string = "error creating raft node"
	ERROR_RAFT_JOIN_CONFIG        string = "error getting config  to join raft cluster"
	ERROR_REMOVING_EXISTING       string = "error removing existing node"
	ERROR_RAFT_JOIN               string = "error joining raft cluster"
	ERROR_RAFT_STATE              string = "error identifying raft node state"
	ERROR_BUF_WRITE               string = "error wrting to buffer"
	ERROR_MARSHAL_PROTO           string = "error marshalling proto"
	ERROR_RAFT_APPLY              string = "error apply raft command"
	ERROR_RAFT_APPLY_RESP         string = "error apply raft command response"
	ERROR_RAFT_TIMEOUT            string = "error: raft timed out"
	ERROR_RAFT_SHUTDOWN           string = "error: raft shutdown"
)

var (
	ErrRaftTimeout = errors.NewAppError(ERROR_RAFT_TIMEOUT)
)

type Config struct {
	Raft struct {
		raft.Config
		BindAddr    string
		StreamLayer *StreamLayer
		Bootstrap   bool
	}
	recorder.Config
	Logger logger.AppLogger
}

type DistributedStores struct {
	config           Config
	stores           models.Stores
	raft             *raft.Raft
	shutdownCallback func() error
}

func NewDistributedStores(dataDir string, config Config) (*DistributedStores, error) {
	config.Logger.Info("creating distributed store instance: ", zap.Any("node id", config.Raft.Config.LocalID))
	l := &DistributedStores{
		config: config,
	}
	if err := l.setupStores(dataDir); err != nil {
		return nil, err
	}
	if err := l.setupRaft(dataDir); err != nil {
		return nil, err
	}
	return l, nil
}

func (ds *DistributedStores) setupStores(dataDir string) error {
	ss, err := store.NewStoreService(ds.config.Logger)
	if err != nil {
		return err
	}
	ds.stores = ss
	return nil
}

func (ds *DistributedStores) setupRaft(dataDir string) error {
	fsm := &fsm{
		DataDir:      dataDir,
		StoreService: ds.stores,
		logger:       ds.config.Logger,
	}

	ds.config.Logger.Info("creating raft log store folder")
	logDir := filepath.Join(dataDir, "raft", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		ds.config.Logger.Error(ERROR_CREATING_LOG_STORE_DIR, zap.Error(err))
		return errors.WrapError(err, ERROR_CREATING_LOG_STORE_DIR)
	}
	logConfig := ds.config

	ds.config.Logger.Info("creating raft log store")
	logStore, err := newLogStore(logDir, logConfig)
	if err != nil {
		return err
	}

	ds.config.Logger.Info("creating raft stable store")
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft", "stable"))
	if err != nil {
		ds.config.Logger.Error(ERROR_CREATING_STABLE_STORE, zap.Error(err))
		return errors.WrapError(err, ERROR_CREATING_STABLE_STORE)
	}

	ds.config.Logger.Info("creating raft snapshot store")
	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(dataDir, "raft"),
		retain,
		os.Stderr,
	)
	if err != nil {
		ds.config.Logger.Error(ERROR_CREATING_SNAPSHOT_STORE, zap.Error(err))
		return errors.WrapError(err, ERROR_CREATING_SNAPSHOT_STORE)
	}

	ds.config.Logger.Info("setup raft streamlayer transport")
	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
		ds.config.Raft.StreamLayer,
		maxPool,
		timeout,
		os.Stderr,
	)

	ds.config.Logger.Info("setup raft config")
	config := raft.DefaultConfig()
	config.LocalID = ds.config.Raft.LocalID
	if ds.config.Raft.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = ds.config.Raft.HeartbeatTimeout
	}
	if ds.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = ds.config.Raft.ElectionTimeout
	}
	if ds.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = ds.config.Raft.LeaderLeaseTimeout
	}
	if ds.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = ds.config.Raft.CommitTimeout
	}
	if ds.config.Raft.SnapshotInterval != 0 {
		config.SnapshotInterval = ds.config.Raft.SnapshotInterval
	}
	if ds.config.Raft.SnapshotThreshold != 0 {
		config.SnapshotThreshold = ds.config.Raft.SnapshotThreshold
	}

	ds.config.Logger.Info("creating raft node")
	ds.raft, err = raft.NewRaft(
		config,
		fsm,
		logStore,
		stableStore,
		snapshotStore,
		transport,
	)
	if err != nil {
		ds.config.Logger.Error(ERROR_CREATING_RAFT, zap.Error(err))
		return errors.WrapError(err, ERROR_CREATING_RAFT)
	}

	ds.config.Logger.Info("identifying raft node state")
	hasState, err := raft.HasExistingState(
		logStore,
		stableStore,
		snapshotStore,
	)
	if err != nil {
		ds.config.Logger.Error(ERROR_RAFT_STATE, zap.Error(err))
		return errors.WrapError(err, ERROR_RAFT_STATE)
	}
	ds.config.Logger.Info("raft node state", zap.Bool("is-bootstrap", ds.config.Raft.Bootstrap), zap.Bool("has-state", hasState), zap.String("bind-addr", ds.config.Raft.BindAddr), zap.Any("transport-addr", transport.LocalAddr()), zap.Any("raft-addr", raft.ServerAddress(ds.config.Raft.BindAddr)))
	if ds.config.Raft.Bootstrap && !hasState {
		config := raft.Configuration{
			Servers: []raft.Server{{
				ID:      config.LocalID,
				Address: raft.ServerAddress(ds.config.Raft.BindAddr),
			}},
		}
		ds.config.Logger.Info("bootstrapping raft cluster")
		err = ds.raft.BootstrapCluster(config).Error()
	}

	ds.shutdownCallback = func() error {
		err := logStore.Close()
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

func (ds *DistributedStores) AddStore(ctx context.Context, s *models.Store) (*models.Store, error) {
	ds.config.Logger.Debug("distributed store AddStore()", zap.String("node", string(ds.config.Raft.LocalID)))
	res, err := ds.apply(
		AddStoreRequestType,
		&api.AddStoreRequest{
			Name:      s.Name,
			Org:       s.Org,
			City:      s.City,
			Country:   s.Country,
			Longitude: float32(s.Longitude),
			Latitude:  float32(s.Latitude),
			StoreId:   s.StoreId,
		},
	)
	if err != nil {
		return nil, err
	}
	store := models.MapProtoToStore(res.(*api.AddStoreResponse).Store)
	return store, nil
}

func (ds *DistributedStores) apply(reqType RequestType, req proto.Message) (
	interface{},
	error,
) {
	ds.config.Logger.Debug("distributed store apply()", zap.String("node", string(ds.config.Raft.LocalID)))
	var buf bytes.Buffer
	_, err := buf.Write([]byte{byte(reqType)})
	if err != nil {
		ds.config.Logger.Error(ERROR_BUF_WRITE, zap.Error(err))
		return nil, errors.WrapError(err, ERROR_BUF_WRITE)
	}
	b, err := proto.Marshal(req)
	if err != nil {
		ds.config.Logger.Error(ERROR_MARSHAL_PROTO, zap.Error(err))
		return nil, errors.WrapError(err, ERROR_MARSHAL_PROTO)
	}
	_, err = buf.Write(b)
	if err != nil {
		ds.config.Logger.Error(ERROR_BUF_WRITE, zap.Error(err))
		return nil, errors.WrapError(err, ERROR_BUF_WRITE)
	}
	timeout := 10 * time.Second
	future := ds.raft.Apply(buf.Bytes(), timeout)
	if future.Error() != nil {
		ds.config.Logger.Error(ERROR_RAFT_APPLY, zap.Error(future.Error()))
		return nil, future.Error()
	}
	res := future.Response()
	if err, ok := res.(error); ok {
		ds.config.Logger.Error(ERROR_RAFT_APPLY_RESP, zap.Error(err))
		return nil, err
	}
	return res, nil
}

func (ds *DistributedStores) GetStore(ctx context.Context, id string) (*models.Store, error) {
	ds.config.Logger.Debug("distributed store GetStore()", zap.String("node", string(ds.config.Raft.LocalID)))
	return ds.stores.GetStore(ctx, id)
}

func (ds *DistributedStores) GetStoresForGeoPoint(ctx context.Context, lat, long float64, dist int) ([]*models.StoreGeo, error) {
	ds.config.Logger.Debug("distributed store GetStoresForGeoPoint()", zap.String("node", string(ds.config.Raft.LocalID)))
	return ds.stores.GetStoresForGeoPoint(ctx, lat, long, dist)
}

func (ds *DistributedStores) GetStoreStats() models.StoreStats {
	ds.config.Logger.Debug("distributed store GetStoreStats()", zap.String("node", string(ds.config.Raft.LocalID)))
	return ds.stores.GetStoreStats()
}

func (ds *DistributedStores) Join(id, addr string) error {
	ds.config.Logger.Info("distributed store joining", zap.String("node", string(ds.config.Raft.LocalID)))
	configFuture := ds.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		ds.config.Logger.Error(ERROR_RAFT_JOIN_CONFIG, zap.Error(err))
		return err
	}
	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == serverID || srv.Address == serverAddr {
			if srv.ID == serverID && srv.Address == serverAddr {
				// server has already joined
				return nil
			}
			// remove the existing server
			removeFuture := ds.raft.RemoveServer(serverID, 0, 0)
			if err := removeFuture.Error(); err != nil {
				ds.config.Logger.Error(ERROR_REMOVING_EXISTING, zap.Error(err))
				return err
			}
		}
	}

	addFuture := ds.raft.AddNonvoter(serverID, serverAddr, 0, 0)
	if err := addFuture.Error(); err != nil {
		ds.config.Logger.Error(ERROR_RAFT_JOIN, zap.Error(err))
		return err
	}
	return nil
}

func (ds *DistributedStores) Leave(id string) error {
	ds.config.Logger.Info("distributed store leaving", zap.String("node", string(raft.ServerID(id))))
	removeFuture := ds.raft.RemoveServer(raft.ServerID(id), 0, 0)
	return removeFuture.Error()
}

func (ds *DistributedStores) WaitForLeader(timeout time.Duration) error {
	ds.config.Logger.Info("distributed store waiting for leader", zap.String("node", string(ds.config.Raft.LocalID)))
	timeoutc := time.After(timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeoutc:
			ds.config.Logger.Error(ERROR_RAFT_TIMEOUT, zap.Error(ErrRaftTimeout))
			return ErrRaftTimeout
		case <-ticker.C:
			if rLead := ds.raft.Leader(); rLead != "" {
				ds.config.Logger.Debug("leader: ", zap.Any("leader", rLead))
				return nil
			}
		}
	}
}

func (ds *DistributedStores) Reader(ctx context.Context, dataDir string) (*os.File, error) {
	return ds.stores.Reader(ctx, dataDir)
}

func (ds *DistributedStores) GetServers(ctx context.Context) ([]*api.Server, error) {
	future := ds.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}
	var servers []*api.Server
	for _, server := range future.Configuration().Servers {
		servers = append(servers, &api.Server{
			Id:       string(server.ID),
			RpcAddr:  string(server.Address),
			IsLeader: ds.raft.Leader() == server.Address,
		})
	}
	return servers, nil
}

func (ds *DistributedStores) ServerId() string {
	return string(ds.config.Raft.LocalID)
}

func (ds *DistributedStores) Close() error {
	ds.config.Logger.Info("closing distributed store", zap.String("node", string(ds.config.Raft.LocalID)))

	f := ds.raft.Shutdown()
	if err := f.Error(); err != nil {
		ds.config.Logger.Error(ERROR_RAFT_SHUTDOWN, zap.Error(err))
		return errors.WrapError(err, ERROR_RAFT_SHUTDOWN)
	}
	if err := ds.shutdownCallback(); err != nil {
		return err
	}
	ds.stores.Close()
	return nil
}

func (ds *DistributedStores) SetReady(ctx context.Context, ready bool) {
	ds.stores.SetReady(ctx, ready)
}
