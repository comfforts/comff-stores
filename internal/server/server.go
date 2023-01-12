package server

import (
	"context"
	"fmt"
	"time"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/pkg/logging"
	fileModels "github.com/comfforts/comff-stores/pkg/models/file"
	geoModels "github.com/comfforts/comff-stores/pkg/models/geo"
	storeModels "github.com/comfforts/comff-stores/pkg/models/store"
)

var _ api.StoresServer = (*grpcServer)(nil)

const (
	objectWildcard = "*"
	addAction      = "add"
	getAction      = "get"
	statsAction    = "stats"
	searchAction   = "search"
	uploadAction   = "upload"
	locateAction   = "locate"
	serversAction  = "servers"
)

type subjectContextKey struct{}

func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}

type Authorizer interface {
	Authorize(subject, object, action string) error
}

type Servicer interface {
	GetServers(context.Context) ([]*api.Server, error)
}

type Config struct {
	StoreService storeModels.Stores
	GeoService   geoModels.GeoCoder
	StoreLoader  fileModels.Loader
	Authorizer   Authorizer
	Servicer     Servicer
	Logger       *logging.AppLogger
}

type grpcServer struct {
	api.StoresServer
	*Config
}

func newGrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}
	return srv, nil
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	logger := zap.L().Named("comff-stores")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(
			func(duration time.Duration) zapcore.Field {
				return zap.Int64(
					"grpc.time_ns",
					duration.Nanoseconds(),
				)
			},
		),
	}

	// halfSampler := trace.ProbabilitySampler(0.5)
	// trace.ApplyConfig(trace.Config{
	// 	DefaultSampler: func(p trace.Sampler) trace.Sampler {
	// 		if strings.Contains(p.Name, "AddStore") {
	// 			return trace.SamplingParameters{Sample: true}
	// 		}
	// 		return halfSampler(p)
	// 	},
	// })

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, err
	}

	opts = append(opts,
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
				grpc_auth.UnaryServerInterceptor(authenticate),
			),
		),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

	// opts = append(opts,
	// 	grpc.UnaryInterceptor(
	// 		grpc_middleware.ChainUnaryServer(
	// 			grpc_ctxtags.UnaryServerInterceptor(),
	// 			grpc_auth.UnaryServerInterceptor(authenticate),
	// 		),
	// 	),
	// )

	config.Logger.Info("creating gRPC server instance")
	gsrv := grpc.NewServer(opts...)

	config.Logger.Info("creating stores gRPC server instance")
	srv, err := newGrpcServer(config)
	if err != nil {
		return nil, err
	}

	config.Logger.Info("registring stores gRPC server instance")
	api.RegisterStoresServer(gsrv, srv)

	config.Logger.Info("enabling reflection for stores api")
	reflection.Register(gsrv)

	config.Logger.Info("creating health gRPC server instance")
	hserv := health.NewServer()
	hserv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	config.Logger.Info("registring health gRPC server instance")
	healthpb.RegisterHealthServer(gsrv, hserv)

	return gsrv, nil
}

func (s *grpcServer) AddStore(ctx context.Context, req *api.AddStoreRequest) (*api.AddStoreResponse, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		addAction,
	); err != nil {
		return nil, err
	}

	store, err := s.StoreService.AddStore(ctx, &storeModels.Store{
		Name:      req.Name,
		Org:       req.Org,
		Longitude: float64(req.Longitude),
		Latitude:  float64(req.Latitude),
		City:      req.City,
		Country:   req.Country,
		StoreId:   req.StoreId,
	})
	if store == nil || err != nil {
		s.Logger.Error("store already exists", zap.Error(err))
		st := status.New(codes.AlreadyExists, "store already exists")
		return &api.AddStoreResponse{
			Ok:    false,
			Store: nil,
		}, st.Err()
	}

	return &api.AddStoreResponse{
		Ok:    true,
		Store: storeModels.MapStoreModelToResponse(store),
	}, nil
}

