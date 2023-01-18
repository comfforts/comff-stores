package file

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/comfforts/errors"
)

const (
	ERROR_NO_FILE           string = "%s doesn't exist"
	ERROR_FILE_INACCESSIBLE string = "%s inaccessible"
	ERROR_OPENING_FILE      string = "opening file %s"
	ERROR_DECODING_RESULT   string = "error decoding result json"
	ERROR_START_TOKEN       string = "error reading start token"
	ERROR_END_TOKEN         string = "error reading end token"
	ERROR_CREATING_FILE     string = "creating file %s"
	ERROR_CREATING_DATA_DIR string = "creating data directory %s"
)

var (
	ErrStartToken = errors.NewAppError(ERROR_START_TOKEN)
	ErrEndToken   = errors.NewAppError(ERROR_END_TOKEN)
)

type Bases string

const (
	Internal Bases = "internal"
	Utils    Bases = "utils"
)

func CreateDirectory(path string) error {
	_, err := os.Stat(filepath.Dir(path))
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

func GetRoot(path string) string {
	idx := strings.Index(path, "/")
	if idx < 1 {
		return ""
	}
	root := path[0:idx]
	return root

}

// checks if file exists
func FileStats(filePath string) (fs.FileInfo, error) {
	fStats, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fStats, errors.WrapError(err, ERROR_NO_FILE, filePath)
		} else {
			return fStats, errors.WrapError(err, ERROR_FILE_INACCESSIBLE, filePath)
		}
	}
	return fStats, nil
}

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func HomeDir() string {
	home := rootDir()
	base := filepath.Base(home)
	if base == string(Internal) {
		home = filepath.Join("..")
	} else if base == string(Utils) {
		home = filepath.Join("../..")
	}
	fmt.Printf("    base: %s, home: %s\n", base, home)
	return home
}
