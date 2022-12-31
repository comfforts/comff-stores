package file

import (
	"testing"
)

func TestHomeDir(t *testing.T) {
	dir := HomeDir()
	t.Logf("home dir: %s", dir)
}

func TestRootDir(t *testing.T) {
	dir := rootDir()
	t.Logf("root dir: %s", dir)
}
