package file

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/comfforts/comff-stores/pkg/errors"
)

const (
	ERROR_NO_FILE                 string = "%s doesn't exist"
	ERROR_FILE_INACCESSIBLE       string = "%s inaccessible"
	ERROR_NOT_A_FILE              string = "%s not a file"
	ERROR_OPENING_FILE            string = "opening file %s"
	ERROR_READING_FILE            string = "reading file %s"
	ERROR_DECODING_RESULT         string = "error decoding result json"
	ERROR_START_TOKEN             string = "error reading start token"
	ERROR_END_TOKEN               string = "error reading end token"
	ERROR_CLOSING_FILE            string = "closing file %s"
	ERROR_CREATING_FILE           string = "creating file %s"
	ERROR_WRITING_FILE            string = "writing file %s"
	ERROR_UPLOADING_FILE          string = "uploading file %s"
	ERROR_DOWNLOADING_FILE        string = "downloading file %s"
	ERROR_CREATING_STORAGE_CLIENT string = "error creating storage client"
	ERROR_LISTING_OBJECTS         string = "error listing storage bucket objects"
	ERROR_DELETING_OBJECTS        string = "error deleting storage bucket objects"
	ERROR_MISSING_BUCKET_NAME     string = "bucket name missing"
	ERROR_MISSING_FILE_NAME       string = "file name missing"
	ERROR_CREATING_DATA_DIR       string = "creating data directory %s"
	ERROR_STALE_UPLOAD            string = "storage bucket object has updates"
	ERROR_STALE_DOWNLOAD          string = "file object has updates"
)

var (
	ErrStartToken        = errors.NewAppError(ERROR_START_TOKEN)
	ErrEndToken          = errors.NewAppError(ERROR_END_TOKEN)
	ErrBucketNameMissing = errors.NewAppError(ERROR_MISSING_BUCKET_NAME)
	ErrFileNameMissing   = errors.NewAppError(ERROR_MISSING_FILE_NAME)
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
