package log

import "github.com/comfforts/comff-stores/pkg/logging"

type Config struct {
	Segment struct {
		MaxIndexSize  uint64
		InitialOffset uint64
	}
	logger *logging.AppLogger
}
