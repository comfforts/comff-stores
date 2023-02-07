package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"syscall"
	"time"

	"github.com/comfforts/cloudstorage"
	"github.com/comfforts/geocode"
	"github.com/comfforts/logger"

	"github.com/comfforts/comff-stores/internal/auth"
	"github.com/comfforts/comff-stores/internal/config"
	"github.com/comfforts/comff-stores/internal/server"
	appConfig "github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/jobs"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"go.opencensus.io/examples/exporter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

func main() {
	ctx := context.Background()

	fmt.Println("  initializing app logger instance")
	logCfg := &logger.AppLoggerConfig{
		Level: zapcore.DebugLevel,
		Name:  "comff-stores",
	}
	logger := logger.NewAppLogger(logCfg)

	logger.Info("creating app configuration")
	appCfg, err := appConfig.GetAppConfig("", logger)
	if err != nil {
		logger.Fatal("unable to setup config", zap.Error(err))
		return
	}

	logger.Info("setting up telemetry")
	telemetryCleanupCallbk, err := setupTelemetry(logger)
	if err != nil {
		logger.Error("error setting up telemetery exporter", zap.Error(err))
		panic(err)
	}

	logger.Info("opening a tcp socket address")
	servicePort := fmt.Sprintf(":%d", appCfg.ServicePort)
	listener, err := net.Listen("tcp", servicePort)
	if err != nil {
		logger.Error("error opening a tcp socket address", zap.Error(err))
		panic(err)
	}

	server, serverCleanupCallbk, err := setupServer(appCfg, listener.Addr().String(), logger)
	if err != nil {
		logger.Error("error starting server", zap.Error(err))
		panic(err)
	}

	go func() {
		logger.Info("server will start listening for requests", zap.String("port", listener.Addr().String()))
		if err := server.Serve(listener); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Error("server failed to start serving", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
		}
	}()

	logger.Info("waiting for interrupt signal to gracefully shutdown the server")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("starting server shutdown, flushing telemetry metrics, cache and clearing up stores")
	serverCleanupCallbk()
	telemetryCleanupCallbk()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer func() {
		if err := listener.Close(); err != nil {
			logger.Error("error closing network listener", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
		}
		logger.Info("stopping server")
		server.GracefulStop()
		cancel()
	}()
	<-ctx.Done()
	logger.Info("server exiting")
}

// sets up server and returns server cleanup callback
func setupServer(appCfg *appConfig.Configuration, addr string, logger logger.AppLogger) (*grpc.Server, func(), error) {
	logger.Info("creating cloud storage client")

	cscCfg := cloudstorage.CloudStorageClientConfig{
		CredsPath: appCfg.Services.CloudStorageClientCfg.CredsPath,
	}

	csc, err := cloudstorage.NewCloudStorageClient(cscCfg, logger)
	if err != nil {
		logger.Error("error creating cloud storage client", zap.Error(err))
	}

	logger.Info("creating geo coding service instance")
	gscCfg := geocode.GeoCodeServiceConfig{
		DataDir:     appCfg.Services.GeoCodeCfg.DataDir,
		BucketName:  appCfg.Services.GeoCodeCfg.BucketName,
		GeocoderKey: appCfg.Services.GeoCodeCfg.GeocoderKey,
	}
	geoServ, err := geocode.NewGeoCodeService(gscCfg, csc, logger)
	if err != nil {
		logger.Fatal("error initializing maps client", zap.Error(err))
		return nil, nil, err
	}

	logger.Info("initializing store service instance")
	storeServ, err := store.NewStoreService(logger)
	if err != nil {
		logger.Fatal("error initializing store service", zap.Error(err))
		return nil, nil, err
	}

	callbk := func() {
		logger.Info("clearing server store data")
		storeServ.Close()

		logger.Info("clearing server geo code data")
		geoServ.Clear()
	}

	logger.Info("initializing store loader instance")
	storeLoader, err := jobs.NewStoreLoader(appCfg.Jobs.StoreLoaderConfig, storeServ, csc, logger)
	if err != nil {
		logger.Error("error creating store loader", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
	}

	logger.Info("setting up authorizer")
	authorizer, err := auth.NewAuthorizer(config.PolicyFile(config.ACLModelFile), config.PolicyFile(config.ACLPolicyFile), logger)
	if err != nil {
		logger.Error("error initializing authorizer instance", zap.Error(err))
		return nil, nil, err
	}

	logger.Info("setting up server config")
	servCfg := &server.Config{
		StoreService: storeServ,
		GeoService:   geoServ,
		StoreLoader:  storeLoader,
		Authorizer:   authorizer,
		Logger:       logger,
	}

	logger.Info("setting up server TLS config")
	srvTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.CertFile(config.ServerCertFile),
		KeyFile:       config.CertFile(config.ServerKeyFile),
		CAFile:        config.CertFile(config.CAFile),
		ServerAddress: addr,
		Server:        true,
	})
	if err != nil {
		logger.Error("error setting up TLS config", zap.Error(err))
		return nil, callbk, err
	}
	srvCreds := credentials.NewTLS(srvTLSConfig)

	logger.Info("initializing grpc server instance")
	server, err := server.NewGRPCServer(servCfg, grpc.Creds(srvCreds))
	if err != nil {
		logger.Error("error initializing server", zap.Error(err))
		return nil, callbk, err
	}

	return server, callbk, nil
}

// sets up telemetry and returns telemetery cleanup callback
func setupTelemetry(logger logger.AppLogger) (func(), error) {
	mfPath := filepath.Join("logs", "metrics.log")

	if err := fileUtils.CreateDirectory(mfPath); err != nil {
		logger.Error("error setting up metrics store", zap.Error(err))
		return nil, err
	}

	metricsLogFile, err := os.Create(mfPath)
	if err != nil {
		logger.Error("error creating metrics log file", zap.Error(err))
		return nil, err
	}
	logger.Info("metrics log file ", zap.String("name", metricsLogFile.Name()))

	tfPath := filepath.Join("logs", "traces.log")
	tracesLogFile, err := os.Create(tfPath)
	if err != nil {
		logger.Error("error creating traces log file", zap.Error(err))
		return nil, err
	}
	logger.Info("traces log file ", zap.String("name", tracesLogFile.Name()))

	telemetryExporter, err := exporter.NewLogExporter(exporter.Options{
		MetricsLogFile:    metricsLogFile.Name(),
		TracesLogFile:     tracesLogFile.Name(),
		ReportingInterval: time.Second,
	})
	if err != nil {
		logger.Error("error initializing telemetery exporter", zap.Error(err))
		return nil, err
	}

	err = telemetryExporter.Start()
	if err != nil {
		logger.Error("error starting telemetery exporter", zap.Error(err))
		return nil, err
	}
	return func() {
		telemetryExporter.Stop()
		telemetryExporter.Close()
	}, nil
}
