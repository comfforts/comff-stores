package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"github.com/comfforts/comff-stores/pkg/utils/file"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestStoreLoaderJob(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		loader *StoreLoader,
		testDir string,
	){
		"file operations check succeeds": testFileStats,
		"storing data file succeeds":     testSaveDataFile,
		"uploading data file succeeds":   testUploadDataFile,
		"local file processing succeeds": testProcessingLocalFile,
		"cloud file processing succeeds": testProcessingCloudFile,
	} {
		testDir := "testing_store_loader/"
		t.Run(scenario, func(t *testing.T) {
			loader, teardown := setupStoreLoader(t, testDir)
			defer teardown()
			fn(t, loader, testDir)
		})
	}
}

func setupStoreLoader(t *testing.T, testDir string) (
	sl *StoreLoader,
	teardown func(),
) {
	t.Helper()

	t.Logf(" setupStoreLoader: creating test directory %s", testDir)
	err := file.CreateDirectory(testDir)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)
	ss := store.NewStoreService(appLogger)

	t.Logf(" setupStoreLoader: getting app-config/loader-config %s", "test-config.json")
	appCfg, err := config.GetAppConfig(appLogger, "test-config.json")
	require.NoError(t, err)

	slCfg := appCfg.Jobs.StoreLoaderConfig

	sl, err = NewStoreLoader(slCfg, ss, appLogger)
	require.NoError(t, err)

	return sl, func() {
		t.Logf(" TestStoreService ended, will clear store data")
		sl.stores.Clear()

		err := os.RemoveAll(testDir)
		require.NoError(t, err)

		err = os.RemoveAll(slCfg.DataPath)
		require.NoError(t, err)
	}
}

func testFileStats(t *testing.T, loader *StoreLoader, testDir string) {
	name := "starbucks"
	fPath, err := createJSONFile(filepath.Join(testDir, "test"), name)
	require.NoError(t, err)

	require.Equal(t, ".json", filepath.Ext(fPath))
	require.Equal(t, "starbucks.json", filepath.Base(fPath))

	dir, err := os.Getwd()
	require.NoError(t, err)
	t.Logf("working directory: %s", dir)

	testPath := filepath.Join(dir, "../../cmd/store")
	t.Logf("testPath: %s", testPath)
}

func testSaveDataFile(t *testing.T, loader *StoreLoader, testDir string) {
	name := "test"
	fPath, err := createJSONFile(testDir, name)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = loader.StoreDataFile(ctx, fPath)
	require.NoError(t, err)
}

func testUploadDataFile(t *testing.T, loader *StoreLoader, testDir string) {
	name := "test"
	fPath, err := createJSONFile(testDir, name)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = loader.StoreDataFile(ctx, fPath)
	require.NoError(t, err)

	err = loader.UploadDataFile(ctx, fPath)
	require.NoError(t, err)
}

func testProcessingLocalFile(t *testing.T, loader *StoreLoader, testDir string) {
	name := "test"
	fPath, err := createJSONFile(testDir, name)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = loader.ProcessFile(ctx, fPath)
	require.NoError(t, err)
}

func testProcessingCloudFile(t *testing.T, loader *StoreLoader, testDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fPath := filepath.Join(testDir, "test.json")
	err := loader.ProcessFile(ctx, fPath)
	require.NoError(t, err)
}

func createJSONFile(dir, name string) (string, error) {
	fPath := fmt.Sprintf("%s.json", name)
	if dir != "" {
		fPath = fmt.Sprintf("%s/%s", dir, fPath)
	}
	items := []Result{
		{
			"city":      "Hong Kong",
			"name":      "Plaza Hollywood",
			"country":   "CN",
			"longitude": 114.20169067382812,
			"latitude":  22.340700149536133,
			"store_id":  1,
		},
		{
			"city":      "Hong Kong",
			"name":      "Exchange Square",
			"country":   "CN",
			"longitude": 114.15818786621094,
			"latitude":  22.283939361572266,
			"store_id":  6,
		},
		{
			"city":      "Kowloon",
			"name":      "Telford Plaza",
			"country":   "CN",
			"longitude": 114.21343994140625,
			"latitude":  22.3228702545166,
			"store_id":  8,
		},
	}

	err := file.CreateDirectory(fPath)
	if err != nil {
		return "", err
	}

	f, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	err = encoder.Encode(items)
	if err != nil {
		return "", err
	}
	return fPath, nil
}
