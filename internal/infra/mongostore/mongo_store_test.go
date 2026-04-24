package mongostore_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/comfforts/logger"

	"github.com/comfforts/comff-stores/internal/infra/mongostore"
	envutils "github.com/comfforts/comff-stores/pkg/utils/environ"
)

func TestMongoStore(t *testing.T) {
	// Initialize logger
	l := logger.GetSlogLogger()
	t.Log("TestMongoStore Logger initialized")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	nmCfg := envutils.BuildMongoStoreConfig(true)
	cl, err := mongostore.NewMongoStore(ctx, nmCfg)
	require.NoError(t, err)

	defer func() {
		err := cl.Close(ctx)
		require.NoError(t, err)
	}()
}

func TestConnectionBuilder(t *testing.T) {
	// Initialize logger
	l := logger.GetSlogLogger()
	t.Log("TestConnectionBuilder Logger initialized")

	cfg := envutils.BuildMongoStoreConfig(false)
	builder := mongostore.NewMongoConnectionBuilder(
		cfg.Protocol(),
		cfg.Host(),
	).WithUser(
		cfg.User(),
	).WithPassword(
		cfg.Pwd(),
	).WithPassword(
		cfg.Pwd(),
	).WithDatabase(
		cfg.Name(),
	).WithConnectionParams(
		cfg.Params(),
	)

	connStr, err := builder.Build()
	require.NoError(t, err)

	l.Debug("MongoDB Connection String", "conn-string", connStr)
}
