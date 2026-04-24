package grpchandler

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	config "github.com/comfforts/comff-config"
	"github.com/comfforts/logger"

	api "github.com/comfforts/comff-stores/api/stores/v1"
	stdom "github.com/comfforts/comff-stores/internal/domain/stores"
	"github.com/comfforts/comff-stores/internal/infra/observability"
)

var _ api.StoresServer = (*grpcServer)(nil)

const (
	objectWildcard     = "*"
	addStoreAction     = "add-store"
	getStoreAction     = "get-store"
	updateStoreAction  = "update-store"
	deleteStoreAction  = "delete-store"
	searchStoresAction = "search-stores"
)

const (
	ERR_UNAUTHORIZED_ADD_STORE     = "unauthorized to add store"
	ERR_UNAUTHORIZED_GET_STORE     = "unauthorized to get store"
	ERR_UNAUTHORIZED_UPDATE_STORE  = "unauthorized to update store"
	ERR_UNAUTHORIZED_DELETE_STORE  = "unauthorized to delete store"
	ERR_UNAUTHORIZED_SEARCH_STORES = "unauthorized to search stores"
)

type subjectContextKey struct{}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}

// Authorizer interface checks if the subject is "authorized-user" of API requested.
type Authorizer interface {
	Authorize(subject, object, action string) error
}

type Config struct {
	Authorizer Authorizer
	stdom.StoresService
}

func BuildServerConfig(ctx context.Context, ss stdom.StoresService) (*Config, error) {
	// Initialize the authorizer for the geo service
	authorizer, err := config.SetupAuthorizer()
	if err != nil {
		return nil, err
	}

	servCfg := &Config{
		StoresService: ss,
		Authorizer:    authorizer,
	}
	return servCfg, nil
}

type grpcServer struct {
	*Config
	nodeName string
	metrics  observability.Metrics
	api.StoresServer
}

// newGrpcServer initializes a new grpcServer instance with the provided Config.
func newGrpcServer(config *Config) (srv *grpcServer, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error getting host name: %w", err)
	}
	metrics, err := observability.NewMetrics()
	if err != nil {
		return nil, fmt.Errorf("error initializing metrics: %w", err)
	}

	srv = &grpcServer{
		Config:   config,
		nodeName: hostname,
		metrics:  metrics,
	}
	return srv, nil
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	srv, err := newGrpcServer(config)
	if err != nil {
		return nil, err
	}

	opts = append(opts,
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_auth.StreamServerInterceptor(authenticate),
				grpc_auth.StreamServerInterceptor(decorateContext),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_auth.UnaryServerInterceptor(authenticate),
				grpc_auth.UnaryServerInterceptor(decorateContext),
				grpc_auth.UnaryServerInterceptor(metadataLogger),
				UnaryLoggingInterceptor(),
				UnaryMetricsInterceptor(srv.metrics),
			),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	gsrv := grpc.NewServer(opts...)

	api.RegisterStoresServer(gsrv, srv)

	reflection.Register(gsrv)

	hserv := health.NewServer()
	hserv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(gsrv, hserv)

	return gsrv, nil
}

func (s *grpcServer) AddStore(ctx context.Context, req *api.AddStoreRequest) (*api.AddStoreResponse, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	// Authorization check
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		addStoreAction,
	); err != nil {
		st := status.New(codes.Unauthenticated, ERR_UNAUTHORIZED_ADD_STORE)
		return nil, st.Err()
	}

	// Validate request
	if req == nil {
		l.Error("AddStore called with nil request")
		st := status.New(codes.InvalidArgument, "request cannot be nil")
		return nil, st.Err()
	}
	params := stdom.MapToAddStoreParams(req)

	storeID, err := s.StoresService.AddStore(ctx, params)
	if err != nil {
		l.Error("error adding store", "error", err.Error())
		st := status.New(codes.Internal, "error adding store")
		return nil, st.Err()
	}

	return &api.AddStoreResponse{
		Ok: true,
		Id: &storeID,
	}, nil
}

func (s *grpcServer) GetStore(ctx context.Context, req *api.GetStoreRequest) (*api.GetStoreResponse, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	// Authorization check
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		getStoreAction,
	); err != nil {
		st := status.New(codes.Unauthenticated, ERR_UNAUTHORIZED_GET_STORE)
		return nil, st.Err()
	}

	if req == nil || req.GetId() == "" {
		l.Error("GetStore called with invalid request: missing store ID")
		st := status.New(codes.InvalidArgument, "store ID is required")
		return nil, st.Err()
	}

	store, err := s.StoresService.GetStore(ctx, req.GetId())
	if err != nil {
		l.Error("error getting store", "error", err.Error(), "store_id", req.GetId())
		st := status.New(codes.Internal, "error getting store")
		return nil, st.Err()
	}

	return &api.GetStoreResponse{
		Store: stdom.MapToStoreProto(store),
	}, nil
}

func (s *grpcServer) UpdateStore(ctx context.Context, req *api.UpdateStoreRequest) (*api.UpdateStoreResponse, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	// Authorization check
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		updateStoreAction,
	); err != nil {
		st := status.New(codes.Unauthenticated, ERR_UNAUTHORIZED_UPDATE_STORE)
		return nil, st.Err()
	}

	if req == nil || req.GetId() == "" {
		l.Error("UpdateStore called with invalid request: missing store ID")
		st := status.New(codes.InvalidArgument, "store ID is required")
		return nil, st.Err()
	}

	params := stdom.MapToUpdateStoreParams(req)

	err = s.StoresService.UpdateStore(ctx, req.GetId(), params)
	if err != nil {
		l.Error("error updating store", "error", err.Error(), "store_id", req.GetId())
		st := status.New(codes.Internal, "error updating store")
		return nil, st.Err()
	}

	return &api.UpdateStoreResponse{
		Ok: true,
	}, nil
}

