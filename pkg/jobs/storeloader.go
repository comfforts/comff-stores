package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/constants"
	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/filestorage"
	"github.com/comfforts/comff-stores/pkg/services/store"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
	"go.uber.org/zap"
)

const (
	ERROR_CREATING_STORAGE_CLIENT string = "error creating storage client"
	ERROR_UPLOADING_FILE          string = "uploading file %s"
	ERROR_CLOSING_FILE            string = "closing file %s"
	ERROR_NO_FILE                 string = "%s doesn't exist"
	ERROR_CREATING_FILE           string = "creating file %s"
	ERROR_DOWNLOADING_FILE        string = "downloading file %s"
	ERROR_CREATING_DATA_DIR       string = "creating data directory %s"
)

type Result = map[string]interface{}

type StoreLoader struct {
	logger       *logging.AppLogger
	stores       *store.StoreService
	config       config.StoreLoaderConfig
	localStorage *filestorage.LocalStorageClient
	cloudStorage *filestorage.CloudStorageClient
}

func NewStoreLoader(cfg config.StoreLoaderConfig, ss *store.StoreService, l *logging.AppLogger) (*StoreLoader, error) {
	if ss == nil || l == nil || cfg.CloudStorageCfg.BucketName == "" || cfg.CloudStorageCfg.CredsPath == "" {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}

	storeLoader := StoreLoader{
		logger: l,
		stores: ss,
		config: cfg,
	}

	ll, err := filestorage.NewLocalStorageClient(l)
	if err != nil {
		return nil, err
	}
	storeLoader.localStorage = ll

	cl, err := filestorage.NewCloudStorageClient(l, cfg.CloudStorageCfg)
	if err != nil {
		return nil, err
	}
	storeLoader.cloudStorage = cl
	return &storeLoader, nil
}

func (jd *StoreLoader) ProcessFile(ctx context.Context, filePath string) error {
	err := jd.processingPreCheck(ctx, filePath)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, constants.FilePathContextKey, filePath)
	ctx = context.WithValue(ctx, constants.ReadRateContextKey, 2)
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		jd.logger.Info("finished setting up store data")
		cancel()
	}()

	storeQueue := make(chan *store.Store)
	var fileProcessingWaitGrp sync.WaitGroup
	var storeUpdateWaitGrp sync.WaitGroup

	fileProcessingWaitGrp.Add(1)
	go jd.queueFileStreaming(ctx, cancel, &fileProcessingWaitGrp, &storeUpdateWaitGrp, storeQueue)

	fileProcessingWaitGrp.Add(1)
	go jd.queueStoreUpdate(ctx, cancel, &fileProcessingWaitGrp, &storeUpdateWaitGrp, storeQueue)

	fileProcessingWaitGrp.Wait()

	jd.publishStats(ctx)
	return nil
}

func (jd *StoreLoader) processingPreCheck(ctx context.Context, filePath string) error {
	// check if file exists
	err := filestorage.FileStats(filePath)
	if err != nil {
		// check if file path exists in data folder
		dataPath := fmt.Sprintf("%s/%s", jd.config.DataPath, filePath)
		err := filestorage.FileStats(dataPath)
		if err != nil {
			// try fetching from cloud storage
			err = jd.getFromStorage(ctx, dataPath)
			if err != nil {
				return err
			}
		}
	}

	dataVol := fileUtils.GetRoot(filePath)
	// if path root folder is not loader's data folder
	if dataVol != jd.config.DataPath {
		dataPath := fmt.Sprintf("%s/%s", jd.config.DataPath, filePath)
		err := filestorage.FileStats(dataPath)
		// if file path doesn't exist in data folder, copy file path to data folder
		if err != nil {
			jd.logger.Info("copying file", zap.String("filePath", filePath), zap.String("dataPath", dataPath))
			err = fileUtils.CreateDirectory(dataPath)
			if err != nil {
				jd.logger.Error("error creating data directory", zap.Error(err))
				err = errors.WrapError(err, ERROR_CREATING_DATA_DIR, dataPath)
				return err
			}
			n, err := jd.localStorage.Copy(filePath, dataPath)
			if err != nil {
				return err
			}
			jd.logger.Info("copied file", zap.String("filePath", filePath), zap.String("dataPath", dataPath), zap.Int64("bytes", n))
		}
	}
	return nil
}

func (jd *StoreLoader) queueFileStreaming(
	ctx context.Context,
	cancel func(),
	fileProcessingWaitGrp *sync.WaitGroup,
	storeUpdateWaitGrp *sync.WaitGroup,
	storeQueue chan *store.Store,
) {
	jd.logger.Info("start reading store data file")
	fp := ctx.Value(constants.FilePathContextKey).(string)
	resultStream, err := jd.localStorage.ReadFileArray(ctx, cancel, filepath.Join(jd.config.DataPath, fp))
	if err != nil {
		jd.logger.Error("error reading store data file", zap.Error(err))
		cancel()
	}

	for {
		select {
		case <-ctx.Done():
			jd.logger.Info("store data file read context done")
			return
		case r, ok := <-resultStream:
			if !ok {
				jd.logger.Info("store data file result stream closed")
				fileProcessingWaitGrp.Done()
				close(storeQueue)
				return
			}
			storeUpdateWaitGrp.Add(1)
			if r.Result != nil {
				jd.queueStoreProcessing(r.Result, storeQueue, cancel)
			}
			if r.Error != nil {
				jd.logger.Info("store data file result stream closed")
			}
		}
	}
}

