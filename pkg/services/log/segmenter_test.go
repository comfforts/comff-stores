package log

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/comfforts/comff-stores/api/v1"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

func TestSegmenter(t *testing.T) {
	dir := fmt.Sprintf("%s/", TEST_DATA_DIR)
	err := fileUtils.CreateDirectory(dir)
	require.NoError(t, err)

	defer func() {
		err = os.RemoveAll(TEST_DATA_DIR)
		require.NoError(t, err)
	}()

	want := &api.Record{Value: []byte("hello world")}

	c := Config{}
	c.Segment.MaxIndexSize = 5

	s, err := newSegmenter(dir, 16, c)
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset)
	require.False(t, s.IsMaxed())

	for i := uint64(0); i < 5; i++ {
		off, err := s.Append(want)
		require.NoError(t, err)
		require.Equal(t, 16+i, off)

		got, err := s.Read(off)
		require.NoError(t, err)

		wantVal := string(want.Value)
		gotVal := string(got.Value)
		t.Logf("want: %s, got %s", wantVal, gotVal)
		require.Equal(t, want.Value, got.Value)
	}

	maxed := s.IsMaxed()
	t.Log("is maxed: ", maxed)

	_, err = s.Append(want)
	require.Equal(t, io.EOF, err)

	// maxed index
	require.True(t, s.IsMaxed())

	err = s.Close()
	require.NoError(t, err)

	c.Segment.MaxIndexSize = 3

	s, err = newSegmenter(dir, 16, c)
	require.NoError(t, err)
	// maxed store
	require.True(t, s.IsMaxed())

	err = s.Close()
	require.NoError(t, err)

	c.Segment.MaxIndexSize = 6

	s, err = newSegmenter(dir, 16, c)
	require.NoError(t, err)
	require.False(t, s.IsMaxed())
}
