package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/comfforts/comff-stores/pkg/errors"
	"google.golang.org/protobuf/proto"

	api "github.com/comfforts/comff-stores/api/v1"
)

const (
	ERROR_OPENING_FILER      string = "error opening filer %s"
	ERROR_OPENING_INDEX      string = "error opening index %s"
	ERROR_REMOVING_FILER     string = "error removing filer %s"
	ERROR_REMOVING_INDEX     string = "error removing index %s"
	ERROR_MARSHALLING_RECORD string = "error marshalling record"
)

type segmenter struct {
	filer                  *filer
	indexer                *indexer
	baseOffset, nextOffset uint64
	config                 Config
	closed                 bool
}

func newSegmenter(dir string, baseOffset uint64, c Config) (*segmenter, error) {
	s := &segmenter{
		baseOffset: baseOffset,
		config:     c,
	}

	fPath := path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".filer"))
	var filerFile *os.File
	_, err := os.Stat(fPath)
	if err != nil {
		filerFile, err = os.Create(fPath)
	} else {
		filerFile, err = os.Open(fPath)
		log.Printf("opened existing filer: %s", fPath)
	}
	if err != nil {
		return nil, errors.WrapError(err, ERROR_OPENING_FILER, fPath)
	}

	if s.filer, err = newFiler(filerFile); err != nil {
		return nil, err
	}

	var indexFile *os.File
	iPath := path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index"))
	_, err = os.Stat(iPath)
	if err != nil {
		indexFile, err = os.Create(iPath)
	} else {
		indexFile, err = os.Open(iPath)
		log.Printf("opened existing index: %s", iPath)
	}
	if err != nil {
		return nil, errors.WrapError(err, ERROR_OPENING_INDEX, iPath)
	}

	if s.indexer, err = newIndexer(indexFile, c); err != nil {
		return nil, err
	}
	log.Printf("indexer size: %d", s.indexer.size)
	if off, _, err := s.indexer.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1
	}
	return s, nil
}

func (s *segmenter) Append(record *api.Record) (offset uint64, err error) {
	if s.IsMaxed() {
		return 0, io.EOF
	}

	cur := s.nextOffset
	record.Offset = cur

	p, err := proto.Marshal(record)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_MARSHALLING_RECORD)
	}

	_, pos, err := s.filer.Append(p)
	if err != nil {
		return 0, err
	}
	if err = s.indexer.Write(
		// index offsets are relative to base offset
		uint32(s.nextOffset-uint64(s.baseOffset)),
		pos,
	); err != nil {
		return 0, err
	}
	s.nextOffset++
	return cur, nil
}

func (s *segmenter) Read(off uint64) (*api.Record, error) {
	_, pos, err := s.indexer.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}
	log.Printf("reading position: %d", pos)
	p, err := s.filer.Read(pos)
	if err != nil {
		return nil, err
	}
	record := &api.Record{}
	err = proto.Unmarshal(p, record)
	return record, err
}

func (s *segmenter) IsMaxed() bool {
	return s.indexer.size >= s.config.Segment.MaxIndexSize
}

func (s *segmenter) Close() error {
	log.Printf("closing segmenter - offset - base: %d, next: %d", s.baseOffset, s.nextOffset)
	if err := s.indexer.Close(); err != nil {
		return err
	}
	if err := s.filer.Close(); err != nil {
		return err
	}
	s.closed = true
	return nil
}

func (s *segmenter) Remove() error {
	if !s.Closed() {
		if err := s.Close(); err != nil {
			return err
		}
	}
	if err := os.Remove(s.indexer.Name()); err != nil {
		return errors.WrapError(err, ERROR_REMOVING_INDEX, s.indexer.Name())
	}
	if err := os.Remove(s.filer.Name()); err != nil {
		return errors.WrapError(err, ERROR_REMOVING_FILER, s.filer.Name())
	}
	return nil
}

func (s *segmenter) Closed() bool {
	return s.closed
}
