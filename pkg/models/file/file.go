package file

import "context"

type Loader interface {
	ProcessFile(ctx context.Context, filePath string) error
}

type JSONMapper = map[string]interface{}
