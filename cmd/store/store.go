package main

import (
	"context"
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
	"github.com/comfforts/comff-stores/pkg/services/geocode"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()
	servicePort := fmt.Sprintf(":%d", constants.SERVICE_PORT)
	listener, err := net.Listen("tcp", servicePort)
	if err != nil {
		panic(err)
	}

	// initialize app logger instance
	logCfg := &logging.AppLoggerConfig{
		FilePath: "logs/app.log",
		Level:    zapcore.DebugLevel,
	}
	logger := logging.NewAppLogger(nil, logCfg)

	// initialize store service instance
	storeServ := store.NewStoreService(logger)

	// create app config,
	// throws app startup error if required config is missing
	appCfg, err := appConfig.GetAppConfig(logger, "")
	if err != nil {
		logger.Fatal("unable to setup config", zap.Error(err))
		return
	}

	// create geo coding service instance,
	// requires config and logger instance
	// throws app startup error
	geoServ, err := geocode.NewGeoCodeService(appCfg.Services.GeoCodeCfg, logger)
	if err != nil {
		logger.Fatal("error initializing maps client", zap.Error(err))
		return
	}

	storeLoader, err := jobs.NewStoreLoader(appCfg.Jobs.StoreLoaderConfig, storeServ, logger)
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

	// start listening for rpc requests
	go func() {
		logger.Info("server will start listening for requests", zap.Int("port", constants.SERVICE_PORT))
		if err := server.Serve(listener); err != nil {
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
		server.Stop()
		cancel()
	}()
	<-ctx.Done()
	logger.Info("server exiting")
}
