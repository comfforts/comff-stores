package config

import (
	"os"
	"path/filepath"
)

const (
	CAFile               = "ca.pem"
	ServerCertFile       = "server.pem"
	ServerKeyFile        = "server-key.pem"
	ClientCertFile       = "client.pem"
	ClientKeyFile        = "client-key.pem"
	NobodyClientCertFile = "nobody-client.pem"
	NobodyClientKeyFile  = "nobody-client-key.pem"
	ACLModelFile         = "model.conf"
	ACLPolicyFile        = "policy.csv"
)

type FileType string

const (
	Policy FileType = "Policy"
	Certs  FileType = "Certs"
)

func CertFile(fileName string) string {
	return configFile(fileName, Certs)
}

func PolicyFile(fileName string) string {
	return configFile(fileName, Policy)
}

func configFile(fileName string, fileType FileType) string {
	if fileType == Policy {
		if dir := os.Getenv("POLICY_PATH"); dir != "" {
			return filepath.Join(dir, fileName)
		}
		return filepath.Join("policies", fileName)
	}

	if dir := os.Getenv("CERTS_PATH"); dir != "" {
		return filepath.Join(dir, fileName)
	}
	return filepath.Join("certs", fileName)
}
