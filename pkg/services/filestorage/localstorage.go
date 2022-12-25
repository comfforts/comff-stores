package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

const (
	ERROR_OPENING_FILE      string = "opening file %s"
	ERROR_DECODING_RESULT   string = "error decoding result json"
	ERROR_FILE_INACCESSIBLE string = "%s inaccessible"
	ERROR_START_TOKEN       string = "error reading start token"
	ERROR_END_TOKEN         string = "error reading end token"
	ERROR_READING_FILE      string = "reading file %s"
	ERROR_NOT_A_FILE        string = "%s not a file"
	ERROR_FILE_STAT         string = "fetching stats for %s"
	ERROR_WRITING_FILE      string = "writing file %s"
)

const DEFAULT_BUFFER_SIZE = 1000

var (
	ErrStartToken = errors.NewAppError(ERROR_START_TOKEN)
	ErrEndToken   = errors.NewAppError(ERROR_END_TOKEN)
)

type Result = map[string]interface{}

type ReadResponse struct {
	Result Result
	Error  error
}

type WriteResponse struct {
	Error error
}

type LocalStorageClient struct {
	logger *logging.AppLogger
}

func NewLocalStorageClient(logger *logging.AppLogger) (*LocalStorageClient, error) {
	if logger == nil {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}
	loaderClient := &LocalStorageClient{
		logger: logger,
	}

	return loaderClient, nil
}

// ReadFileArray reads an array of json data from existing file, one by one,
// and returns individual result at defined rate through returned channel
func (lc *LocalStorageClient) ReadFileArray(ctx context.Context, cancel func(), filePath string) (<-chan ReadResponse, error) {
	// checks if file exists
	err := FileStats(filePath)
	if err != nil {
		return nil, err
	}

	// Open file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.WrapError(err, ERROR_OPENING_FILE, filePath)
	}

	resultStream := make(chan ReadResponse)
	go lc.readFile(ctx, cancel, filePath, f, resultStream)

	return resultStream, nil
}

func (lc *LocalStorageClient) WriteFile(ctx context.Context, cancel func(), fileName string, reqStream chan Result) <-chan WriteResponse {
	filePath := filepath.Join("data", fileName)

	resultStream := make(chan WriteResponse)
	go lc.writeFile(ctx, cancel, filePath, reqStream, resultStream)

	return resultStream
}

func (lc *LocalStorageClient) Copy(srcPath, destPath string) (int64, error) {
	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_FILE_STAT, srcPath)
	}
	if !srcStat.Mode().IsRegular() {
		return 0, errors.WrapError(err, ERROR_NOT_A_FILE, srcPath)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_OPENING_FILE, srcPath)
	}
	defer src.Close()

	err = fileUtils.CreateDirectory(destPath)
	if err != nil {
		return 0, err
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_CREATING_FILE, destPath)
	}
	defer dest.Close()

	nBytes, err := io.Copy(dest, src)
	return nBytes, err
}

func (lc *LocalStorageClient) CopyBuf(srcPath, destPath string) (int64, error) {
	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_FILE_STAT, srcPath)
	}
	if !srcStat.Mode().IsRegular() {
		return 0, errors.WrapError(err, ERROR_NOT_A_FILE, srcPath)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_OPENING_FILE, srcPath)
	}
	defer src.Close()

	err = fileUtils.CreateDirectory(destPath)
	if err != nil {
		return 0, err
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_CREATING_FILE, destPath)
	}
	defer dest.Close()

	buf := make([]byte, DEFAULT_BUFFER_SIZE)
	var nBytes int64 = 0
	for {
		nr, err := src.Read(buf)
		if err != nil && err != io.EOF {
			return nBytes, errors.WrapError(err, ERROR_READING_FILE, srcPath)
		}
		if nr == 0 {
			break
		}
		nw, err := dest.Write(buf[:nr])
		if err != nil {
			return nBytes, errors.WrapError(err, ERROR_WRITING_FILE, srcPath)
		}
		nBytes = nBytes + int64(nw)
	}

	return nBytes, err
}

func (lc *LocalStorageClient) readFile(ctx context.Context, cancel func(), filePath string, file io.ReadCloser, rrs chan ReadResponse) {
	defer close(rrs)
	defer func() {
		lc.logger.Info("closing result stream and file")
		if err := file.Close(); err != nil {
			rrs <- ReadResponse{
				Error: errors.WrapError(err, ERROR_CLOSING_FILE, filePath),
			}
		}
	}()

	r := bufio.NewReader(file)
	dec := json.NewDecoder(r)

	// read open bracket
	t, err := dec.Token()
	if err != nil || t != json.Delim('[') {
		rrs <- ReadResponse{
			Error: ErrStartToken,
		}
		cancel()
		return
	}

	// while the array contains values
	for dec.More() {
		var result Result
		err := dec.Decode(&result)
		var response = ReadResponse{}
		if err != nil {
			response.Error = errors.WrapError(err, ERROR_DECODING_RESULT)
		} else {
			response.Result = result
		}
		select {
		case <-ctx.Done():
			return
		case rrs <- response:
		}
	}

	// read closing bracket
	t, err = dec.Token()
	if err != nil || t != json.Delim(']') {
		rrs <- ReadResponse{
			Error: ErrEndToken,
		}
		cancel()
		return
	}
}

func (lc *LocalStorageClient) writeFile(ctx context.Context, cancel func(), filePath string, reqStream chan Result, wrs chan WriteResponse) {
	defer func() {
		lc.logger.Info("closing write response stream")
		close(wrs)
	}()
	file, err := os.OpenFile(filePath, os.O_CREATE, os.ModePerm)
	if err != nil {
		wrs <- WriteResponse{
			Error: errors.WrapError(err, ERROR_CREATING_FILE, filePath),
		}
		cancel()
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			wrs <- WriteResponse{
				Error: errors.WrapError(err, ERROR_CLOSING_FILE, filePath),
			}
		}
	}()

	jsonData := []Result{}
	for req := range reqStream {
		jsonData = append(jsonData, req)
	}

	enc := json.NewEncoder(file)
	err = enc.Encode(jsonData)
	if err != nil {
		wrs <- WriteResponse{
			Error: errors.WrapError(err, ERROR_CREATING_FILE, filePath),
		}
		cancel()
		return
	}
}

// checks if file exists
func FileStats(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.WrapError(err, ERROR_NO_FILE, filePath)
		} else {
			return errors.WrapError(err, ERROR_FILE_INACCESSIBLE, filePath)
		}
	}
	return nil
}
