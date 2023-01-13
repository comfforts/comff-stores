package store

import (
	"io"

	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

var _ raft.FSMSnapshot = (*snapshot)(nil)

type snapshot struct {
	reader io.Reader
	logger *logging.AppLogger
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		_ = sink.Cancel()
		s.logger.Error("error persisting snapshot", zap.Error(err))
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {
	s.logger.Info("snapshot done, releasing")
}
