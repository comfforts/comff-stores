package jobs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/comfforts/cloudstorage"
	"github.com/comfforts/logger"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/services/store"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
	testUtils "github.com/comfforts/comff-stores/pkg/utils/test"
	"github.com/stretchr/testify/require"
)

const TEST_DIR = "test-data"

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
		testDir := TEST_DIR
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
	err := fileUtils.CreateDirectory(fmt.Sprintf("%s/", testDir))
	require.NoError(t, err)

	appLogger := logger.NewTestAppLogger(TEST_DIR)
	ss := store.NewStoreService(appLogger)

	t.Logf(" setupStoreLoader: getting app-config/loader-config %s", "test-config.json")
	appCfg, err := config.GetAppConfig("test-config.json", appLogger)
	require.NoError(t, err)

	cscCfg := cloudstorage.CloudStorageClientConfig{
		CredsPath: appCfg.Services.CloudStorageClientCfg.CredsPath,
	}
	csc, err := cloudstorage.NewCloudStorageClient(cscCfg, appLogger)
	require.NoError(t, err)

	slCfg := appCfg.Jobs.StoreLoaderConfig
	slCfg.DataDir = TEST_DIR
	sl, err = NewStoreLoader(slCfg, ss, csc, appLogger)
	require.NoError(t, err)

	return sl, func() {
		t.Logf(" TestStoreService ended, will clear store data")
		sl.stores.Close()

		err := os.RemoveAll(testDir)
		require.NoError(t, err)

		err = os.RemoveAll(slCfg.DataDir)
		require.NoError(t, err)
	}
}

func testFileStats(t *testing.T, loader *StoreLoader, testDir string) {
	name := "test"
	fPath, err := testUtils.CreateJSONFile(testDir, name)
	require.NoError(t, err)

	require.Equal(t, ".json", filepath.Ext(fPath))
	require.Equal(t, "test.json", filepath.Base(fPath))

	dir, err := os.Getwd()
	require.NoError(t, err)
	t.Logf("working directory: %s", dir)

	testPath := filepath.Join(dir, "../../cmd/store")
	t.Logf("testPath: %s", testPath)
}

func testSaveDataFile(t *testing.T, loader *StoreLoader, testDir string) {
	name := "test"
	fPath, err := testUtils.CreateJSONFile(testDir, name)
	require.NoError(t, err)

	fileName := filepath.Base(fPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = loader.ProcessFile(ctx, fileName)
	require.NoError(t, err)

	err = loader.StoreDataFile(ctx, fileName)
	require.NoError(t, err)
}

func testUploadDataFile(t *testing.T, loader *StoreLoader, testDir string) {
	name := "test"
	fPath, err := testUtils.CreateJSONFile(testDir, name)
	require.NoError(t, err)

	fileName := filepath.Base(fPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = loader.StoreDataFile(ctx, fileName)
	require.NoError(t, err)

	err = loader.UploadDataFile(ctx, fileName)
	require.NoError(t, err)
}

func testProcessingLocalFile(t *testing.T, loader *StoreLoader, testDir string) {
	name := "test"
	fPath, err := testUtils.CreateJSONFile(testDir, name)
	require.NoError(t, err)

	fileName := filepath.Base(fPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = loader.ProcessFile(ctx, fileName)
	require.NoError(t, err)
}

func testProcessingCloudFile(t *testing.T, loader *StoreLoader, testDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fPath := filepath.Join(testDir, "test.json")

	fileName := filepath.Base(fPath)

	err := loader.ProcessFile(ctx, fileName)
	require.NoError(t, err)
}
