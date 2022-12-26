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

	"github.com/comfforts/comff-stores/internal/server"
	appConfig "github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/constants"
	"github.com/comfforts/comff-stores/pkg/jobs"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/filestorage"
	"github.com/comfforts/comff-stores/pkg/services/geocode"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()

	// initialize app logger instance
	logCfg := &logging.AppLoggerConfig{
		FilePath: "logs/app.log",
		Level:    zapcore.DebugLevel,
	}
	logger := logging.NewAppLogger(nil, logCfg)

	// create app config
	appCfg, err := appConfig.GetAppConfig(logger, "")
	if err != nil {
		logger.Fatal("unable to setup config", zap.Error(err))
		return
	}

	csc, err := filestorage.NewCloudStorageClient(logger, appCfg.Services.CloudStorageClientCfg)
	if err != nil {
		logger.Error("error creating cloud storage client", zap.Error(err))
	}

	// create geo coding service instance
	geoServ, err := geocode.NewGeoCodeService(appCfg.Services.GeoCodeCfg, csc, logger)
	if err != nil {
		logger.Fatal("error initializing maps client", zap.Error(err))
		return
	}

	// initialize store service instance
	storeServ := store.NewStoreService(logger)

	storeLoader, err := jobs.NewStoreLoader(appCfg.Jobs.StoreLoaderConfig, storeServ, csc, logger)
	if err != nil {
		logger.Error("error creating store loader", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
	}

	servCfg := &server.Config{
		StoreService: storeServ,
		GeoService:   geoServ,
		StoreLoader:  storeLoader,
		Logger:       logger,
	}

	// initialize grpc server instance
	server, err := server.NewGrpcServer(servCfg)
	if err != nil {
		logger.Error("error starting server", zap.Error(err))
		panic(err)
	}

	servicePort := fmt.Sprintf(":%d", constants.SERVICE_PORT)
	listener, err := net.Listen("tcp", servicePort)
	if err != nil {
		panic(err)
	}

	// start listening for rpc requests
	go func() {
		logger.Info("server will start listening for requests", zap.String("port", listener.Addr().String()))
		if err := server.Serve(listener); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Error("server failed to start serving", zap.Error(err), zap.Any("errorType", reflect.TypeOf(err)))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
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
