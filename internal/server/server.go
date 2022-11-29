package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/comfforts/comff-stores/api/v1"
	"github.com/comfforts/comff-stores/pkg/services/store"
)

var _ api.StoresServer = (*grpcServer)(nil)

type Config struct {
	StoreService *store.StoreService
}

type grpcServer struct {
	api.UnimplementedStoresServer
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
	api.RegisterStoresServer(gsrv, srv)
	return gsrv, nil
}

func (s *grpcServer) AddStore(ctx context.Context, req *api.AddStoreRequest) (*api.AddStoreResponse, error) {
	ok, err := s.StoreService.AddStore(ctx, &store.Store{
		Id:        req.StoreId,
		Name:      req.Name,
		Longitude: float64(req.Longitude),
		Latitude:  float64(req.Latitude),
		City:      req.City,
		Country:   req.Country,
	})
	if !ok || err != nil {
		st := status.New(codes.AlreadyExists, "store already exists")
		return nil, st.Err()
	}
	return &api.AddStoreResponse{
		Ok: ok,
	}, nil
}

func (s *grpcServer) GetStore(ctx context.Context, req *api.GetStoreRequest) (*api.GetStoreResponse, error) {
	store, err := s.StoreService.GetStore(ctx, req.StoreId)
	if err != nil {
		st := status.New(codes.NotFound, "store not found")
		return nil, st.Err()
	}
	return &api.GetStoreResponse{
		Store: &api.Store{
			Id:        store.Id,
			Name:      store.Name,
			Longitude: float32(store.Longitude),
			Latitude:  float32(store.Latitude),
			City:      store.City,
			Country:   store.Country,
		},
	}, nil
}
