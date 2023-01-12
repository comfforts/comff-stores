package log

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"sync"

	"github.com/comfforts/comff-stores/pkg/errors"
)

var (
	OFFSET_WIDTH   uint64 = 4
	POSITION_WIDTH uint64 = 8
	ENTRY_WIDTH           = OFFSET_WIDTH + POSITION_WIDTH
)

const (
	ERROR_DECODING_INDEX_FILE string = "error decoding file %s"
	ERROR_ENCODING_INDEX_FILE string = "error encoding file %s"
	ERROR_DUPLICATE_OFFSET    string = "error offset already exists"
	ERROR_GETTING_RECORD_POS  string = "error fetching record position"
)

var (
	ErrDuplicateOffset = errors.NewAppError(ERROR_DUPLICATE_OFFSET)
	ErrRecordPosition  = errors.NewAppError(ERROR_GETTING_RECORD_POS)
)

type Mapper = map[uint32]uint64

type indexer struct {
	file   *os.File
	size   uint64
	mapper Mapper
	mu     sync.Mutex
}

func newIndexer(f *os.File, c Config) (*indexer, error) {
	idx := &indexer{
		file: f,
	}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, errors.WrapError(err, ERROR_NO_FILE, f.Name())
	}

	log.Printf("indexer file size: %d", fi.Size())

	if fi.Size() > 0 {
		decoder := gob.NewDecoder(f)
		err = decoder.Decode(&idx.mapper)
		if err != nil {
			return nil, errors.WrapError(err, ERROR_DECODING_INDEX_FILE, f.Name())
		}
		idx.size = uint64(len(idx.mapper) - 1)
	} else {
		idx.mapper = Mapper{}
	}

	return idx, nil
}

func (i *indexer) Write(off uint32, pos uint64) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if off > uint32(i.size) {
		// TOCHECK should return specific error, rather than io.EOF error?
		return io.EOF
	}

	_, ok := i.mapper[off]
	if ok {
		return ErrDuplicateOffset
	}
	i.mapper[off] = pos
	i.size++
	return nil
}

func (i *indexer) Read(inOff int64) (outOff uint32, pos uint64, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// new index, return io.EOF error as signal for segment to initalize with baseoffset
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	if inOff == -1 {
		outOff = uint32(i.size)
	} else {
		outOff = uint32(inOff)
	}

	if outOff > uint32(i.size) {
		// TOCHECK should return specific error, rather than io.EOF error?
		return 0, 0, io.EOF
	}

	pos, ok := i.mapper[outOff]
	if !ok {
		return 0, 0, ErrRecordPosition
	}
	return outOff, pos, nil
}

func (i *indexer) Close() error {
	i.file.Close()

	fi, err := os.Create(i.Name())
	if err != nil {
		return errors.WrapError(err, ERROR_ENCODING_INDEX_FILE, i.Name())
	}
	encoder := gob.NewEncoder(fi)
	err = encoder.Encode(&i.mapper)
	if err != nil {
		return errors.WrapError(err, ERROR_ENCODING_INDEX_FILE, i.Name())
	}

	fs, err := os.Stat(fi.Name())
	if err != nil {
		return errors.WrapError(err, ERROR_NO_FILE, fi.Name())
	}

	i.file = fi
	log.Printf("indexer file saved and closed, file: %s, file size: %d, index size: %d", i.Name(), fs.Size(), i.size)
	return i.file.Close()
}

func (i *indexer) Name() string {
	return i.file.Name()
}
