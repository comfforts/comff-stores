package jobs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/comfforts/cloudstorage"
	"github.com/comfforts/errors"
	"github.com/comfforts/localstorage"
	"github.com/comfforts/logger"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/constants"
	storeModels "github.com/comfforts/comff-stores/pkg/models/store"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
	"go.uber.org/zap"
)

type StoreLoader struct {
	logger       logger.AppLogger
	stores       storeModels.Stores
	config       config.StoreLoaderConfig
	localStorage localstorage.LocalStorage
	cloudStorage cloudstorage.CloudStorage
}

func NewStoreLoader(cfg config.StoreLoaderConfig, ss storeModels.Stores, csc cloudstorage.CloudStorage, l logger.AppLogger) (*StoreLoader, error) {
	if ss == nil || l == nil || cfg.BucketName == "" {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}

	storeLoader := StoreLoader{
		logger:       l,
		stores:       ss,
		config:       cfg,
		cloudStorage: csc,
	}

	ll, err := localstorage.NewLocalStorageClient(l)
	if err != nil {
		return nil, err
	}
	storeLoader.localStorage = ll

	return &storeLoader, nil
}

// filePath can be either "<path/to/localFile.json>" string or
// GCP storage bucket "uploadPath/storedFile.json" string
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

	storeQueue := make(chan *storeModels.Store)
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
	dataPath := fmt.Sprintf("%s/%s", jd.config.DataDir, filePath)

	// check if file exists
	if _, err := fileUtils.FileStats(filePath); err != nil {
		// check if file exists in store
		_, err := fileUtils.FileStats(dataPath)
		if err != nil {
			// try fetching from cloud storage
			err = jd.getFromStorage(ctx, dataPath)
			if err != nil {
				return err
			}
		}
	} else {
		dataVol := fileUtils.GetRoot(filePath)
		// if file is not placed in file store, copy file to file store
		if dataVol != jd.config.DataDir {
			dataPath := fmt.Sprintf("%s/%s", jd.config.DataDir, filePath)
			_, err := fileUtils.FileStats(dataPath)
			if err != nil {
				jd.logger.Info("copying file", zap.String("src", filePath), zap.String("dest", dataPath))
				err = fileUtils.CreateDirectory(dataPath)
				if err != nil {
					jd.logger.Error("error creating data directory", zap.Error(err))
					err = errors.WrapError(err, fileUtils.ERROR_CREATING_DATA_DIR, dataPath)
					return err
				}
				n, err := jd.localStorage.Copy(filePath, dataPath)
				if err != nil {
					return err
				}
				jd.logger.Info("copied file", zap.String("src", filePath), zap.String("dest", dataPath), zap.Int64("bytes", n))
			}
		}
	}
	return nil
}

func (jd *StoreLoader) queueFileStreaming(
	ctx context.Context,
	cancel func(),
	fileProcessingWaitGrp *sync.WaitGroup,
	storeUpdateWaitGrp *sync.WaitGroup,
	storeQueue chan *storeModels.Store,
) {
	jd.logger.Info("start reading store data file")
	fp := ctx.Value(constants.FilePathContextKey).(string)
	resultStream, err := jd.localStorage.ReadFileArray(ctx, cancel, filepath.Join(jd.config.DataDir, fp))
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
			storeUpdateWaitGrp.Wait()
		}
	}
}

