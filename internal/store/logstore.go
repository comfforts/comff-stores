package store

import (
	"github.com/comfforts/errors"
	"github.com/comfforts/logger"
	"github.com/comfforts/recorder"

	api "github.com/comfforts/recorder/api/v1"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

var _ raft.LogStore = (*logStore)(nil)

const (
	ERROR_CREATING_LOGSTORE     string = "error creating raft log store"
	ERROR_GETTING_LOGSTORE_LOGS string = "error getting raft logstore logs"
	ERROR_STORING_LOGSTORE_LOGS string = "error storing raft logstore logs"
)

type logStore struct {
	recorder.Recorder
	logger logger.AppLogger
}

func newLogStore(dir string, c Config) (*logStore, error) {
	log, err := recorder.NewRecorder(dir, c.Config)
	if err != nil {
		c.Logger.Error(ERROR_CREATING_LOGSTORE, zap.Error(err))
		return nil, errors.WrapError(err, ERROR_CREATING_LOGSTORE)
	}
	return &logStore{log, c.Logger}, nil
}

func (l *logStore) FirstIndex() (uint64, error) {
	lowest, err := l.LowestOffset()
	l.logger.Debug("lowest offset", zap.Uint64("lowest", lowest), zap.Error(err))
	return lowest, err
}

func (l *logStore) LastIndex() (uint64, error) {
	off, err := l.HighestOffset()
	l.logger.Debug("highest offset", zap.Uint64("highest", off), zap.Error(err))
	return off, err
}

func (l *logStore) GetLog(index uint64, out *raft.Log) error {
	in, err := l.Read(index)
	if err != nil {
		l.logger.Error(ERROR_GETTING_LOGSTORE_LOGS, zap.Error(err))
		return errors.WrapError(err, ERROR_GETTING_LOGSTORE_LOGS)
	}
	out.Data = in.Value
	out.Index = in.Offset
	out.Type = raft.LogType(in.Type)
	out.Term = in.Term
	l.logger.Debug("got log record", zap.String("value", string(in.Value)), zap.Uint64("offset", in.Offset))
	return nil
}

func (l *logStore) StoreLog(record *raft.Log) error {
	l.logger.Debug("storing log record", zap.String("value", string(record.Data)))
	return l.StoreLogs([]*raft.Log{record})
}

func (l *logStore) StoreLogs(records []*raft.Log) error {
	l.logger.Debug("storing logs", zap.Int("len", len(records)))
	for _, record := range records {
		if _, err := l.Append(&api.Record{
			Value: record.Data,
			Term:  record.Term,
			Type:  uint32(record.Type),
		}); err != nil {
			l.logger.Error(ERROR_STORING_LOGSTORE_LOGS, zap.Error(err))
			return errors.WrapError(err, ERROR_STORING_LOGSTORE_LOGS)
		}
	}
	return nil
}

func (l *logStore) DeleteRange(min, max uint64) error {
	return l.Truncate(max)
}
