package store

import (
	"io"

	"github.com/comfforts/logger"

	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

var _ raft.FSMSnapshot = (*snapshot)(nil)

type snapshot struct {
	reader io.Reader
	logger logger.AppLogger
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		s.logger.Error("error persisting snapshot", zap.Error(err))
		erro := sink.Cancel()
		s.logger.Error("error canceling snapshot sink", zap.Error(erro))
		return err
	}
	err := sink.Close()
	if err != nil {
		s.logger.Error("error closing snapshot sink", zap.Error(err))
		return err
	}
	return nil
}

func (s *snapshot) Release() {
	s.logger.Info("snapshot done, releasing")
}
