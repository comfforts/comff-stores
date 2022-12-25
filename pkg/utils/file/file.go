package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func CreateDirectory(path string) error {
	_, err := os.Stat(filepath.Dir(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
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
