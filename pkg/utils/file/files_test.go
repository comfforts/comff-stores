package file

import (
	"path"
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

func TestPathJoin(t *testing.T) {
	rootDir := ""
	policyDir := ""
	fPath := path.Join(rootDir, policyDir, "file.json")
	t.Logf("fPath: %s", fPath)

	policyDir = "policies"
	fPath = path.Join(rootDir, policyDir, "file.json")
	t.Logf("fPath: %s", fPath)
}
