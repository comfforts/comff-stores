package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/comfforts/comff-stores/internal/auth"
	"github.com/comfforts/comff-stores/internal/config"
	"github.com/comfforts/comff-stores/internal/server"
	appConfig "github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/jobs"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/filestorage"
	"github.com/comfforts/comff-stores/pkg/services/geocode"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	ctx := context.Background()

	fmt.Println("  initializing app logger instance")
	logCfg := &logging.AppLoggerConfig{
		FilePath: "logs/app.log",
		Level:    zapcore.DebugLevel,
	}
	logger := logging.NewAppLogger(nil, logCfg)

	logger.Info("creating app configuration")
	appCfg, err := appConfig.GetAppConfig(logger, "")
	if err != nil {
		logger.Fatal("unable to setup config", zap.Error(err))
		return
	}

	logger.Info("creating cloud storage client")
	csc, err := filestorage.NewCloudStorageClient(logger, appCfg.Services.CloudStorageClientCfg)
	if err != nil {
		logger.Error("error creating cloud storage client", zap.Error(err))
	}

	logger.Info("creating geo coding service instance")
	geoServ, err := geocode.NewGeoCodeService(appCfg.Services.GeoCodeCfg, csc, logger)
	if err != nil {
		logger.Fatal("error initializing maps client", zap.Error(err))
		return
	}

	logger.Info("initializing store service instance")
	storeServ := store.NewStoreService(logger)

	logger.Info("initializing store loader instance")
	storeLoader, err := jobs.NewStoreLoader(appCfg.Jobs.StoreLoaderConfig, storeServ, csc, logger)
	if err != nil {
		logger.Error("error creating store loader", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
	}

	logger.Info("initializing authorizer instance")
	authorizer, err := auth.NewAuthorizer(config.PolicyFile(config.ACLModelFile), config.PolicyFile(config.ACLPolicyFile), logger)
	if err != nil {
		logger.Error("error initializing authorizer instance", zap.Error(err))
		panic(err)
	}

	logger.Info("setting up server config")
	servCfg := &server.Config{
		StoreService: storeServ,
		GeoService:   geoServ,
		StoreLoader:  storeLoader,
		Authorizer:   authorizer,
		Logger:       logger,
	}

	logger.Info("opening a tcp socket address")
	servicePort := fmt.Sprintf(":%d", appCfg.ServicePort)
	listener, err := net.Listen("tcp", servicePort)
	if err != nil {
		logger.Error("error opening a tcp socket address", zap.Error(err))
		panic(err)
	}

	logger.Info("setting up server TLS config")
	srvTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.CertFile(config.ServerCertFile),
		KeyFile:       config.CertFile(config.ServerKeyFile),
		CAFile:        config.CertFile(config.CAFile),
		ServerAddress: listener.Addr().String(),
		Server:        true,
	})
	if err != nil {
		logger.Error("error setting up TLS config", zap.Error(err))
		panic(err)
	}
	srvCreds := credentials.NewTLS(srvTLSConfig)

	logger.Info("initializing grpc server instance")
	server, err := server.NewGRPCServer(servCfg, grpc.Creds(srvCreds))
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
	logger.Info("starting server shutdown")

	logger.Info("clearing server store data")
	storeServ.Clear()

	logger.Info("clearing server geo code data")
	geoServ.Clear()

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
