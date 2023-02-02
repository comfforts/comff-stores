package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/comfforts/logger"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/comfforts/comff-stores/api/v1"
	"github.com/comfforts/comff-stores/pkg/models"
	"github.com/comfforts/comff-stores/pkg/services/store"
)

type httpServer struct {
	StoreService models.Stores
	logger       logger.AppLogger
}

func newHTTPServer(ss models.Stores, logger logger.AppLogger) *httpServer {
	return &httpServer{
		StoreService: ss,
		logger:       logger,
	}
}

func NewHTTPServer(addr string, logger logger.AppLogger) *http.Server {
	css := store.NewStoreService(logger)
	httpsrv := newHTTPServer(css, logger)
	r := mux.NewRouter()

	r.HandleFunc("/add", httpsrv.AddStore).Methods("POST")

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

type AddStoreRequest struct {
	Store models.JSONMapper `json:"store"`
}
type AddStoreResponse struct {
	Store models.JSONMapper `json:"store"`
	Ok    bool              `json:"ok"`
}

func (s *httpServer) AddStore(w http.ResponseWriter, r *http.Request) {
	var req models.JSONMapper
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.logger.Error("error decoding request", zap.Error(err))
		st := status.New(codes.InvalidArgument, "invalid request")
		http.Error(w, st.Err().Error(), http.StatusInternalServerError)
	}

	stMod, err := models.MapResultToStore(req)
	if err != nil {
		s.logger.Error("error decoding store attributes", zap.Error(err))
		st := status.New(codes.InvalidArgument, "invalid request")
		http.Error(w, st.Err().Error(), http.StatusInternalServerError)
	}

	ctx := r.Context()
	st, err := s.StoreService.AddStore(ctx, stMod)
	if st == nil || err != nil {
		s.logger.Error("store already exists", zap.Error(err))
		st := status.New(codes.AlreadyExists, "store already exists")
		http.Error(w, st.Err().Error(), http.StatusInternalServerError)
	}

	resp := AddStoreResponse{
		Ok:    true,
		Store: models.MapStoreToJSON(st),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		s.logger.Error("error encoding response", zap.Error(err))
		st := status.New(codes.Internal, "error encoding response")
		http.Error(w, st.Err().Error(), http.StatusInternalServerError)
	}
}

type GetStoreRequest struct {
	Id string `json:"id"`
}
type GetStoreResponse struct {
	Store models.JSONMapper `json:"store"`
	Ok    bool              `json:"ok"`
}

func (s *httpServer) GetStore(w http.ResponseWriter, r *http.Request) {
	var req GetStoreRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	ctx := r.Context()

	st, err := s.StoreService.GetStore(ctx, req.Id)
	if err != nil {
		s.logger.Error("store not found", zap.Error(err), zap.String("id", req.Id))
		st := status.New(codes.NotFound, "store not found")
		http.Error(w, st.Err().Error(), http.StatusInternalServerError)
	}

	resp := GetStoreResponse{
		Ok:    true,
		Store: models.MapStoreToJSON(st),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		s.logger.Error("error encoding response", zap.Error(err))
		st := status.New(codes.Internal, "error encoding response")
		http.Error(w, st.Err().Error(), http.StatusInternalServerError)
	}
}

func (s *httpServer) GetStats(ctx context.Context, req *api.StatsRequest) (*api.StatsResponse, error) {
	stats := s.StoreService.GetStoreStats()
	return &api.StatsResponse{
		Count:     uint32(stats.Count),
		HashCount: uint32(stats.HashCount),
		Ready:     stats.Ready,
	}, nil
}
