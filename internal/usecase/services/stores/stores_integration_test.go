package stores_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	geocl "github.com/comfforts/comff-geo-client"
	geo_v1 "github.com/comfforts/comff-geo/api/geo/v1"
	"github.com/comfforts/logger"

	stdom "github.com/comfforts/comff-stores/internal/domain/stores"
	"github.com/comfforts/comff-stores/internal/infra/mongostore"
	"github.com/comfforts/comff-stores/internal/infra/observability"
	strepo "github.com/comfforts/comff-stores/internal/repo/stores"
	"github.com/comfforts/comff-stores/internal/usecase/services/stores"
	envutils "github.com/comfforts/comff-stores/pkg/utils/environ"
	testutils "github.com/comfforts/comff-stores/pkg/utils/test"
)

func TestStoresService(t *testing.T) {
	// Initialize logger
	l := logger.GetSlogLogger()
	l.Debug("TestStoresRepo started")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	metrics, err := observability.NewMetrics()
	require.NoError(t, err)

	// Initialize MongoDB store
	nmCfg := envutils.BuildMongoStoreConfig(true)
	ms, err := mongostore.NewMongoStore(ctx, nmCfg)
	require.NoError(t, err)

	// Initialize stores repository
	sr, err := strepo.NewStoresRepo(ctx, ms, metrics)
	require.NoError(t, err)
	defer func() {
		err := sr.Close(ctx)
		require.NoError(t, err)
	}()

	// Initialize geo client options
	clientOpts := geocl.NewDefaultClientOption()
	clientOpts.Caller = "geo-service-geo-client-test"

	// Initialize geo client
	gc, err := geocl.NewClient(ctx, clientOpts)
	require.NoError(t, err)
	defer func() {
		err := gc.Close(ctx)
		require.NoError(t, err)
	}()

	// Initialize stores service
	_, err = stores.NewStoresService(ctx, sr, gc, metrics)
	require.NoError(t, err)
	l.Debug("TestStoresRepo done")
}

func TestStoresServiceCRUD(t *testing.T) {
	// Initialize logger
	l := logger.GetSlogLogger()
	l.Debug("TestStoresServiceCRUD started")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	metrics, err := observability.NewMetrics()
	require.NoError(t, err)

	// Initialize MongoDB store
	nmCfg := envutils.BuildMongoStoreConfig(true)
	ms, err := mongostore.NewMongoStore(ctx, nmCfg)
	require.NoError(t, err)

	// Initialize stores repository
	sr, err := strepo.NewStoresRepo(ctx, ms, metrics)
	require.NoError(t, err)
	defer func() {
		err := sr.Close(ctx)
		require.NoError(t, err)
	}()

	// Initialize geo client options
	clientOpts := geocl.NewDefaultClientOption()
	clientOpts.Caller = "geo-service-geo-client-test"

	// Initialize geo client
	gc, err := geocl.NewClient(ctx, clientOpts)
	require.NoError(t, err)
	defer func() {
		err := gc.Close(ctx)
		require.NoError(t, err)
	}()

	// Initialize stores service
	ss, err := stores.NewStoresService(ctx, sr, gc, metrics)
	require.NoError(t, err)

	// Test AddStore with valid data
	storeId, err := ss.AddStore(ctx, &stdom.AddStoreParams{
		Name:      "Test Store",
		Org:       "Test Org",
		AddressId: "dacdbddabcadccbdacac", // Use a valid address ID for testing
	})
	require.NoError(t, err)
	require.NotEmpty(t, storeId)

	// Test GetStore
	store, err := ss.GetStore(ctx, storeId)
	require.NoError(t, err)
	require.NotNil(t, store)
	require.Equal(t, "Test Store", store.Name)
	require.Equal(t, "Test Org", store.Org)
	require.Equal(t, "dacdbddabcadccbdacac", store.AddressId)

	err = ss.UpdateStore(ctx, storeId, &stdom.UpdateStoreParams{
		Name: "Updated Test Store",
	})
	require.NoError(t, err)

	store, err = ss.GetStore(ctx, storeId)
	require.NoError(t, err)
	require.NotNil(t, store)
	require.Equal(t, "Updated Test Store", store.Name)

	// Test DeleteStore
	err = ss.DeleteStore(ctx, storeId)
	require.NoError(t, err)

	// Verify store is deleted
	deletedStore, err := ss.GetStore(ctx, storeId)
	require.Error(t, err)
	require.Equal(t, err, strepo.ErrNoStore)
	require.Nil(t, deletedStore)
	l.Debug("TestStoresServiceCRUD done")
}

