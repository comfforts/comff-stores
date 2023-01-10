package log

type Config struct {
	Segment struct {
		MaxIndexSize  uint64
		InitialOffset uint64
	}
}
