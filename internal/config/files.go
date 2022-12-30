package config

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
)

var (
	CAFile         = configFile("ca.pem")
	ServerCertFile = configFile("server.pem")
	ServerKeyFile  = configFile("server-key.pem")
	ClientCertFile = configFile("client.pem")
	ClientKeyFile  = configFile("client-key.pem")
)

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func configFile(fileName string) string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, fileName)
	}

	rootDir := rootDir()
	return filepath.Join(rootDir, ".certs", fileName)

	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	panic(err)
	// }
	// return filepath.Join(homeDir, ".certs", fileName)
}
