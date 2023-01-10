package filestorage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	fileUtils "github.com/comfforts/comff-stores/pkg/utils/file"
)

type CloudStorageClient struct {
	client *storage.Client
	config config.CloudStorageClientConfig
	logger *logging.AppLogger
}

type CloudFileRequest struct {
	bucket  string
	file    string
	path    string
	modTime int64
}

func NewCloudFileRequest(bucketName, fileName, path string, modTime int64) (CloudFileRequest, error) {
	if bucketName == "" {
		return CloudFileRequest{}, fileUtils.ErrBucketNameMissing
	}
	return CloudFileRequest{
		bucket:  bucketName,
		file:    fileName,
		path:    path,
		modTime: modTime,
	}, nil
}

func NewCloudStorageClient(cfg config.CloudStorageClientConfig, logger *logging.AppLogger) (*CloudStorageClient, error) {
	if logger == nil {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.CredsPath)
	client, err := storage.NewClient(context.Background())
	if err != nil {
		logger.Error(fileUtils.ERROR_CREATING_STORAGE_CLIENT, zap.Error(err))
		return nil, errors.WrapError(err, fileUtils.ERROR_CREATING_STORAGE_CLIENT)
	}

	loaderClient := &CloudStorageClient{
		client: client,
		config: cfg,
		logger: logger,
	}

	return loaderClient, nil
}

func (lc *CloudStorageClient) UploadFile(ct context.Context, file io.Reader, cfr CloudFileRequest) (int64, error) {
	if cfr.file == "" {
		return 0, fileUtils.ErrFileNameMissing
	}
	fPath := cfr.file
	if cfr.path != "" {
		fPath = filepath.Join(cfr.path, cfr.file)
	}

	ctx, cancel := context.WithTimeout(ct, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	obj := lc.client.Bucket(cfr.bucket).Object(fPath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		lc.logger.Error(fileUtils.ERROR_FILE_INACCESSIBLE, zap.Error(err), zap.String("filepath", fPath))
	} else {
		objcr := attrs.Created.Unix()
		objmod := attrs.Updated.Unix()
		lc.logger.Info("object created time", zap.Any("created", objcr), zap.Any("updated", objmod), zap.String("filepath", fPath))
	}

	wc := obj.NewWriter(ctx)
	defer func() {
		if err := wc.Close(); err != nil {
			lc.logger.Error(fileUtils.ERROR_CLOSING_FILE, zap.Error(err), zap.String("filepath", fPath))
		}
	}()

	nBytes, err := io.Copy(wc, file)
	if err != nil {
		lc.logger.Error(fileUtils.ERROR_UPLOADING_FILE, zap.Error(err), zap.String("filepath", fPath))
		return 0, errors.WrapError(err, fileUtils.ERROR_UPLOADING_FILE, fPath)
	}

	return nBytes, nil
}

func (lc *CloudStorageClient) DownloadFile(ct context.Context, file io.Writer, cfr CloudFileRequest) (int64, error) {
	if cfr.file == "" {
		return 0, fileUtils.ErrFileNameMissing
	}
	fPath := cfr.file
	if cfr.path != "" {
		fPath = filepath.Join(cfr.path, cfr.file)
	}

	ctx, cancel := context.WithTimeout(ct, time.Second*50)
	defer cancel()

	// download an object with storage.Reader.
	obj := lc.client.Bucket(cfr.bucket).Object(fPath)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		lc.logger.Error(fileUtils.ERROR_FILE_INACCESSIBLE, zap.Error(err), zap.String("filepath", fPath))
		return 0, errors.WrapError(err, fileUtils.ERROR_FILE_INACCESSIBLE, fPath)
	}
	objcr := attrs.Created.Unix()
	objmod := attrs.Updated.Unix()
	lc.logger.Info("object created time", zap.Any("created", objcr), zap.Any("updated", objmod), zap.String("filepath", fPath))

	rc, err := obj.NewReader(ctx)
	if err != nil {
		lc.logger.Error(fileUtils.ERROR_DOWNLOADING_FILE, zap.Error(err), zap.String("filepath", fPath))
		return 0, errors.WrapError(err, fileUtils.ERROR_DOWNLOADING_FILE, fPath)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			lc.logger.Error(fileUtils.ERROR_CLOSING_FILE, zap.Error(err), zap.String("filepath", fPath))
		}
	}()

	nBytes, err := io.Copy(file, rc)
	if err != nil {
		lc.logger.Error(fileUtils.ERROR_DOWNLOADING_FILE, zap.Error(err), zap.String("filepath", fPath))
		return 0, errors.WrapError(err, fileUtils.ERROR_DOWNLOADING_FILE, fPath)
	}

	return nBytes, nil
}

func (lc *CloudStorageClient) ListObjects(ctx context.Context, cfr CloudFileRequest) ([]string, error) {
	bucket := lc.client.Bucket(cfr.bucket)
	it := bucket.Objects(ctx, nil)
	names := []string{}
	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			} else {
				lc.logger.Error(fileUtils.ERROR_LISTING_OBJECTS, zap.Error(err))
				return names, errors.WrapError(err, fileUtils.ERROR_LISTING_OBJECTS)
			}
		}
		names = append(names, objAttrs.Name)
	}
	return names, nil
}

func (lc *CloudStorageClient) DeleteObjects(ctx context.Context, cfr CloudFileRequest) error {
	bucket := lc.client.Bucket(cfr.bucket)
	it := bucket.Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			} else {
				lc.logger.Error(fileUtils.ERROR_LISTING_OBJECTS, zap.Error(err))
				return errors.WrapError(err, fileUtils.ERROR_LISTING_OBJECTS)
			}
		}
		lc.logger.Info("object attributes", zap.Any("objAttrs", objAttrs))
		if err := bucket.Object(objAttrs.Name).Delete(ctx); err != nil {
			lc.logger.Error(fileUtils.ERROR_DELETING_OBJECTS, zap.Error(err))
			return errors.WrapError(err, fileUtils.ERROR_DELETING_OBJECTS)
		}
	}
	return nil
}
