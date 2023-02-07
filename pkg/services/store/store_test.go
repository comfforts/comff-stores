package store

import (
	"context"
	"os"
	"testing"

	"github.com/comfforts/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comfforts/comff-stores/pkg/models"

	testUtils "github.com/comfforts/comff-stores/pkg/utils/test"
)

const TEST_DIR = "test-data"

func TestStoreService(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		ss *StoreService,
	){
		"mapping result json to store succeeds":                 testMapResultToStore,
		"adding store json to store and getting stats succeeds": testAddStoreGetStats,
		"adding store json to store and getting store succeeds": testAddStoreGetStore,
	} {
		t.Run(scenario, func(t *testing.T) {
			ss, teardown := setupStoreTest(t)
			defer teardown()
			fn(t, ss)
		})
	}
}

func setupStoreTest(t *testing.T) (
	ss *StoreService,
	teardown func(),
) {
	t.Helper()

	appLogger := logger.NewTestAppLogger(TEST_DIR)
	ss, err := NewStoreService(appLogger)
	require.NoError(t, err)

	return ss, func() {
		t.Logf(" TestStoreService ended, will clear store data")
		ss.Close()

		err := os.RemoveAll(TEST_DIR)
		require.NoError(t, err)
	}
}

func testMapResultToStore(t *testing.T, ss *StoreService) {
	storeId, name, org, city, country := 1, "Plaza Hollywood", "starbucks", "Hong Kong", "CN"
	storeJSON := testUtils.CreateStoreJSON(uint64(storeId), name, org, city, country)

	store, err := models.MapResultToStore(storeJSON)
	require.NoError(t, err)

	assert.Equal(t, store.StoreId, uint64(storeId), "storeId should be mapped")
	assert.Equal(t, store.Name, name, "name should be mapped")
	assert.Equal(t, store.Org, org, "org should be mapped")
	assert.Equal(t, store.City, city, "city should be mapped")
}

func testAddStoreGetStats(t *testing.T, ss *StoreService) {
	t.Helper()

	storeId, name, org, city, country := 1, "Plaza Hollywood", "starbucks", "Hong Kong", "CN"
	sj := testUtils.CreateStoreJSON(uint64(storeId), name, org, city, country)

	store, err := models.MapResultToStore(sj)
	require.NoError(t, err)

	ctx := context.Background()
	st, err := ss.AddStore(ctx, store)
	require.NoError(t, err)
	assert.Equal(t, st.ID != "", true, "adding new store should be success")

	storeStats := ss.GetStoreStats()
	assert.Equal(t, storeStats.Count, 1, "store count should be equal to added store count")
	assert.Equal(t, storeStats.HashCount, 1, "geo hash count should be equal to added store count")
	assert.Equal(t, storeStats.Ready, false, "store ready status should be false")
}

func testAddStoreGetStore(t *testing.T, ss *StoreService) {
	t.Helper()

	storeId, name, org, city, country := 1, "Plaza Hollywood", "starbucks", "Hong Kong", "CN"
	sj := testUtils.CreateStoreJSON(uint64(storeId), name, org, city, country)

	store, err := models.MapResultToStore(sj)
	require.NoError(t, err)

	ctx := context.Background()
	st, err := ss.AddStore(ctx, store)
	require.NoError(t, err)
	assert.Equal(t, st.ID != "", true, "adding new store should be success")

	storeStats := ss.GetStoreStats()
	assert.Equal(t, storeStats.Count, 1, "store count should be equal to added store count")
	assert.Equal(t, storeStats.HashCount, 1, "geo hash count should be equal to added store count")
	assert.Equal(t, storeStats.Ready, false, "store ready status should be false")

	lat, ok := sj["latitude"].(float64)
	require.Equal(t, true, ok)

	long, ok := sj["longitude"].(float64)
	require.Equal(t, true, ok)

	id, err := ss.buildId(lat, long, org)
	require.NoError(t, err)

	savedStore, err := ss.GetStore(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, savedStore.ID, id, "store id should match")
	assert.Equal(t, savedStore.Org, org, "store id should match")
	assert.Equal(t, savedStore.Name, name, "store name should match")
	assert.Equal(t, savedStore.City, city, "city should match")
}

// func TestMapResultToStore(t *testing.T) {
// 	storeId, name, org, city := "1", "Plaza Hollywood", "starbucks", "Hong Kong"
// 	storeJSON := DefaulTestStoreJSON(storeId, name, org, city)

// 	store, err := MapResultToStore(storeJSON)
// 	require.NoError(t, err)

// 	assert.Equal(t, store.StoreId, storeId, "storeId should be mapped")
// 	assert.Equal(t, store.Name, name, "name should be mapped")
// 	assert.Equal(t, store.Org, org, "org should be mapped")
// 	assert.Equal(t, store.City, city, "city should be mapped")
// }

// func TestAddStoreGetStats(t *testing.T) {
// 	appLogger := logging.NewTestAppLogger(TEST_DIR)
// 	css := NewStoreService(appLogger)

// 	storeId, name, org, city := "1", "Plaza Hollywood", "starbucks", "Hong Kong"
// 	storeJSON := DefaulTestStoreJSON(storeId, name, org, city)

// 	testAddStore(t, css, storeJSON)
// 	testGetStoreStats(t, css, 1)
// }

// func TestAddStoreGetStore(t *testing.T) {
//  appLogger := logging.NewTestAppLogger(TEST_DIR)
// 	css := NewStoreService(appLogger)

// 	storeId, name, org, city := "1", "Plaza Hollywood", "starbucks", "Hong Kong"
// 	sj := DefaulTestStoreJSON(storeId, name, org, city)

// 	lat, ok := sj["latitude"].(float64)
// 	require.Equal(t, true, ok)

// 	long, ok := sj["longitude"].(float64)
// 	require.Equal(t, true, ok)

// 	id, err := BuildId(lat, long, org)
// 	require.NoError(t, err)

// 	testAddStore(t, css, sj)
// 	testGetStore(t, css, sj, id)
// }

// func testAddStore(t *testing.T, ss *StoreService, sj map[string]interface{}) {
// 	t.Helper()

// 	store, err := MapResultToStore(sj)
// 	require.NoError(t, err)

// 	ctx := context.Background()
// 	st, err := ss.AddStore(ctx, store)
// 	require.NoError(t, err)
// 	assert.Equal(t, st.ID != "", true, "adding new store should be success")
// }

// func testGetStoreStats(t *testing.T, ss *StoreService, count int) {
// 	t.Helper()
// 	storeStats := ss.GetStoreStats()
// 	assert.Equal(t, storeStats.Count, count, "store count should be equal to added store count")
// 	assert.Equal(t, storeStats.HashCount, count, "geo hash count should be equal to added store count")
// 	assert.Equal(t, storeStats.Ready, false, "store ready status should be false")
// }

// func testGetStore(t *testing.T, ss *StoreService, sj map[string]interface{}, id string) {
// 	t.Helper()
// 	ctx := context.Background()

// 	org, ok := sj["org"].(string)
// 	require.Equal(t, true, ok)

// 	name, ok := sj["name"].(string)
// 	require.Equal(t, true, ok)

// 	city, ok := sj["city"].(string)
// 	require.Equal(t, true, ok)

// 	store, err := ss.GetStore(ctx, id)
// 	require.NoError(t, err)
// 	assert.Equal(t, store.ID, id, "store id should match")
// 	assert.Equal(t, store.Org, org, "store id should match")
// 	assert.Equal(t, store.Name, name, "store name should match")
// 	assert.Equal(t, store.City, city, "city should match")
// }
