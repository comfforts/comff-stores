package stores_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/comfforts/logger"

	stdom "github.com/comfforts/comff-stores/internal/domain/stores"
	"github.com/comfforts/comff-stores/internal/infra/mongostore"
	strepo "github.com/comfforts/comff-stores/internal/repo/stores"
	envutils "github.com/comfforts/comff-stores/pkg/utils/environ"
)

func TestStoresRepo(t *testing.T) {
	// Initialize logger
	l := logger.GetSlogLogger()
	l.Debug("TestStoresRepo Logger initialized")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	nmCfg := envutils.BuildMongoStoreConfig(true)
	cl, err := mongostore.NewMongoStore(ctx, nmCfg)
	require.NoError(t, err)

	storesRepo, err := strepo.NewStoresRepo(ctx, cl, nil)
	require.NoError(t, err)

	defer func() {
		err := storesRepo.Close(ctx)
		require.NoError(t, err)
	}()
}

func TestStoresCRUD(t *testing.T) {
	// Initialize logger
	l := logger.GetSlogLogger()
	l.Debug("TestStoresCRUD Logger initialized")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	nmCfg := envutils.BuildMongoStoreConfig(true)
	cl, err := mongostore.NewMongoStore(ctx, nmCfg)
	require.NoError(t, err)

	storesRepo, err := strepo.NewStoresRepo(ctx, cl, nil)
	require.NoError(t, err)

	defer func() {
		err := storesRepo.Close(ctx)
		require.NoError(t, err)
	}()

	id, err := storesRepo.AddStore(ctx, &stdom.Store{
		Name:      "Test Store",
		Org:       "Test Org",
		AddressId: "Test Address ID",
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	store, err := storesRepo.GetStore(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, store)
	require.Equal(t, "Test Store", store.Name)
	require.Equal(t, "Test Org", store.Org)
	require.Equal(t, "Test Address ID", store.AddressId)

	err = storesRepo.UpdateStore(ctx, id, &stdom.UpdateStoreQuery{
		Name: "Updated Store",
	})
	require.NoError(t, err)
	store, err = storesRepo.GetStore(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Updated Store", store.Name)

	_, err = storesRepo.AddStore(ctx, &stdom.Store{
		Name:      "Test Store",
		Org:       "Test Org",
		AddressId: "Test Address ID",
	})
	require.ErrorIs(t, err, strepo.ErrDuplicateStore)

	err = storesRepo.DeleteStore(ctx, id)
	require.NoError(t, err)

	_, err = storesRepo.GetStore(ctx, id)
	require.ErrorIs(t, err, strepo.ErrNoStore)
}
