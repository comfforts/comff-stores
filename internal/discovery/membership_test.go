package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/require"
)

const TEST_DIR = "test-data"

func TestMembership(t *testing.T) {
	h := &handler{}

	m := setupMember(t, nil, h)
	mems := m[0].Members()
	require.Equal(t, len(mems), 1)
	for _, v := range mems {
		t.Logf("  Member node - Name: %s, Status: %s, Port: %d, addr: %s", v.Name, v.Status, v.Port, v.Addr.String())
		require.Equal(t, serf.StatusAlive, v.Status)
	}

	m = setupMember(t, m, h)
	mems = m[0].Members()
	require.Equal(t, len(mems), 2)
	for _, v := range mems {
		t.Logf("  Member node - Name: %s, Status: %s, Port: %d, addr: %s", v.Name, v.Status, v.Port, v.Addr.String())
		require.Equal(t, serf.StatusAlive, v.Status)
	}

	m = setupMember(t, m, h)
	mems = m[0].Members()
	require.Equal(t, len(mems), 3)
	for _, v := range mems {
		t.Logf("  Member node - Name: %s, Status: %s, Port: %d, addr: %s", v.Name, v.Status, v.Port, v.Addr.String())
		require.Equal(t, serf.StatusAlive, v.Status)
	}

	t.Logf("  Removing node - Name: %s, addr: %s", m[2].NodeName, m[2].BindAddr)
	err := m[2].Leave()
	require.NoError(t, err)
	mems = m[0].Members()
	for _, v := range mems {
		t.Logf("  Member node - Name: %s, Status: %s, Port: %d, addr: %s", v.Name, v.Status, v.Port, v.Addr.String())
		if v.Name == "2" {
			require.Equal(t, serf.StatusLeft, v.Status)
		} else {
			require.Equal(t, serf.StatusAlive, v.Status)
		}
	}

	t.Logf("  Removing node - Name: %s, addr: %s", m[1].NodeName, m[1].BindAddr)
	err = m[1].Leave()
	require.NoError(t, err)
	mems = m[0].Members()
	for _, v := range mems {
		t.Logf("  Member node - Name: %s, Status: %s, Port: %d, addr: %s", v.Name, v.Status, v.Port, v.Addr.String())
		if v.Name == "0" {
			require.Equal(t, serf.StatusAlive, v.Status)
		} else {
			require.Equal(t, serf.StatusLeft, v.Status)
		}
	}

	t.Logf("  Removing node - Name: %s, addr: %s", m[0].NodeName, m[0].BindAddr)
	err = m[0].Leave()
	require.NoError(t, err)
	mems = m[0].Members()
	for _, v := range mems {
		t.Logf("  Member node - Name: %s, Status: %s, Port: %d, addr: %s", v.Name, v.Status, v.Port, v.Addr.String())
		require.Equal(t, serf.StatusLeft, v.Status)
	}
	err = os.RemoveAll(TEST_DIR)
	require.NoError(t, err)
}

func setupMember(t *testing.T, members []*Membership, h *handler) []*Membership {
	id := len(members)
	port := getPort(id)
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", port)
	tags := map[string]string{
		"rpc_addr": addr,
	}
	c := Config{
		NodeName: fmt.Sprintf("%d", id),
		BindAddr: addr,
		Tags:     tags,
	}

	dataDir := filepath.Join(TEST_DIR, fmt.Sprintf("node-%d/", len(members)))

	c.Logger = logging.NewTestAppLogger(dataDir)

	if len(members) == 0 {
		h.joins = make(chan map[string]string, 3)
		h.leaves = make(chan string, 3)
	} else {
		c.PeerNodeAddrs = []string{
			members[0].BindAddr,
		}
	}
	t.Logf("  Adding node - Name: %s, addr: %s", c.NodeName, c.BindAddr)
	m, err := NewMembership(h, c)
	require.NoError(t, err)
	members = append(members, m)
	return members
}

func getPort(id int) int {
	return 15001 + id
}

type handler struct {
	joins  chan map[string]string
	leaves chan string
}

func (h *handler) Join(id, addr string) error {
	if h.joins != nil {
		h.joins <- map[string]string{
			"id":   id,
			"addr": addr,
		}
	}
	return nil
}

func (h *handler) Leave(id string) error {
	if h.leaves != nil {
		h.leaves <- id
	}
	return nil
}
