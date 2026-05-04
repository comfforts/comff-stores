package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	config "github.com/comfforts/comff-config"
	geocl "github.com/comfforts/comff-geo/clients/go"
	"github.com/comfforts/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	grpchandler "github.com/comfforts/comff-stores/internal/delivery/stores/grpc_handler"
	"github.com/comfforts/comff-stores/internal/infra/mongostore"
	"github.com/comfforts/comff-stores/internal/infra/observability"
	strepo "github.com/comfforts/comff-stores/internal/repo/stores"
	"github.com/comfforts/comff-stores/internal/usecase/services/stores"
	envutils "github.com/comfforts/comff-stores/pkg/utils/environ"
)

const SERVICE_PORT = 62151
const DEFAULT_SERVICE_HOST = "stores-service"

func main() {
	// Initialize logger
	nodeName := DEFAULT_SERVICE_HOST
	if host, err := os.Hostname(); err == nil && host != "" {
		nodeName = host
	}
	l := logger.GetSlogLogger().With(
		"service", "stores-server",
		"component", "server",
		"node", nodeName,
		"env", os.Getenv("ENV"),
		"go_env", os.Getenv("GO_ENV"),
		"infra", os.Getenv("INFRA"),
	)

	l.Info("stores server starting")

	// Set up server port from environment variable or use default
	serverPort, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil || serverPort == 0 {
		if err != nil {
			l.Debug("error parsing server port, using default", "error", err.Error())
		} else {
			l.Debug("no server port provided, using default server port")
		}
		serverPort = SERVICE_PORT
	}
	servicePort := fmt.Sprintf(":%d", serverPort)

	// Set up tcp socket listener on the specified port
	l.Info("opening tcp listener", "listen_addr", servicePort)
	listener, err := net.Listen("tcp", servicePort)
	if err != nil {
		l.Error("failed to open tcp listener", "listen_addr", servicePort, "error", err.Error())
		panic(err)
	}

	metrics, err := observability.NewMetrics()
	if err != nil {
		l.Error("failed to initialize metrics", "error", err.Error())
		panic(err)
	}

	startCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	startCtx = logger.WithLogger(startCtx, l)

	// Initialize MongoDB store
	nmCfg := envutils.BuildMongoStoreConfig(true)
	ms, err := mongostore.NewMongoStore(startCtx, nmCfg)
	if err != nil {
		l.Error("failed to initialize mongo store", "error", err.Error())
		panic(err)
	}

	// Initialize stores repository
	sr, err := strepo.NewStoresRepo(startCtx, ms, metrics)
	if err != nil {
		l.Error("failed to initialize stores repository", "error", err.Error())
		panic(err)
	}

	// Initialize geo client options
	clientOpts := geocl.NewDefaultClientOption()
	clientOpts.Caller = "stores-service-geo-client"

	// Initialize geo client
	gc, err := geocl.NewClient(startCtx, clientOpts)
	if err != nil {
		l.Error("failed to initialize geo client", "error", err.Error())
		panic(err)
	}

	// Initialize stores service
	ss, err := stores.NewStoresService(startCtx, sr, gc, metrics)
	if err != nil {
		l.Error("failed to initialize stores service", "error", err.Error())
		panic(err)
	}

	// Build gRPC server config
	cfg, err := grpchandler.BuildServerConfig(startCtx, ss)
	if err != nil {
		l.Error("failed to build gRPC server config", "error", err.Error())
		panic(err)
	}

	srvTLSCfg := envutils.BuildServerTLSConfig()

	// Server TLS config
	srvTLSConfig, err := config.SetupTLSConfig(&config.ConfigOpts{
		Target: config.SERVER,
		Addr:   listener.Addr().String(),
		Opts: &config.CustomOpts{
			CAFilePath:   srvTLSCfg.CAFilePath,
			CertFilePath: srvTLSCfg.CertFilePath,
			KeyFilePath:  srvTLSCfg.KeyFilePath,
		},
	})
	if err != nil {
		l.Error("failed to set up TLS config", "error", err.Error())
		panic(err)
	}
	srvCreds := credentials.NewTLS(srvTLSConfig)

	// Initialize observability (metrics server and tracing)
	metricAddr, otelEndpoint := envutils.BuildMetricsConfig()
	metricsOpts := observability.InitOptions{
		ServiceName:  "stores-server",
		MetricsAddr:  metricAddr,
		OTLPEndpoint: otelEndpoint,
	}
	l.Info(
		"initializing observability",
		"metrics_addr", metricAddr,
		"otlp_endpoint", otelEndpoint,
	)
	metricsServer, err := observability.NewMetricsServer(startCtx, metricsOpts)
	if err != nil {
		l.Error("error initializing metrics server", "error", err.Error())
		panic(err)
	}

	// Initialize the gRPC server with the configuration and TLS credentials
	l.Info("initializing stores grpc server instance")
	server, err := grpchandler.NewGRPCServer(cfg, grpc.Creds(srvCreds))
	if err != nil {
		l.Error("error initializing stores grpc server", "error", err.Error())
		panic(err)
	}
	l.Info("stores grpc server initialized")

	// Start the gRPC server in a goroutine
	go func() {
		l.Info(
			"stores grpc server serving",
			"listen_addr", listener.Addr().String(),
			"service_info", server.GetServiceInfo(),
		)
		if err := server.Serve(listener); err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, grpc.ErrServerStopped) {
			l.Error("stores server failed to start serving", "error", err.Error())
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the server
	l.Info("stores grpc server waiting for shutdown signal")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	// On stop signal, gracefully stop the server
	l.Info("stores grpc server received shutdown signal", "signal", sig.String(), "shutdown_timeout", "10s")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		if err := listener.Close(); err != nil {
			l.Error("error closing stores server network listener", "error", err.Error())
		}
		server.GracefulStop()
		cancel()
	}()
	shutdownCtx = logger.WithLogger(shutdownCtx, l)

	err = metricsServer.Shutdown(shutdownCtx)
	if err != nil {
		l.Error("failed to shut down stores metrics server", "error", err.Error())
	}

	if err = sr.Close(shutdownCtx); err != nil {
		l.Error("error closing stores repository", "error", err.Error())
	}

	if err := gc.Close(shutdownCtx); err != nil {
		l.Error("error closing geo client", "error", err.Error())
	}

	<-shutdownCtx.Done()
	l.Info("stores server stopped", "shutdown_reason", shutdownCtx.Err())
}