func (s *grpcServer) GetStore(ctx context.Context, req *api.GetStoreRequest) (*api.GetStoreResponse, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		getAction,
	); err != nil {
		return nil, err
	}

	store, err := s.StoreService.GetStore(ctx, req.Id)
	if err != nil {
		s.Logger.Error("store not found", zap.Error(err), zap.String("id", req.Id))
		st := status.New(codes.NotFound, "store not found")
		return nil, st.Err()
	}
	return &api.GetStoreResponse{
		Store: storeModels.MapStoreModelToResponse(store),
	}, nil
}

func (s *grpcServer) GetStats(ctx context.Context, req *api.StatsRequest) (*api.StatsResponse, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		getAction,
	); err != nil {
		return nil, err
	}

	stats := s.StoreService.GetStoreStats()
	return &api.StatsResponse{
		Count:     uint32(stats.Count),
		HashCount: uint32(stats.HashCount),
		Ready:     stats.Ready,
	}, nil
}

func (s *grpcServer) SearchStore(ctx context.Context, req *api.SearchStoreRequest) (*api.SearchStoreResponse, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		getAction,
	); err != nil {
		return nil, err
	}

	if (req.Latitude == 0 || req.Longitude == 0) && req.PostalCode == "" {
		st := status.New(codes.NotFound, "missing required search params")
		s.Logger.Error("missing required search params")
		return nil, st.Err()
	}

	if (req.Latitude == 0 || req.Longitude == 0) && req.PostalCode != "" {
		point, err := s.GeoService.Geocode(ctx, req.PostalCode, "US")
		if err != nil {
			s.Logger.Error("error geocoding postalcode", zap.Error(err), zap.String("postalcode", req.PostalCode))
			st := status.New(codes.NotFound, "error geocoding postalcode")
			return nil, st.Err()
		}
		req.Latitude = float32(point.Latitude)
		req.Longitude = float32(point.Longitude)
	}

	stores, err := s.StoreService.GetStoresForGeoPoint(ctx, float64(req.Latitude), float64(req.Longitude), int(req.Distance))
	if err != nil {
		s.Logger.Error("error getting stores", zap.Error(err), zap.Float32("latitude", req.Latitude), zap.Float32("longitude", req.Longitude))
		st := status.New(codes.NotFound, "no store found")
		return nil, st.Err()
	}

	return &api.SearchStoreResponse{
		Stores: storeModels.MapStoreListToResponse(stores),
		Geo: &api.Point{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
	}, nil
}

func (s *grpcServer) GeoLocate(ctx context.Context, req *api.GeoLocationRequest) (*api.GeoLocationResponse, error) {
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		getAction,
	); err != nil {
		return nil, err
	}

	point, err := s.GeoService.Geocode(ctx, req.PostalCode, "US")
	if err != nil {
		s.Logger.Error("error geocoding postalcode", zap.Error(err), zap.String("postalcode", req.PostalCode))
		st := status.New(codes.NotFound, "error geocoding postalcode")
		return nil, st.Err()
	}

	return &api.GeoLocationResponse{
		Point: &api.Point{
			Latitude:  float32(point.Latitude),
			Longitude: float32(point.Longitude),
		},
	}, nil
}

func (s *grpcServer) StoreUpload(ctx context.Context, req *api.StoreUploadRequest) (*api.StoreUploadResponse, error) {
	fPath := req.FileName
	if req.Path != "" {
		fPath = fmt.Sprintf("%s/%s", req.Path, req.FileName)
	}

	err := s.StoreLoader.ProcessFile(ctx, req.FileName)
	if err != nil {
		s.Logger.Error("error completing store upload request", zap.Error(err), zap.String("file", fPath))
		st := status.New(codes.NotFound, "error completing store upload request")
		return nil, st.Err()
	}

	return &api.StoreUploadResponse{
		Ok: true,
	}, nil
}

func (s *grpcServer) GetServers(ctx context.Context, req *api.GetServersRequest) (*api.GetServersResponse, error) {
	servers, err := s.Servicer.GetServers(ctx)
	if err != nil {
		s.Logger.Error("error completing get servers request", zap.Error(err))
		st := status.New(codes.NotFound, "error completing get servers request")
		return nil, st.Err()
	}

	return &api.GetServersResponse{
		Servers: servers,
	}, nil
}

func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
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