func (jd *StoreLoader) queueStoreProcessing(
	r map[string]interface{},
	storeQueue chan *store.Store,
	cancel func(),
) {
	store, err := store.MapResultToStore(r)
	if err != nil {
		jd.logger.Error("error processing store data", zap.Error(err), zap.Any("storeJson", r))
	} else {
		storeQueue <- store
	}
}

func (jd *StoreLoader) queueStoreUpdate(
	ctx context.Context,
	cancel func(),
	fileProcessingWaitGrp *sync.WaitGroup,
	storeUpdateWaitGrp *sync.WaitGroup,
	storeQueue chan *store.Store,
) {
	jd.logger.Info("start updating store data")

	for {
		select {
		case <-ctx.Done():
			jd.logger.Info("store data update context done")
			return
		case store, ok := <-storeQueue:
			if !ok {
				jd.logger.Info("store notification channel closed")
				fileProcessingWaitGrp.Done()
				return
			}
			addedStr, err := jd.addStore(ctx, store)
			if addedStr == nil || err != nil {
				jd.logger.Error("error processing store data", zap.Error(err), zap.Any("store", store))
			}
			storeUpdateWaitGrp.Done()
		}
	}
}

func (jd *StoreLoader) addStore(ctx context.Context, s *store.Store) (*store.Store, error) {
	if s.Org == "" {
		fileName := filepath.Base(ctx.Value(constants.FilePathContextKey).(string))
		s.Org = fileName[0:strings.Index(fileName, ".")]
	}
	return jd.stores.AddStore(ctx, s)
}

func (jd *StoreLoader) publishStats(ctx context.Context) {
	jd.stores.SetReady(ctx, true)
	stats := jd.stores.GetStoreStats()
	jd.logger.Info("gateway status", zap.Any("stats", stats))
}

func (jd *StoreLoader) getFromStorage(ctx context.Context, filePath string) error {
	err := fileUtils.CreateDirectory(filePath)
	if err != nil {
		jd.logger.Error("error creating data directory", zap.Error(err))
		err = errors.WrapError(err, ERROR_CREATING_DATA_DIR, filePath)
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return errors.WrapError(err, ERROR_CREATING_FILE, filePath)
	}
	_, err = jd.cloudStorage.DownloadFile(ctx, f, filepath.Base(filePath), filepath.Dir(filePath))
	if err != nil {
		return err
	}
	err = filestorage.FileStats(filePath)
	if err != nil {
		return err
	}
	return nil
}

func (jd *StoreLoader) StoreDataFile(ctx context.Context, filePath string) error {
	stores := jd.stores.GetAllStores()
	data := []Result{}
	for _, v := range stores {
		data = append(data, Result{
			"store_id":  v.StoreId,
			"city":      v.City,
			"name":      v.Name,
			"country":   v.Country,
			"longitude": v.Longitude,
			"latitude":  v.Latitude,
		})
	}

	dataPath := filepath.Join(jd.config.DataPath, filePath)
	err := fileUtils.CreateDirectory(dataPath)
	if err != nil {
		jd.logger.Error("error creating data directory", zap.Error(err))
		err = errors.WrapError(err, ERROR_CREATING_DATA_DIR, filePath)
		return err
	}

	f, err := os.Create(dataPath)
	if err != nil {
		jd.logger.Error("error creating file", zap.Error(err), zap.String("filepath", filePath))
		return errors.WrapError(err, ERROR_CREATING_FILE, filePath)
	}
	defer func() {
		if err := f.Close(); err != nil {
			jd.logger.Error("error closing file", zap.Error(err), zap.String("filepath", filePath))
		}
	}()

	enc := json.NewEncoder(f)
	err = enc.Encode(data)
	if err != nil {
		jd.logger.Error("error saving data file", zap.Error(err), zap.String("filepath", filePath))
		return err
	}
	return nil
}

func (jd *StoreLoader) UploadDataFile(ctx context.Context, filePath string) error {
	dataPath := filepath.Join(jd.config.DataPath, filePath)
	file, err := os.Open(dataPath)
	if err != nil {
		jd.logger.Error("error accessing file", zap.Error(err))
		return errors.WrapError(err, ERROR_NO_FILE, filePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			jd.logger.Error("error closing file", zap.Error(err), zap.String("filepath", filePath))
		}
	}()

	n, err := jd.cloudStorage.UploadFile(ctx, file, filepath.Base(dataPath), filepath.Dir(dataPath))
	if err != nil {
		jd.logger.Error("error uploading file", zap.Error(err))
		return err
	}
	jd.logger.Info("uploaded file", zap.String("file", filepath.Base(dataPath)), zap.String("path", filepath.Dir(dataPath)), zap.Int64("bytes", n))
	return nil
}