func (s *grpcServer) DeleteStore(ctx context.Context, req *api.DeleteStoreRequest) (*api.DeleteStoreResponse, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	// Authorization check
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		deleteStoreAction,
	); err != nil {
		st := status.New(codes.Unauthenticated, ERR_UNAUTHORIZED_DELETE_STORE)
		return nil, st.Err()
	}

	if req == nil || req.GetId() == "" {
		l.Error("DeleteStore called with invalid request: missing store ID")
		st := status.New(codes.InvalidArgument, "store ID is required")
		return nil, st.Err()
	}

	err = s.StoresService.DeleteStore(ctx, req.GetId())
	if err != nil {
		l.Error("error deleting store", "error", err.Error(), "store_id", req.GetId())
		st := status.New(codes.Internal, "error deleting store")
		return nil, st.Err()
	}

	return &api.DeleteStoreResponse{
		Ok: true,
	}, nil
}

func (s *grpcServer) SearchStore(ctx context.Context, req *api.SearchStoreRequest) (*api.SearchStoreResponse, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	// Authorization check
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		searchStoresAction,
	); err != nil {
		st := status.New(codes.Unauthenticated, ERR_UNAUTHORIZED_SEARCH_STORES)
		return nil, st.Err()
	}

	if req == nil {
		l.Error("SearchStores called with nil request")
		st := status.New(codes.InvalidArgument, "request cannot be nil")
		return nil, st.Err()
	}
	params := stdom.MapToSearchStoreParams(req)

	stores, err := s.StoresService.SearchStores(ctx, params)
	if err != nil {
		l.Error("error searching stores", "error", err.Error())
		st := status.New(codes.Internal, "error searching stores")
		return nil, st.Err()
	}

	var storeGeoProtos []*api.StoreGeo
	for _, st := range stores {
		storeGeoProtos = append(storeGeoProtos, &api.StoreGeo{
			Store: stdom.MapToStoreProto(st),
		})
	}

	return &api.SearchStoreResponse{
		Stores: storeGeoProtos,
	}, nil
}

func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.PermissionDenied,
			"couldn't find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}

func decorateContext(ctx context.Context) (context.Context, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		// If there's no logger in the context, add a new one
		l = logger.GetSlogLogger()
	}

	attrs := []any{"handler", "geo-grpc"}
	p, ok := peer.FromContext(ctx)
	if ok {
		attrs = append(attrs, "peer", p.Addr.String())
	}
	if subj, ok := ctx.Value(subjectContextKey{}).(string); ok && subj != "" {
		attrs = append(attrs, "subject", subj)
	}
	l = logger.WithAttrs(l, attrs...)
	ctx = logger.WithLogger(ctx, l)
	ctx = logger.WithTraceAttrs(ctx)
	return ctx, nil
}

func metadataLogger(ctx context.Context) (context.Context, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		// If there's no logger in the context, add a new one
		l = logger.GetSlogLogger()
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		l.Warn("missing metadata")
		return ctx, nil
	}
	if requestID := firstMetadataValue(md, "x-request-id"); requestID != "" {
		ctx = logger.WithContextAttrs(ctx, "request_id", requestID)
	}
	if correlationID := firstMetadataValue(md, "x-correlation-id"); correlationID != "" {
		ctx = logger.WithContextAttrs(ctx, "correlation_id", correlationID)
	}
	if userAgent := firstMetadataValue(md, "user-agent"); userAgent != "" {
		ctx = logger.WithContextAttrs(ctx, "user_agent", userAgent)
	}
	return ctx, nil
}

func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		service, method := splitFullMethod(info.FullMethod)
		ctx = logger.WithContextAttrs(
			ctx,
			"rpc.system", "grpc",
			"rpc.service", service,
			"rpc.method", method,
			"rpc.full_method", info.FullMethod,
		)
		ctx = logger.WithTraceAttrs(ctx)

		l, logErr := logger.LoggerFromContext(ctx)
		if logErr != nil {
			l = logger.GetSlogLogger()
		}

		start := time.Now()
		resp, err = handler(ctx, req)

		args := []any{
			"rpc.status", status.Code(err).String(),
			"rpc.duration_ms", time.Since(start).Milliseconds(),
		}
		if err != nil {
			l.Error("grpc request failed", append(args, "error", err.Error())...)
		} else {
			l.Info("grpc request completed", args...)
		}
		return resp, err
	}
}

func UnaryMetricsInterceptor(m observability.Metrics) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()
		m.AddInflightRequest(ctx, info.FullMethod, 1)
		defer m.AddInflightRequest(ctx, info.FullMethod, -1)
		resp, err = handler(ctx, req)

		statusCode := status.Code(err).String()
		m.IncRequest(ctx, info.FullMethod, statusCode)
		m.ObserveRequestDuration(ctx, info.FullMethod, statusCode, time.Since(start))

		return resp, err
	}
}

func splitFullMethod(fullMethod string) (string, string) {
	service := path.Dir(fullMethod)
	method := path.Base(fullMethod)
	return strings.TrimPrefix(service, "/"), method
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
