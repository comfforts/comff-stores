package filestorage

import (
	"context"
	"fmt"
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
)

const (
	ERROR_CREATING_STORAGE_CLIENT string = "error creating storage client"
	ERROR_LISTING_OBJECTS         string = "error listing storage bucket objects"
	ERROR_DELETING_OBJECTS        string = "error deleting storage bucket objects"
	ERROR_UPLOADING_FILE          string = "uploading file %s"
	ERROR_CLOSING_FILE            string = "closing file %s"
	ERROR_NO_FILE                 string = "%s doesn't exist"
	ERROR_CREATING_FILE           string = "creating file %s"
	ERROR_DOWNLOADING_FILE        string = "downloading file %s"
)

type CloudStorageClient struct {
	client *storage.Client
	config config.CloudStorageClientConfig
	logger *logging.AppLogger
}

func NewCloudStorageClient(logger *logging.AppLogger, cfg config.CloudStorageClientConfig) (*CloudStorageClient, error) {
	if logger == nil {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.CredsPath)
	client, err := storage.NewClient(context.Background())
	if err != nil {
		logger.Error(ERROR_CREATING_STORAGE_CLIENT, zap.Error(err))
		return nil, errors.WrapError(err, ERROR_CREATING_STORAGE_CLIENT)
	}

	loaderClient := &CloudStorageClient{
		client: client,
		config: cfg,
		logger: logger,
	}

	return loaderClient, nil
}

func (lc *CloudStorageClient) UploadFile(ct context.Context, file io.Reader, fileName, uploadPath string) (int64, error) {
	ctx, cancel := context.WithTimeout(ct, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := lc.client.Bucket(lc.config.BucketName).Object(filepath.Join(uploadPath, fileName)).NewWriter(ctx)
	nBytes, err := io.Copy(wc, file)
	if err != nil {
		return 0, errors.WrapError(err, ERROR_UPLOADING_FILE, filepath.Join(uploadPath, fileName))
	}
	err = wc.Close()
	if err != nil {
		return 0, errors.WrapError(err, ERROR_CLOSING_FILE, filepath.Join(uploadPath, fileName))
	}

	return nBytes, nil
}

func (lc *CloudStorageClient) DownloadFile(ct context.Context, file io.Writer, fileName, uploadPath string) (int64, error) {
	ctx, cancel := context.WithTimeout(ct, time.Second*50)
	defer cancel()

	// download an object with storage.Reader.
	rc, err := lc.client.Bucket(lc.config.BucketName).Object(filepath.Join(uploadPath, fileName)).NewReader(ctx)
	if err != nil {
		lc.logger.Error(ERROR_DOWNLOADING_FILE, zap.Error(err), zap.String("filename", fileName))
		return 0, errors.WrapError(err, ERROR_DOWNLOADING_FILE, filepath.Join(uploadPath, fileName))
	}
	defer func() {
		err := rc.Close()
		if err != nil {
			lc.logger.Error(ERROR_CLOSING_FILE, zap.Error(err), zap.String("path", filepath.Join(uploadPath, fileName)))
		}
	}()

	nBytes, err := io.Copy(file, rc)
	if err != nil {
		lc.logger.Error(ERROR_DOWNLOADING_FILE, zap.Error(err), zap.String("filename", fileName))
		return 0, errors.WrapError(err, ERROR_DOWNLOADING_FILE, filepath.Join(uploadPath, fileName))
	}

	return nBytes, nil
}

func (lc *CloudStorageClient) ListObjects(ctx context.Context) ([]string, error) {
	bucket := lc.client.Bucket(lc.config.BucketName)
	it := bucket.Objects(ctx, nil)
	names := []string{}
	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			} else {
				lc.logger.Error(ERROR_LISTING_OBJECTS, zap.Error(err))
				return names, errors.WrapError(err, ERROR_LISTING_OBJECTS)
			}
		}
		names = append(names, objAttrs.Name)
	}
	return names, nil
}

func (lc *CloudStorageClient) DeleteObjects(ctx context.Context) error {
	bucket := lc.client.Bucket(lc.config.BucketName)
	it := bucket.Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			} else {
				lc.logger.Error(ERROR_LISTING_OBJECTS, zap.Error(err))
				return errors.WrapError(err, ERROR_LISTING_OBJECTS)
			}
		}
		fmt.Printf(" object attributes: %v\n", objAttrs)
		if err := bucket.Object(objAttrs.Name).Delete(ctx); err != nil {
			lc.logger.Error(ERROR_DELETING_OBJECTS, zap.Error(err))
			return errors.WrapError(err, ERROR_DELETING_OBJECTS)
		}
	}
	return nil
}