func TestStoresServiceSearchStores(t *testing.T) {
	// Initialize logger
	l := logger.GetSlogLogger()
	l.Debug("TestStoresServiceSearchStores started")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ctx = logger.WithLogger(ctx, l)

	metrics, err := observability.NewMetrics()
	require.NoError(t, err)

	// Initialize MongoDB store
	nmCfg := envutils.BuildMongoStoreConfig(true)
	ms, err := mongostore.NewMongoStore(ctx, nmCfg)
	require.NoError(t, err)

	// Initialize stores repository
	sr, err := strepo.NewStoresRepo(ctx, ms, metrics)
	require.NoError(t, err)
	defer func() {
		err := sr.Close(ctx)
		require.NoError(t, err)
	}()

	// Initialize geo client options
	clientOpts := geocl.NewDefaultClientOption()
	clientOpts.Caller = "stores-service-geo-client-test"

	// Initialize geo client
	gc, err := geocl.NewClient(ctx, clientOpts)
	require.NoError(t, err)
	defer func() {
		err := gc.Close(ctx)
		require.NoError(t, err)
	}()

	// Initialize stores service
	ss, err := stores.NewStoresService(ctx, sr, gc, metrics)
	require.NoError(t, err)

	addrIdMap := map[string]*geo_v1.Point{}
	addrIds := []string{}

	dests := testutils.BuildPetalumaSet1()
	for i, dest := range dests {
		l.Debug("Testing GeoLocate for destination", "index", i, "destination", dest)
		resp, err := gc.GeoLocate(ctx, dest)
		require.NoError(t, err)
		require.NotNil(t, resp)
		addrIdMap[resp.Point.Hash] = resp.Point
		addrIds = append(addrIds, resp.Point.Hash)
	}

	dests = testutils.BuildPetalumaSet2()
	for i, dest := range dests {
		l.Debug("Testing GeoLocate for destination", "index", i, "destination", dest)
		resp, err := gc.GeoLocate(ctx, dest)
		require.NoError(t, err)
		require.NotNil(t, resp)
		addrIdMap[resp.Point.Hash] = resp.Point
		addrIds = append(addrIds, resp.Point.Hash)
	}

	stIds := []string{}
	for i, addrId := range addrIds {
		storeId, err := ss.AddStore(ctx, &stdom.AddStoreParams{
			Name:      fmt.Sprintf("Test Store %d", i),
			Org:       fmt.Sprintf("Test Org %d", i%2), // Alternate between 2 orgs
			AddressId: addrId,
		})
		require.NoError(t, err)
		require.NotEmpty(t, storeId)
		l.Debug("Added store for address", "index", i, "addressId", addrId, "storeId", storeId)
		stIds = append(stIds, storeId)
	}
	defer func() {
		for i, stId := range stIds {
			err := ss.DeleteStore(ctx, stId)
			require.NoError(t, err)
			l.Debug("Deleted store", "index", i, "storeId", stId)
		}
	}()

	// org search
	sts, err := ss.SearchStores(ctx, &stdom.SearchStoreParams{
		Org: "Test Org 0",
	})
	require.NoError(t, err)
	require.NotNil(t, sts)
	require.GreaterOrEqual(t, len(sts), 1)
	l.Debug("SearchStores returned stores", "count", len(sts))

	// name search
	sts, err = ss.SearchStores(ctx, &stdom.SearchStoreParams{
		Name: "Test Store 0",
	})
	require.NoError(t, err)
	require.NotNil(t, sts)
	require.GreaterOrEqual(t, len(sts), 1)
	l.Debug("SearchStores returned stores", "count", len(sts))

	// address id search
	sts, err = ss.SearchStores(ctx, &stdom.SearchStoreParams{
		AddressId: addrIds[0],
	})
	require.NoError(t, err)
	require.NotNil(t, sts)
	require.Equal(t, 1, len(sts))
	l.Debug("SearchStores returned stores", "count", len(sts))

	// lat/long search
	sts, err = ss.SearchStores(ctx, &stdom.SearchStoreParams{
		Latitude:  addrIdMap[addrIds[1]].Latitude,
		Longitude: addrIdMap[addrIds[1]].Longitude,
		Distance:  50000, // meters
	})
	require.NoError(t, err)
	require.NotNil(t, sts)
	require.GreaterOrEqual(t, len(sts), 1)
	l.Debug("SearchStores returned stores", "count", len(sts))

	// address string search
	sts, err = ss.SearchStores(ctx, &stdom.SearchStoreParams{
		AddressStr: addrIdMap[addrIds[1]].FormattedAddress,
		Distance:   50000, // meters
	})
	require.NoError(t, err)
	require.NotNil(t, sts)
	require.GreaterOrEqual(t, len(sts), 1)
	l.Debug("SearchStores returned stores", "count", len(sts))

	l.Debug("TestStoresServiceSearchStores done")
}
