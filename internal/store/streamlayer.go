package store

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/comfforts/logger"

	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

var _ raft.StreamLayer = (*StreamLayer)(nil)

const RaftRPC = 1

type StreamLayer struct {
	ln              net.Listener
	serverTLSConfig *tls.Config
	peerTLSConfig   *tls.Config
	logger          logger.AppLogger
}

func NewStreamLayer(
	ln net.Listener,
	serverTLSConfig,
	peerTLSConfig *tls.Config,
	logger logger.AppLogger,
) *StreamLayer {
	return &StreamLayer{
		ln:              ln,
		serverTLSConfig: serverTLSConfig,
		peerTLSConfig:   peerTLSConfig,
		logger:          logger,
	}
}

func (s *StreamLayer) Dial(
	addr raft.ServerAddress,
	timeout time.Duration,
) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	var conn, err = dialer.Dial("tcp", string(addr))
	if err != nil {
		s.logger.Error("error dialing TCP", zap.Error(err), zap.String("addr", string(addr)))
		return nil, err
	}
	// identify to mux this is a raft rpc
	_, err = conn.Write([]byte{byte(RaftRPC)})
	if err != nil {
		s.logger.Error("error writing raft RPC signal", zap.Error(err))
		return nil, err
	}
	if s.peerTLSConfig != nil {
		conn = tls.Client(conn, s.peerTLSConfig)
	}
	return conn, err
}

func (s *StreamLayer) Accept() (net.Conn, error) {
	conn, err := s.ln.Accept()
	if err != nil {
		s.logger.Error("error accepting connection", zap.Error(err))
		return nil, err
	}
	b := make([]byte, 1)
	_, err = conn.Read(b)
	if err != nil {
		s.logger.Error("error reading connection", zap.Error(err))
		return nil, err
	}
	if bytes.Compare([]byte{byte(RaftRPC)}, b) != 0 {
		s.logger.Error("not a raft RPC")
		return nil, fmt.Errorf("not a raft rpc")
	}
	if s.serverTLSConfig != nil {
		return tls.Server(conn, s.serverTLSConfig), nil
	}
	return conn, nil
}

func (s *StreamLayer) Close() error {
	err := s.ln.Close()
	if err != nil {
		s.logger.Error("error closing connection", zap.Error(err))
	}
	return err
}

func (s *StreamLayer) Addr() net.Addr {
	return s.ln.Addr()
}
