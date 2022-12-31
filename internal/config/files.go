package config

import (
	"os"
	"path/filepath"

	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

var (
	CAFile               = certFile("ca.pem")
	ServerCertFile       = certFile("server.pem")
	ServerKeyFile        = certFile("server-key.pem")
	ClientCertFile       = certFile("client.pem")
	ClientKeyFile        = certFile("client-key.pem")
	NobodyClientCertFile = certFile("nobody-client.pem")
	NobodyClientKeyFile  = certFile("nobody-client-key.pem")
	ACLModelFile         = policyFile("model.conf")
	ACLPolicyFile        = policyFile("policy.csv")
)

type FileType string

const (
	Policy FileType = "Policy"
	Certs  FileType = "Certs"
)

func certFile(fileName string) string {
	return configFile(fileName, Certs)
}

func policyFile(fileName string) string {
	return configFile(fileName, Policy)
}

func configFile(fileName string, fileType FileType) string {
	homeDir := fileUtils.HomeDir()
	if fileType == Policy {
		if dir := os.Getenv("POLICY_PATH"); dir != "" {
			return filepath.Join(dir, fileName)
		}
		return filepath.Join(homeDir, ".policies", fileName)
	}

	if dir := os.Getenv("CERTS_PATH"); dir != "" {
		return filepath.Join(dir, fileName)
	}

	return filepath.Join(homeDir, ".certs", fileName)
}
