package environ

import (
	"fmt"
	"os"

	indom "github.com/comfforts/comff-stores/internal/domain/infra"
	"github.com/comfforts/comff-stores/internal/infra/mongostore"
)

func BuildMongoStoreConfig(direct bool) indom.StoreConfig {
	dbProtocol := os.Getenv("MONGO_PROTOCOL")
	dbUser := os.Getenv("MONGO_USERNAME")
	dbPwd := os.Getenv("MONGO_PASSWORD")
	dbName := os.Getenv("MONGO_DBNAME")

	dbHost := os.Getenv("MONGO_HOST_LIST")
	dbParams := os.Getenv("MONGO_CLUS_CONN_PARAMS")
	if direct {
		dbParams = os.Getenv("MONGO_DIR_CONN_PARAMS")
		dbHost = os.Getenv("MONGO_HOST_NAME")
	}
	return mongostore.NewMongoDBConfig(dbProtocol, dbHost, dbUser, dbPwd, dbParams, dbName)
}

func BuildMetricsConfig() (string, string) {
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = ":9463"
	} else {
		metricsPort = fmt.Sprintf(":%s", metricsPort)
	}
	otelEndpoint := os.Getenv("OTEL_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "otel-collector:4317"
	}
	return metricsPort, otelEndpoint
}

func BuildServerTLSConfig() indom.TLSConfig {
	caFilePath := os.Getenv("TLS_CA_FILE")
	certFilePath := os.Getenv("TLS_CERT_FILE")
	keyFilePath := os.Getenv("TLS_KEY_FILE")
	return indom.TLSConfig{
		CAFilePath:   caFilePath,
		CertFilePath: certFilePath,
		KeyFilePath:  keyFilePath,
	}
}

func BuildClientTLSConfig() indom.TLSConfig {
	caFilePath := os.Getenv("CLIENT_TLS_CA_FILE")
	certFilePath := os.Getenv("CLIENT_TLS_CERT_FILE")
	keyFilePath := os.Getenv("CLIENT_TLS_KEY_FILE")
	return indom.TLSConfig{
		CAFilePath:   caFilePath,
		CertFilePath: certFilePath,
		KeyFilePath:  keyFilePath,
	}
}

func BuildNobodyClientTLSConfig() indom.TLSConfig {
	caFilePath := os.Getenv("NB_CLIENT_TLS_CA_FILE")
	certFilePath := os.Getenv("NB_CLIENT_TLS_CERT_FILE")
	keyFilePath := os.Getenv("NB_CLIENT_TLS_KEY_FILE")
	return indom.TLSConfig{
		CAFilePath:   caFilePath,
		CertFilePath: certFilePath,
		KeyFilePath:  keyFilePath,
	}
}
