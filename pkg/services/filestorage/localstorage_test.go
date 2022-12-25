package filestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/utils/file"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestLocalFileStorage(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client *LocalStorageClient,
		testDir string,
	){
		"local storage file read succeeds":                   testReadFileArray,
		"local storage file read missing token throws error": testReadFileArrayMissingTokens,
		"local storage copy json file succeeds":              testCopy,
		"local storage copy json file buffered succeeds":     testCopyBuffer,
		"file stats test succeeds":                           testFileStats,
		// "local storage copy json file buffered large succeeds": testCopyLargeBuffer,
		// "local storage file write succeeds":                    testWrite,
		// "read write file array succeeds":                       testReadWriteFileArray,
		// "local storage across folders succeds":                 testCopyAcross,
	} {
		testDir := "testing_local_storage/"
		t.Run(scenario, func(t *testing.T) {
			client, teardown := setupLocalTest(t, testDir)
			defer teardown()
			fn(t, client, testDir)
		})
	}
}

func setupLocalTest(t *testing.T, testDir string) (
	client *LocalStorageClient,
	teardown func(),
) {
	t.Helper()

	err := file.CreateDirectory(testDir)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)

	lsc, err := NewLocalStorageClient(appLogger)
	require.NoError(t, err)

	return lsc, func() {
		t.Logf(" test ended, will remove %s folder", testDir)
		err := os.RemoveAll(testDir)
		require.NoError(t, err)
	}
}

func testReadFileArray(t *testing.T, client *LocalStorageClient, testDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	errCount := 0
	defer func() {
		require.Equal(t, 3, count)
		require.Equal(t, 0, errCount)
		cancel()
	}()

	name := "data"
	fPath, err := createJSONFile(testDir, name)
	require.NoError(t, err)

	resultStream, err := client.ReadFileArray(ctx, cancel, fPath)
	require.NoError(t, err)

	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-resultStream:
			if !ok {
				t.Log("	testReadFileArray: resultstream closed, returning")
				return
			} else {
				if r.Result != nil {
					t.Logf(" testReadFileArray: result: %v", r.Result)
					count++
				}
				if r.Error != nil {
					t.Logf(" testReadFileArray: error: %v", r.Error)
					errCount++
				}
			}
		}
	}
}

func testReadFileArrayMissingTokens(t *testing.T, client *LocalStorageClient, testDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	errCount := 0
	defer func() {
		require.Equal(t, 0, count)
		require.Equal(t, 1, errCount)
		cancel()
	}()

	name := "data_missing_tokens"
	fPath, err := createSingleJSONFile(testDir, name)
	require.NoError(t, err)

	resultStream, err := client.ReadFileArray(ctx, cancel, fPath)
	require.NoError(t, err)

	for {
		select {
		case <-ctx.Done():
			return
		case r, ok := <-resultStream:
			if !ok {
				t.Log("	testReadFileArrayMissingTokens: resultstream closed, returning")
				return
			} else {
				if r.Result != nil {
					t.Logf(" testReadFileArrayMissingTokens: result: %v", r.Result)
					count++
				}
				if r.Error != nil {
					t.Logf(" testReadFileArrayMissingTokens: error: %v", r.Error)
					errCount++
				}
			}
		}
	}
}

func testCopy(t *testing.T, client *LocalStorageClient, testDir string) {
	name := "test"
	srcName, err := createJSONFile(testDir, name)
	require.NoError(t, err)

	destName := fmt.Sprintf("%s/%s-copy.json", testDir, name)
	n, err := client.Copy(srcName, destName)
	require.NoError(t, err)
	t.Logf(" testCopy: %d bytes written", n)
	require.Equal(t, true, n > 0)

}

func testCopyBuffer(t *testing.T, client *LocalStorageClient, testDir string) {
	name := "test"
	srcName, err := createJSONFile(testDir, name)
	require.NoError(t, err)

	destName := fmt.Sprintf("%s/%s-copy-buf.json", testDir, name)
	n, err := client.CopyBuf(srcName, destName)
	require.NoError(t, err)
	t.Logf(" testCopy: %d bytes written", n)
	require.Equal(t, true, n > 0)
}

