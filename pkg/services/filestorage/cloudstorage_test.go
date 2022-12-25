package filestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/utils/file"
)

func TestCloudFileStorage(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client *CloudStorageClient,
		testDir string,
	){
		"cloud storage file upload succeeds":   testUpload,
		"cloud storage file download succeeds": testDownload,
		"list objects succeeds":                testListObjects,
		"delete cloud bucket objects succeeds": testDeleteObjects,
		// "cloud storage file upload across folders succeeds": testUploadAcross,
	} {
		testDir := "testing_cloud_storage/"
		t.Run(scenario, func(t *testing.T) {
			client, teardown := setupCloudTest(t, testDir)
			defer teardown()
			fn(t, client, testDir)
		})
	}
}

func setupCloudTest(t *testing.T, testDir string) (
	client *CloudStorageClient,
	teardown func(),
) {
	t.Helper()

	err := file.CreateDirectory(testDir)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)

	appCfg, err := config.GetAppConfig(appLogger, "test-config.json")
	require.NoError(t, err)

	cscCfg := appCfg.Services.CloudStorageCfg
	fsc, err := NewCloudStorageClient(appLogger, cscCfg)
	require.NoError(t, err)

	return fsc, func() {
		t.Logf(" test ended, will remove %s folder", testDir)
		err := os.RemoveAll(testDir)
		require.NoError(t, err)
	}
}

func testUpload(t *testing.T, client *CloudStorageClient, testDir string) {
	name := "test"
	filePath, err := createJSONFile(testDir, name)
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

	n, err := client.UploadFile(ctx, file, destName, testDir)
	require.NoError(t, err)
	t.Logf(" testUpload: %d bytes written", n)
	require.Equal(t, true, n > 0)
}

func testDownload(t *testing.T, client *CloudStorageClient, testDir string) {
	name := "test"
	filePath, err := createJSONFile(testDir, name)
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

	n, err := client.UploadFile(ctx, file, destName, testDir)
	require.NoError(t, err)
	t.Logf(" testUpload: %d bytes written", n)
	require.Equal(t, true, n > 0)

	localFilePath := filepath.Join(testDir, fmt.Sprintf("%s-copy.json", name))
	lFile, err := os.Create(localFilePath)
	require.NoError(t, err)
	defer func() {
		err := lFile.Close()
		require.NoError(t, err)
	}()

	n, err = client.DownloadFile(ctx, lFile, destName, testDir)
	require.NoError(t, err)
	t.Logf(" testDownload: %d bytes written to file %s", n, localFilePath)
	require.Equal(t, true, n > 0)
}

func testListObjects(t *testing.T, client *CloudStorageClient, testDir string) {
	name := "test"
	filePath, err := createJSONFile(testDir, name)
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

	n, err := client.UploadFile(ctx, file, destName, testDir)
	require.NoError(t, err)
	t.Logf(" testUpload: %d bytes written", n)
	require.Equal(t, true, n > 0)

	names, err := client.ListObjects(ctx)
	require.NoError(t, err)
	require.Equal(t, true, len(names) > 0)
}

func testDeleteObjects(t *testing.T, client *CloudStorageClient, testDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := client.DeleteObjects(ctx)
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

	file, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(items)
	if err != nil {
		return "", err
	}
	return fPath, nil
}
