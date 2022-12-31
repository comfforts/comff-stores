package config

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/comfforts/comff-stores/pkg/errors"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

const (
	ERROR_TLS_CONFIG    string = "error creating TLS config"
	ERROR_PARSING_CERTS string = "error parsing certificate file %s"
)

type TLSConfig struct {
	CertFile      string
	KeyFile       string
	CAFile        string
	ServerAddress string
	Server        bool
}

func SetupTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		_, err = fileUtils.FileStats(cfg.CertFile)
		if err != nil {
			return nil, errors.WrapError(err, fileUtils.ERROR_NO_FILE, cfg.CertFile)
		}

		_, err = fileUtils.FileStats(cfg.KeyFile)
		if err != nil {
			return nil, errors.WrapError(err, fileUtils.ERROR_NO_FILE, cfg.KeyFile)
		}

		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, errors.WrapError(err, ERROR_TLS_CONFIG)
		}
	}
	if cfg.CAFile != "" {
		b, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, errors.WrapError(err, fileUtils.ERROR_NO_FILE, cfg.CAFile)
		}
		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM([]byte(b))
		if !ok {
			return nil, errors.NewAppError(ERROR_PARSING_CERTS, cfg.CAFile)
		}
		if cfg.Server {
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.RootCAs = ca
		}
		tlsConfig.ServerName = cfg.ServerAddress
	}

	return tlsConfig, nil
}