package constants

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

const ReadRateContextKey = ContextKey("readrate")

const FileNameContextKey = ContextKey("filename")
const FilePathContextKey = ContextKey("filepath")