// func testCopyLargeBuffer(t *testing.T, client *LocalStorageClient, testDir string) {
// 	name := "starbucks"
// 	n, err := client.CopyBuf(fmt.Sprintf("data/%s.json", name), fmt.Sprintf("%s/%s-copy-buf.json", testDir, name))
// 	require.NoError(t, err)
// 	t.Logf(" testCopy: %d bytes written", n)
// 	require.Equal(t, true, n > 0)
// }

func testFileStats(t *testing.T, client *LocalStorageClient, testDir string) {
	name := "test"
	fPath, err := createJSONFile(testDir, name)
	require.NoError(t, err)

	err = FileStats(fPath)
	require.NoError(t, err)
}

// func testCopyAcross(t *testing.T, client *LocalStorageClient, testDir string) {
// 	dir, err := os.Getwd()
// 	require.NoError(t, err)
// 	t.Logf(" testCopyAcross: working directory: %s", dir)

// 	srcPath := filepath.Join(dir, "../../../cmd/store/data/starbucks.json")
// 	t.Logf(" testCopyAcross: srcPath: %s", srcPath)

// 	destPath := filepath.Join(dir, "../../jobs/data/starbucks.json")
// 	t.Logf(" testCopyAcross: destPath: %s", destPath)

// 	n, err := client.Copy(srcPath, destPath)
// 	require.NoError(t, err)
// 	t.Logf(" testCopyAcross: %d bytes written", n)
// 	require.Equal(t, true, n > 0)

// }

// func testReadWriteFileArray(t *testing.T, client *LocalStorageClient, testDir string) {
// 	rCtx, rCancel := context.WithCancel(context.Background())
// 	defer func() {
// 		rCancel()
// 	}()

// 	name := "data"
// 	fPath, err := createJSONFile(testDir, name)
// 	require.NoError(t, err)

// 	resultStream, err := client.ReadFileArray(rCtx, rCancel, fPath)
// 	require.NoError(t, err)

// 	wCtx, wCancel := context.WithCancel(context.Background())
// 	defer func() {
// 		wCancel()
// 	}()
// 	writeFileName := "dataWrite.json"
// 	requestStream := make(chan Result)
// 	respStream := client.WriteFile(wCtx, wCancel, writeFileName, requestStream)

// 	func() {
// 		defer close(requestStream)
// 		for {
// 			select {
// 			case <-rCtx.Done():
// 				t.Log("rCtx done")
// 			case r, ok := <-resultStream:
// 				if !ok {
// 					t.Log(" testReadWriteFileArray: resultstream closed, returning")
// 					return
// 				} else {
// 					if r.Result != nil {
// 						t.Logf(" testReadWriteFileArray: result: %v", r.Result)
// 						requestStream <- r.Result
// 					}
// 					if r.Error != nil {
// 						t.Logf(" testReadWriteFileArray: error: %v", r.Error)
// 					}
// 				}
// 			}
// 		}
// 	}()

// 	func() {
// 		for {
// 			select {
// 			case <-wCtx.Done():
// 				t.Log("wCtx done, returning")
// 				return
// 			case r, ok := <-respStream:
// 				if !ok {
// 					t.Log("WriteFile resultstream closed, returning")
// 					return
// 				} else {
// 					if r.Error != nil {
// 						t.Logf("WriteFile error: %v", r.Error)
// 					}
// 				}
// 			}
// 		}
// 	}()
// }

func createSingleJSONFile(dir, name string) (string, error) {
	fPath := fmt.Sprintf("%s.json", name)
	if dir != "" {
		fPath = fmt.Sprintf("%s/%s", dir, fPath)
	}
	item := Result{
		"city":      "Hong Kong",
		"name":      "Plaza Hollywood",
		"country":   "CN",
		"longitude": 114.20169067382812,
		"latitude":  22.340700149536133,
		"store_id":  1,
	}

	file, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(item)
	if err != nil {
		return "", err
	}
	return fPath, nil
}
