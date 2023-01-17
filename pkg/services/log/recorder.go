package log

import (
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/comfforts/comff-stores/pkg/errors"

	api "github.com/comfforts/comff-stores/api/v1"
)

const (
	ERROR_OFFSET_OUT_OF_RANGE string = "requested offset is outside the log's range: %d"
)

type Recorder struct {
	mu sync.RWMutex

	Dir    string
	Config Config

	activeSegment *segmenter
	segments      []*segmenter
}

func NewRecorder(dir string, c Config) (*Recorder, error) {
	if c.Segment.MaxIndexSize == 0 {
		c.Segment.MaxIndexSize = 100
	}
	r := &Recorder{
		Dir:    dir,
		Config: c,
	}

	return r, r.setup()
}

func (r *Recorder) setup() error {
	files, err := os.ReadDir(r.Dir)
	if err != nil {
		log.Printf("recorder.setup() - error reading direcotry, %s", r.Dir)
		return err
	}
	var baseOffsets []uint64
	for _, file := range files {
		offStr := strings.TrimSuffix(
			file.Name(),
			path.Ext(file.Name()),
		)
		off, _ := strconv.ParseUint(offStr, 10, 0)
		baseOffsets = append(baseOffsets, off)
	}
	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})
	for i := 0; i < len(baseOffsets); i++ {
		if err = r.newSegmenter(baseOffsets[i]); err != nil {
			log.Printf("recorder.setup() - error creating segment with baseoffset %d", baseOffsets[i])
			return err
		}
		// baseOffset contains dup for index and store so we skip
		// the dup
		i++
	}
	if r.segments == nil {
		log.Printf("recorder.setup() - initializing segment, initial offset: %d", r.Config.Segment.InitialOffset)
		if err = r.newSegmenter(r.Config.Segment.InitialOffset); err != nil {
			log.Printf("recorder.setup() - error initializing segment, initial offset %d", r.Config.Segment.InitialOffset)
			return err
		}
	}
	return nil
}

func (r *Recorder) newSegmenter(off uint64) error {
	log.Printf("recorder.newSegmenter() - creating new segment, offset: %d", off)
	s, err := newSegmenter(r.Dir, off, r.Config)
	if err != nil {
		log.Printf("recorder.newSegmenter() - error creating new segment, offset: %d, error: %v", off, err)
		return err
	}
	r.segments = append(r.segments, s)
	if r.activeSegment != nil && r.activeSegment.baseOffset != r.Config.Segment.InitialOffset {
		r.activeSegment.Close()
	}
	r.activeSegment = s
	return nil
}

func (r *Recorder) Append(record *api.Record) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	off, err := r.activeSegment.Append(record)
	if err != nil {
		log.Printf("recorder.Append() - error appending record, offset: %d, error: %v", off, err)
		return 0, err
	}
	log.Printf("recorder.Append() - appended record, offset: %d", off)
	if r.activeSegment.IsMaxed() {
		err = r.newSegmenter(off + 1)
	}
	if err != nil {
		log.Printf("recorder.Append() - error appending record, offset: %d, error: %v", off, err)
	}
	return off, err
}

func (r *Recorder) Read(off uint64) (*api.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var s *segmenter
	log.Printf("recorder.Read() - read offset: %d, has segments: %d", off, len(r.segments))
	for _, segment := range r.segments {
		if segment.baseOffset <= off && off < segment.nextOffset {
			s = segment
			break
		}
	}
	if s == nil || s.nextOffset <= off {
		log.Printf("recorder.Read() - nil segment or out of bounds ofset, segmenter: %v, segments: %d", s, len(r.segments))
		return nil, errors.NewAppError(ERROR_OFFSET_OUT_OF_RANGE, off)
	}
	log.Printf("recorder.Read() - read segment's base offset: %d, nextOffset: %d", s.baseOffset, s.nextOffset)
	return s.Read(off)
}

func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.Printf("recorder.Close() - closing recorder")
	for _, segment := range r.segments {
		if !segment.Closed() {
			if err := segment.Close(); err != nil {
				log.Printf("recorder.Close() - error closing recorder, err: %v", err)
				return err
			}
		}
	}
	return nil
}

func (r *Recorder) Remove() error {
	if err := r.Close(); err != nil {
		log.Printf("recorder.Remove() - error removing recorder")
		return err
	}
	return os.RemoveAll(r.Dir)
}

func (r *Recorder) Reset() error {
	if err := r.Remove(); err != nil {
		log.Printf("recorder.Reset() - error resetting recorder")
		return err
	}
	return r.setup()
}

func (r *Recorder) LowestOffset() (uint64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.segments[0].baseOffset, nil
}

func (r *Recorder) HighestOffset() (uint64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	off := r.segments[len(r.segments)-1].nextOffset
	if off == 0 {
		return 0, nil
	}
	return off - 1, nil
}

func (r *Recorder) Truncate(lowest uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var segments []*segmenter
	for _, s := range r.segments {
		if s.nextOffset <= lowest+1 {
			if err := s.Remove(); err != nil {
				log.Printf("recorder.Truncate() - error truncating recorder")
				return err
			}
			continue
		}
		segments = append(segments, s)
	}
	r.segments = segments
	return nil
}

func (r *Recorder) Reader() io.Reader {
	r.mu.RLock()
	defer r.mu.RUnlock()
	readers := make([]io.Reader, len(r.segments))
	for i, segment := range r.segments {
		readers[i] = &originReader{segment.filer, 0}
	}
	return io.MultiReader(readers...)
}

type originReader struct {
	*filer
	off int64
}

func (o *originReader) Read(p []byte) (int, error) {
	n, err := o.ReadAt(p, o.off)
	o.off += int64(n)
	return n, err
}
