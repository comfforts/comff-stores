package store

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	api "github.com/comfforts/comff-stores/api/v1"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	fileModels "github.com/comfforts/comff-stores/pkg/models/file"
	storeModels "github.com/comfforts/comff-stores/pkg/models/store"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

var _ raft.FSM = (*fsm)(nil)

type fsm struct {
	DataDir      string
	StoreService storeModels.Stores
	logger       *logging.AppLogger
}

type RequestType uint8

const (
	AddStoreRequestType    RequestType = 0
	StoreUploadRequestType RequestType = 1
)

func (fs *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	reqType := RequestType(buf[0])
	switch reqType {
	case AddStoreRequestType:
		return fs.applyAddStore(buf[1:])
	}
	return nil
}

func (fs *fsm) applyAddStore(b []byte) interface{} {
	var req api.AddStoreRequest
	err := proto.Unmarshal(b, &req)
	if err != nil {
		return err
	}

	ctx := context.Background()
	store, err := fs.StoreService.AddStore(ctx, storeModels.MapStoreRequestToStore(&req))
	if err != nil {
		return err
	}

	return &api.AddStoreResponse{
		Ok:    true,
		Store: storeModels.MapStoreModelToResponse(store),
	}
}

func (fs *fsm) Snapshot() (raft.FSMSnapshot, error) {
	ctx := context.Background()
	r, err := fs.StoreService.Reader(ctx, fs.DataDir)
	if err != nil {
		return nil, err
	}
	return &snapshot{
		reader: r,
		logger: fs.logger,
	}, nil
}

func (fs *fsm) Restore(rf io.ReadCloser) error {
	ctx := context.Background()

	r := bufio.NewReader(rf)
	dec := json.NewDecoder(r)

	// read open bracket
	t, err := dec.Token()
	if err != nil || t != json.Delim('[') {
		fs.logger.Error("start token error", zap.Error(err))
		return fileUtils.ErrStartToken
	}

	// while the array contains values
	for dec.More() {
		var result fileModels.JSONMapper
		err := dec.Decode(&result)
		if err != nil {
			fs.logger.Error(fileUtils.ERROR_DECODING_RESULT, zap.Error(err))
		} else {
			store, err := storeModels.MapResultToStore(result)
			if err != nil {
				fs.logger.Error("error processing store data", zap.Error(err), zap.Any("storeJson", r))
			}

			addedStr, err := fs.StoreService.AddStore(ctx, store)
			if addedStr == nil || err != nil {
				fs.logger.Error("error processing store data", zap.Error(err), zap.Any("store", store))
			}
		}
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil || t != json.Delim(']') {
		fs.logger.Error("end token error", zap.Error(err))
		return fileUtils.ErrEndToken
	}
	return nil
}
