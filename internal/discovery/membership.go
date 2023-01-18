package discovery

import (
	"net"

	"github.com/comfforts/logger"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

type Config struct {
	NodeName      string
	BindAddr      string
	Tags          map[string]string
	PeerNodeAddrs []string
	Logger        logger.AppLogger
}

type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf
	events  chan serf.Event
}

func NewMembership(handler Handler, config Config) (*Membership, error) {
	c := &Membership{
		Config:  config,
		handler: handler,
	}
	if err := c.setupSerf(); err != nil {
		config.Logger.Error("error setting up serf", zap.Error(err))
		return nil, err
	}
	return c, nil
}

func (m *Membership) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

func (m *Membership) Leave() error {
	return m.serf.Leave()
}

func (m *Membership) setupSerf() (err error) {
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)
	if err != nil {
		m.Config.Logger.Error("error setting up serf", zap.Error(err))
		return err
	}
	config := serf.DefaultConfig()
	config.Init()
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port
	m.events = make(chan serf.Event)
	config.EventCh = m.events
	config.Tags = m.Tags
	config.NodeName = m.Config.NodeName
	m.serf, err = serf.Create(config)
	if err != nil {
		m.Config.Logger.Error("error setting up serf", zap.Error(err))
		return err
	}
	go m.eventHandler()

	m.Config.Logger.Info("setupSerf()", zap.String("name", m.Config.NodeName), zap.Any("peer_node_addrs", m.PeerNodeAddrs))
	if m.PeerNodeAddrs != nil {
		_, err = m.serf.Join(m.PeerNodeAddrs, true)
		if err != nil {
			m.Config.Logger.Error("error setting up serf", zap.Error(err))
			return err
		}
	}
	return nil
}

func (m *Membership) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				m.Config.Logger.Info("serf.EventMemberJoin event", zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberLeave:
			for _, member := range e.(serf.MemberEvent).Members {
				m.Config.Logger.Info("serf.EventMemberLeave event", zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
				if m.isLocal(member) {
					return
				}
				m.handleLeave(member)
			}
		case serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				m.Config.Logger.Info("serf.EventMemberFailed event", zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
				if m.isLocal(member) {
					return
				}
				m.handleLeave(member)
			}
		}
	}
}

func (m *Membership) handleJoin(member serf.Member) {
	if err := m.handler.Join(
		member.Name,
		member.Tags["rpc_addr"],
	); err != nil {
		if err == raft.ErrNotLeader {
			m.Config.Logger.Debug("error joining serf", zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
		} else {
			m.Config.Logger.Error("error joining serf", zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
		}
	}
	m.Config.Logger.Info("member joined serf", zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
}

func (m *Membership) handleLeave(member serf.Member) {
	if err := m.handler.Leave(
		member.Name,
	); err != nil {
		if err == raft.ErrNotLeader {
			m.Config.Logger.Debug("error leaving serf", zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
		} else {
			m.Config.Logger.Error("error leaving serf", zap.Error(err), zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
		}
	}
	m.Config.Logger.Info("member left serf", zap.String("name", member.Name), zap.String("rpc_addr", member.Tags["rpc_addr"]))
}
