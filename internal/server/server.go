package server

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	api "github.com/comfforts/comff-stores/api/v1"

	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/geocode"
	"github.com/comfforts/comff-stores/pkg/services/store"
)

var _ api.StoresServer = (*grpcServer)(nil)

type Config struct {
	StoreService *store.StoreService
	GeoService   *geocode.GeoCodeService
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

func NewGrpcServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	gsrv := grpc.NewServer(opts...)
	srv, err := newGrpcServer(config)
	if err != nil {
		return nil, err
	}

	// register grpc server
	api.RegisterStoresServer(gsrv, srv)

	// enable reflection
	reflection.Register(gsrv)

	return gsrv, nil
}

func (s *grpcServer) AddStore(ctx context.Context, req *api.AddStoreRequest) (*api.AddStoreResponse, error) {
	store, err := s.StoreService.AddStore(ctx, &store.Store{
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
		Store: MapStoreModelToResponse(store),
	}, nil
}

func (s *grpcServer) GetStore(ctx context.Context, req *api.GetStoreRequest) (*api.GetStoreResponse, error) {
	store, err := s.StoreService.GetStore(ctx, req.Id)
	if err != nil {
		s.Logger.Error("store not found", zap.Error(err), zap.String("id", req.Id))
		st := status.New(codes.NotFound, "store not found")
		return nil, st.Err()
	}
	return &api.GetStoreResponse{
		Store: MapStoreModelToResponse(store),
	}, nil
}

func (s *grpcServer) GetStats(ctx context.Context, req *api.StatsRequest) (*api.StatsResponse, error) {
	stats := s.StoreService.GetStoreStats()
	return &api.StatsResponse{
		Count:     uint32(stats.Count),
		HashCount: uint32(stats.HashCount),
		Ready:     stats.Ready,
	}, nil
}

func (s *grpcServer) SearchStore(ctx context.Context, req *api.SearchStoreRequest) (*api.SearchStoreResponse, error) {
	if (req.Latitude == 0 || req.Longitude == 0) && req.PostalCode == "" {
		st := status.New(codes.NotFound, "missing required search params")
		fmt.Println()
		fmt.Println(req)
		fmt.Println()
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
		Stores: MapStoreListToResponse(stores),
		Geo: &api.Point{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
	}, nil
}

func (s *grpcServer) GeoLocate(ctx context.Context, req *api.GeoLocationRequest) (*api.GeoLocationResponse, error) {
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
