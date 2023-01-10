package filestorage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/logging"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
	testUtils "github.com/comfforts/comff-stores/pkg/utils/test"
)

const TEST_DIR = "test-data"

type TestConfig struct {
	dir    string
	bucket string
}

func TestCloudFileStorage(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client *CloudStorageClient,
		testCfg TestConfig,
	){
		"cloud storage file upload succeeds":   testUpload,
		"cloud storage file download succeeds": testDownload,
		"list objects succeeds":                testListObjects,
		"delete cloud bucket objects succeeds": testDeleteObjects,
		// "cloud storage file upload across folders succeeds": testUploadAcross,
	} {
		testCfg := TestConfig{
			dir:    "cloud_storage_test/",
			bucket: "comfforts-playground",
		}

		t.Run(scenario, func(t *testing.T) {
			client, teardown := setupCloudTest(t, testCfg)
			defer teardown()
			fn(t, client, testCfg)
		})
	}
}

func setupCloudTest(t *testing.T, testCfg TestConfig) (
	client *CloudStorageClient,
	teardown func(),
) {
	t.Helper()

	err := fileUtils.CreateDirectory(testCfg.dir)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)

	appCfg, err := config.GetAppConfig("test-config.json", appLogger)
	require.NoError(t, err)

	cscCfg := appCfg.Services.CloudStorageClientCfg
	fsc, err := NewCloudStorageClient(cscCfg, appLogger)
	require.NoError(t, err)

	return fsc, func() {
		t.Logf(" test ended, will remove %s folder", testCfg.dir)
		// err := os.RemoveAll(testCfg.dir)
		// require.NoError(t, err)
	}
}

func testUpload(t *testing.T, client *CloudStorageClient, testCfg TestConfig) {
	name := "test"
	filePath, err := testUtils.CreateJSONFile(testCfg.dir, name)
	require.NoError(t, err)

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer func() {
		err := file.Close()
		require.NoError(t, err)
	}()

	destName := fmt.Sprintf("%s.json", name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfr, err := NewCloudFileRequest(testCfg.bucket, destName, testCfg.dir, 0)
	require.NoError(t, err)

	n, err := client.UploadFile(ctx, file, cfr)
	require.NoError(t, err)
	t.Logf(" testUpload: %d bytes written", n)
	require.Equal(t, true, n > 0)
}

func testDownload(t *testing.T, client *CloudStorageClient, testCfg TestConfig) {
	name := "test"
	filePath, err := testUtils.CreateJSONFile(testCfg.dir, name)
	require.NoError(t, err)

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer func() {
		err := file.Close()
		require.NoError(t, err)
	}()

	destName := fmt.Sprintf("%s.json", name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfr, err := NewCloudFileRequest(testCfg.bucket, destName, testCfg.dir, 0)
	require.NoError(t, err)

	n, err := client.UploadFile(ctx, file, cfr)
	require.NoError(t, err)
	t.Logf(" testUpload: %d bytes written", n)
	require.Equal(t, true, n > 0)

	localFilePath := filepath.Join(testCfg.dir, fmt.Sprintf("%s-copy.json", name))
	lFile, err := os.Create(localFilePath)
	require.NoError(t, err)
	defer func() {
		err := lFile.Close()
		require.NoError(t, err)
	}()

	n, err = client.DownloadFile(ctx, lFile, cfr)
	require.NoError(t, err)
	t.Logf(" testDownload: %d bytes written to file %s", n, localFilePath)
	require.Equal(t, true, n > 0)
}

func testListObjects(t *testing.T, client *CloudStorageClient, testCfg TestConfig) {
	name := "test"
	filePath, err := testUtils.CreateJSONFile(testCfg.dir, name)
	require.NoError(t, err)

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer func() {
		err := file.Close()
		require.NoError(t, err)
	}()

	destName := fmt.Sprintf("%s.json", name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfr, err := NewCloudFileRequest(testCfg.bucket, destName, testCfg.dir, 0)
	require.NoError(t, err)

	n, err := client.UploadFile(ctx, file, cfr)
	require.NoError(t, err)
	t.Logf(" testUpload: %d bytes written", n)
	require.Equal(t, true, n > 0)

	names, err := client.ListObjects(ctx, cfr)
	require.NoError(t, err)
	require.Equal(t, true, len(names) > 0)
}

func testDeleteObjects(t *testing.T, client *CloudStorageClient, testCfg TestConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfr, err := NewCloudFileRequest(testCfg.bucket, "", "", 0)
	require.NoError(t, err)

	err = client.DeleteObjects(ctx, cfr)
	require.NoError(t, err)
}

// func testUploadAcross(t *testing.T, client *CloudStorageClient, testDir string) {
// 	dir, err := os.Getwd()
// 	require.NoError(t, err)
// 	t.Logf(" testUploadAcross: working directory: %s", dir)

// 	srcPath := filepath.Join(dir, "../../../cmd/store/data/starbucks.json")
// 	t.Logf(" testUploadAcross: srcPath: %s", srcPath)

// 	file, err := os.Open(srcPath)
// 	require.NoError(t, err)
// 	defer func() {
// 		err := file.Close()
// 		require.NoError(t, err)
// 	}()

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	n, err := client.UploadFile(ctx, file, filepath.Base(srcPath), "data")
// 	require.NoError(t, err)
// 	t.Logf(" testUploadAcross: %d bytes written", n)
// 	require.Equal(t, true, n > 0)
// }
