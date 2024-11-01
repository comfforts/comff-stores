package agent

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/soheilhy/cmux"
	"go.opencensus.io/examples/exporter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/comfforts/cloudstorage"
	appErrors "github.com/comfforts/errors"
	"github.com/comfforts/geocode"
	"github.com/comfforts/logger"

	"github.com/comfforts/comff-stores/internal/auth"
	"github.com/comfforts/comff-stores/internal/discovery"
	"github.com/comfforts/comff-stores/internal/server"
	"github.com/comfforts/comff-stores/internal/store"
	appConfig "github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/jobs"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

type Config struct {
	ServerTLSConfig *tls.Config
	PeerTLSConfig   *tls.Config
	// RunDir is the root directory.
	RunDir string
	// DataDir stores the log and raft data.
	DataDir string
	// BindAddr is the address serf runs on.
	BindAddr string
	// RPCPort is the port for client (and Raft) connections.
	RPCPort int
	// Raft server id.
	NodeName string
	// Bootstrap should be set to true when starting the first node of the cluster.
	PeerNodeAddrs []string
	ACLModelFile  string
	ACLPolicyFile string
	Bootstrap     bool
	// Application config file path.
	AppConfigFile string
	// HeartbeatTimeout specifies the time in follower state without contact
	// from a leader before we attempt an election.
	HeartbeatTimeout time.Duration
	// ElectionTimeout specifies the time in candidate state without contact
	// from a leader before we attempt an election.
	ElectionTimeout time.Duration
	// SnapshotInterval controls how often we check if we should perform a
	// snapshot. We randomly stagger between this value and 2x this value to avoid
	// the entire cluster from performing a snapshot at once. The value passed
	// here is the initial setting used. This can be tuned during operation using
	// ReloadConfig.
	SnapshotInterval time.Duration
	// SnapshotThreshold controls how many outstanding logs there must be before
	// we perform a snapshot. This is to prevent excessive snapshotting by
	// replaying a small set of logs instead. The value passed here is the initial
	// setting used. This can be tuned during operation using ReloadConfig.
	SnapshotThreshold uint64
	// LeaderLeaseTimeout is used to control how long the "lease" lasts
	// for being the leader without being able to contact a quorum
	// of nodes. If we reach this interval without contact, we will
	// step down as leader.
	LeaderLeaseTimeout time.Duration
	// MaxIndexSize specifies the maximum number of entries in a segment.
	MaxIndexSize uint64
	// InitialOffset specifies the starting offset
	InitialOffset uint64
}

type Agent struct {
	Config
	logger  logger.AppLogger
	DataDir string

	mux    cmux.CMux
	stores *store.DistributedStores

	server                   *grpc.Server
	membership               *discovery.Membership
	shutdown                 bool
	shutdowns                chan struct{}
	shutdownLock             sync.Mutex
	serverShutdownCallback   func()
	telemetryCleanupCallback func()
}

func NewAgent(config Config) (*Agent, error) {
	if config.DataDir == "" {
		return nil, appErrors.NewAppError("missing log data directory configuration")
	}

	if config.AppConfigFile == "" {
		config.AppConfigFile = path.Join(config.RunDir, appConfig.CONFIG_FILE_NAME)
	}

	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}

	a.telemetryCleanupCallback = func() {
		a.logger.Info("telemetry cleanup callback")
	}
	setup := []func() error{
		a.setupLogger,
		a.setupMux,
		a.setupDistributedStores,
		a.setupTelemetry,
		a.setupServer,
		a.setupMembership,
	}
	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}
	go a.serve()
	return a, nil
}

func (a *Agent) setupLogger() error {
	filePath := filepath.Join(a.Config.DataDir, "logs", "agent.log")

	logCfg := &logger.AppLoggerConfig{
		Level:    zapcore.DebugLevel,
		FilePath: filePath,
		Name:     "comff-stores-agent",
	}

	appLogger := logger.NewAppLogger(logCfg)

	a.logger = appLogger
	return nil
}

func (a *Agent) setupMux() error {
	addr, err := net.ResolveTCPAddr("tcp", a.Config.BindAddr)
	if err != nil {
		return err
	}
	rpcAddr := fmt.Sprintf(
		"%s:%d",
		addr.IP.String(),
		a.Config.RPCPort,
	)
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}
	a.mux = cmux.New(ln)
	return nil
}