func (jd *StoreLoader) queueStoreProcessing(
	r map[string]interface{},
	storeQueue chan *storeModels.Store,
	cancel func(),
) {
	store, err := storeModels.MapResultToStore(r)
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
	storeQueue chan *storeModels.Store,
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

func (jd *StoreLoader) addStore(ctx context.Context, s *storeModels.Store) (*storeModels.Store, error) {
	if s.Org == "" {
		fileName := filepath.Base(ctx.Value(constants.FilePathContextKey).(string))
		s.Org = fileName[0:strings.Index(fileName, ".")]
	}
	// time.Sleep(50 * time.Millisecond)
	return jd.stores.AddStore(ctx, s)
}

func (jd *StoreLoader) publishStats(ctx context.Context) {
	jd.stores.SetReady(ctx, true)
	stats := jd.stores.GetStoreStats()
	jd.logger.Info("gateway status", zap.Any("stats", stats))
}

func (jd *StoreLoader) getFromStorage(ctx context.Context, filePath string) error {
	var fmod int64
	var f *os.File
	fStats, err := fileUtils.FileStats(filePath)
	if err != nil {
		jd.logger.Error("error accessing file", zap.Error(err), zap.String("filepath", filePath))
		err = fileUtils.CreateDirectory(filePath)
		if err != nil {
			jd.logger.Error("error creating data directory", zap.Error(err))
			return err
		}

		f, err = os.Create(filePath)
		if err != nil {
			jd.logger.Error("error creating file", zap.Error(err), zap.String("filepath", filePath))
			return errors.WrapError(err, fileUtils.ERROR_CREATING_FILE, filePath)
		}
		defer func() {
			if err := f.Close(); err != nil {
				jd.logger.Error("error closing file", zap.Error(err), zap.String("filepath", filePath))
			}
		}()
	} else {
		fmod := fStats.ModTime().Unix()
		jd.logger.Info("file mod time", zap.Int64("modtime", fmod), zap.String("filepath", filePath))

		f, err = os.Open(filePath)
		if err != nil {
			jd.logger.Error("error accessing file", zap.Error(err))
			return errors.WrapError(err, fileUtils.ERROR_NO_FILE, filePath)
		}
		defer func() {
			if err := f.Close(); err != nil {
				jd.logger.Error("error closing file", zap.Error(err), zap.String("filepath", filePath))
			}
		}()
	}

	cfr, err := cloudstorage.NewCloudFileRequest(jd.config.BucketName, filepath.Base(filePath), filepath.Dir(filePath), fmod)
	if err != nil {
		jd.logger.Error("error creating request", zap.Error(err), zap.String("filepath", filePath))
		return err
	}

	n, err := jd.cloudStorage.DownloadFile(ctx, f, cfr)
	if err != nil {
		jd.logger.Error("error downloading file", zap.Error(err), zap.String("filepath", filePath))
		return err
	}
	jd.logger.Info("uploaded file", zap.String("file", filepath.Base(filePath)), zap.String("path", filepath.Dir(filePath)), zap.Int64("bytes", n))
	return nil
}

func (jd *StoreLoader) StoreDataFile(ctx context.Context, filePath string) error {
	// dataPath := filepath.Join(jd.config.DataDir, filePath)
	file, err := jd.stores.Reader(ctx, jd.config.DataDir)
	if err != nil {
		jd.logger.Error("error saving data file", zap.Error(err), zap.String("filepath", filePath))
		return err
	}

	err = file.Close()
	if err != nil {
		jd.logger.Error("error closing data file", zap.Error(err), zap.String("filepath", filePath))
		return err
	}
	return nil
}

func (jd *StoreLoader) UploadDataFile(ctx context.Context, filePath string) error {
	dataPath := filepath.Join(jd.config.DataDir, filePath)
	fStats, err := fileUtils.FileStats(dataPath)
	if err != nil {
		jd.logger.Error("error accessing file", zap.Error(err), zap.String("filepath", filePath))
		return err
	}
	fmod := fStats.ModTime().Unix()
	jd.logger.Info("file mod time", zap.Int64("modtime", fmod), zap.String("filepath", filePath))

	file, err := os.Open(dataPath)
	if err != nil {
		jd.logger.Error("error accessing file", zap.Error(err))
		return errors.WrapError(err, fileUtils.ERROR_NO_FILE, filePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			jd.logger.Error("error closing file", zap.Error(err), zap.String("filepath", filePath))
		}
	}()

	cfr, err := cloudstorage.NewCloudFileRequest(jd.config.BucketName, filepath.Base(dataPath), filepath.Dir(dataPath), fmod)
	if err != nil {
		jd.logger.Error("error creating request", zap.Error(err), zap.String("filepath", filePath))
		return err
	}

	n, err := jd.cloudStorage.UploadFile(ctx, file, cfr)
	if err != nil {
		jd.logger.Error("error uploading file", zap.Error(err))
		return err
	}
	jd.logger.Info("uploaded file", zap.String("file", filepath.Base(dataPath)), zap.String("path", filepath.Dir(dataPath)), zap.Int64("bytes", n))
	return nil
}