func (a *Agent) setupDistributedStores() error {
	raftLn := a.mux.Match(func(reader io.Reader) bool {
		b := make([]byte, 1)
		if _, err := reader.Read(b); err != nil {
			return false
		}
		return bytes.Compare(b, []byte{byte(store.RaftRPC)}) == 0
	})

	a.logger.Info("building distributed store config")
	cfg := store.Config{}
	cfg.Raft.StreamLayer = store.NewStreamLayer(
		raftLn,
		a.Config.ServerTLSConfig,
		a.Config.PeerTLSConfig,
		a.logger,
	)

	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}
	cfg.Raft.BindAddr = rpcAddr
	cfg.Raft.LocalID = raft.ServerID(a.Config.NodeName)

	cfg.Raft.Bootstrap = a.Config.Bootstrap

	cfg.Raft.CommitTimeout = 5 * time.Millisecond
	cfg.Raft.HeartbeatTimeout = 20 * time.Millisecond
	if a.Config.HeartbeatTimeout > 0 {
		cfg.Raft.HeartbeatTimeout = a.Config.HeartbeatTimeout
	}
	cfg.Raft.ElectionTimeout = 20 * time.Millisecond
	if a.Config.ElectionTimeout > 0 {
		cfg.Raft.ElectionTimeout = a.Config.ElectionTimeout
	}
	cfg.Raft.LeaderLeaseTimeout = 20 * time.Millisecond
	if a.Config.LeaderLeaseTimeout > 0 {
		cfg.Raft.LeaderLeaseTimeout = a.Config.LeaderLeaseTimeout
	}
	cfg.Raft.SnapshotThreshold = 5
	if a.Config.SnapshotThreshold > 0 {
		cfg.Raft.SnapshotThreshold = a.Config.SnapshotThreshold
	}
	cfg.Raft.SnapshotInterval = 5 * time.Second
	if a.Config.SnapshotInterval > 0 {
		cfg.Raft.SnapshotInterval = a.Config.SnapshotInterval
	}
	cfg.Segment.MaxIndexSize = 20
	if a.Config.MaxIndexSize > 0 {
		cfg.Segment.MaxIndexSize = a.Config.MaxIndexSize
	}
	cfg.Segment.InitialOffset = 1
	if a.Config.InitialOffset > 0 {
		cfg.Segment.InitialOffset = a.Config.InitialOffset
	}
	cfg.Logger = a.logger

	a.logger.Info("building distributed store instance")
	a.stores, err = store.NewDistributedStores(
		a.Config.DataDir,
		cfg,
	)
	if err != nil {
		return err
	}

	if a.Config.Bootstrap {
		a.logger.Info("config is bootstrap, initiating wait for leader")
		err = a.stores.WaitForLeader(3 * time.Second)
	}
	return err
}

// sets up telemetry and returns telemetery cleanup callback
func (a *Agent) setupTelemetry() error {
	mfPath := filepath.Join(a.Config.DataDir, "logs", "metrics.log")

	if err := fileUtils.CreateDirectory(mfPath); err != nil {
		a.logger.Error("error setting up metrics store", zap.Error(err))
		return err
	}

	metricsLogFile, err := os.Create(mfPath)
	if err != nil {
		a.logger.Error("error creating metrics log file", zap.Error(err))
		return err
	}
	a.logger.Info("metrics log file ", zap.String("name", metricsLogFile.Name()))

	tfPath := filepath.Join(a.Config.DataDir, "logs", "traces.log")
	tracesLogFile, err := os.Create(tfPath)
	if err != nil {
		a.logger.Error("error creating traces log file", zap.Error(err))
		return err
	}
	a.logger.Info("traces log file ", zap.String("name", tracesLogFile.Name()))

	telemetryExporter, err := exporter.NewLogExporter(exporter.Options{
		MetricsLogFile:    metricsLogFile.Name(),
		TracesLogFile:     tracesLogFile.Name(),
		ReportingInterval: time.Second,
	})
	if err != nil {
		a.logger.Error("error initializing telemetery exporter", zap.Error(err))
		return err
	}

	err = telemetryExporter.Start()
	if err != nil {
		a.logger.Error("error starting telemetery exporter", zap.Error(err))
		return err
	}

	a.telemetryCleanupCallback = func() {
		telemetryExporter.Stop()
		telemetryExporter.Close()
	}
	return nil
}

func (a *Agent) setupServer() error {
	a.logger.Info("creating app configuration")
	appCfg, err := appConfig.GetAppConfig(a.AppConfigFile, a.logger)
	if err != nil {
		a.logger.Error("unable to setup config", zap.Error(err))
		return err
	}

	a.logger.Info("creating cloud storage client")
	cscCfg := cloudstorage.CloudStorageClientConfig{
		CredsPath: appCfg.Services.CloudStorageClientCfg.CredsPath,
	}
	csc, err := cloudstorage.NewCloudStorageClient(cscCfg, a.logger)
	if err != nil {
		a.logger.Error("error creating cloud storage client", zap.Error(err))
	}

	a.logger.Info("creating geo coding service instance")
	gscCfg := geocode.GeoCodeServiceConfig{
		DataDir:     a.Config.DataDir,
		BucketName:  appCfg.Services.GeoCodeCfg.BucketName,
		GeocoderKey: appCfg.Services.GeoCodeCfg.GeocoderKey,
	}
	geoServ, err := geocode.NewGeoCodeService(gscCfg, csc, a.logger)
	if err != nil {
		a.logger.Error("error initializing maps client", zap.Error(err))
		return err
	}

	a.serverShutdownCallback = func() {
		a.logger.Info("clearing server store data")
		a.stores.Close()

		a.logger.Info("clearing server geo code data")
		geoServ.Clear()
	}

	a.logger.Info("initializing store loader instance")
	appCfg.Jobs.StoreLoaderConfig.DataDir = a.Config.DataDir
	storeLoader, err := jobs.NewStoreLoader(appCfg.Jobs.StoreLoaderConfig, a.stores, csc, a.logger)
	if err != nil {
		a.logger.Error("error creating store loader", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
	}

	a.logger.Info("setting up authorizer")
	authorizer, err := auth.NewAuthorizer(a.ACLModelFile, a.ACLPolicyFile, a.logger)
	if err != nil {
		a.logger.Error("error initializing authorizer instance", zap.Error(err))
		return err
	}

	a.logger.Info("setting up server config")
	servCfg := &server.Config{
		StoreService: a.stores,
		GeoService:   geoServ,
		StoreLoader:  storeLoader,
		Authorizer:   authorizer,
		Servicer:     a.stores,
		Logger:       a.logger,
	}

	var opts []grpc.ServerOption
	if a.Config.ServerTLSConfig != nil {
		creds := credentials.NewTLS(a.Config.ServerTLSConfig)
		opts = append(opts, grpc.Creds(creds))
	}
	a.server, err = server.NewGRPCServer(servCfg, opts...)
	if err != nil {
		a.logger.Error("error initializing server", zap.Error(err))
		return err
	}
	grpcLn := a.mux.Match(cmux.Any())

	go func() {
		a.logger.Info("server will start listening for requests", zap.String("port", grpcLn.Addr().String()))
		if err := a.server.Serve(grpcLn); err != nil && !errors.Is(err, cmux.ErrServerClosed) {
			a.logger.Error("server failed to start serving", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
			_ = a.Shutdown()
		}
	}()

	return err
}

func (a *Agent) setupMembership() error {
	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}

	a.membership, err = discovery.NewMembership(a.stores, discovery.Config{
		NodeName: a.Config.NodeName,
		BindAddr: a.Config.BindAddr,
		Tags: map[string]string{
			"rpc_addr": rpcAddr,
		},
		PeerNodeAddrs: a.Config.PeerNodeAddrs,
		Logger:        a.logger,
	})

	return err
}

func (a *Agent) serve() error {
	a.logger.Info("agent will start serving", zap.String("bindAddr", a.BindAddr), zap.Int("rpcPort", a.RPCPort))
	if err := a.mux.Serve(); err != nil {
		_ = a.Shutdown()
		return err
	}
	return nil
}

func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}

func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()
	if a.shutdown {
		return nil
	}
	a.shutdown = true
	close(a.shutdowns)

	err := a.membership.Leave()
	if err != nil {
		return err
	}

	a.telemetryCleanupCallback()

	a.serverShutdownCallback()
	a.server.GracefulStop()

	a.logger.Info("Agent shut down")
	fmt.Printf("/n/n")

	return nil
}
