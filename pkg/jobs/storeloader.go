package jobs

import (
	"context"
	"strings"
	"sync"

	"github.com/comfforts/comff-stores/pkg/constants"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/store"
	"github.com/comfforts/comff-stores/pkg/utils/loader"
	"go.uber.org/zap"
)

type JSONDataLoader struct {
	logger *logging.AppLogger
	stores *store.StoreService
}

func NewJSONDataLoader(ss *store.StoreService, l *logging.AppLogger) *JSONDataLoader {
	return &JSONDataLoader{
		logger: l,
		stores: ss,
	}
}

func (jd *JSONDataLoader) ProcessFile() {
	defer func() {
		jd.logger.Info("finished setting up store data")
	}()

	ctx := context.WithValue(context.Background(), constants.FileNameContextKey, "starbucks.json")
	ctx = context.WithValue(ctx, constants.ReadRateContextKey, 2)
	ctx, cancel := context.WithCancel(ctx)

	cout := make(chan *store.Store)
	var wgp sync.WaitGroup
	var wgs sync.WaitGroup

	wgp.Add(1)
	go jd.readFile(ctx, cancel, &wgp, &wgs, cout)
	wgp.Add(1)
	go jd.processStore(ctx, cancel, &wgp, &wgs, cout)

	func() {
		wgp.Wait()
		jd.stores.SetReady(ctx, true)
		stats := jd.stores.GetStoreStats()
		jd.logger.Info("gateway status", zap.Any("stats", stats))
	}()
}

func (jd *JSONDataLoader) readFile(
	ctx context.Context,
	cancel func(),
	wgp *sync.WaitGroup,
	wgs *sync.WaitGroup,
	out chan *store.Store,
) {
	jd.logger.Info("start reading store data file")
	fileName := ctx.Value(constants.FileNameContextKey).(string)
	resultStream, err := loader.ReadFileArray(ctx, cancel, fileName)
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
				wgp.Done()
				close(out)
				return
			}
			wgs.Add(1)
			jd.publishStore(r, out, cancel)
		}
	}
}

func (jd *JSONDataLoader) publishStore(
	r map[string]interface{},
	out chan *store.Store,
	cancel func(),
) {
	store, err := store.MapResultToStore(r)
	if err != nil {
		jd.logger.Error("error processing store data", zap.Error(err), zap.Any("storeJson", r))
		// cancel()
	} else {
		out <- store
	}
}

func (jd *JSONDataLoader) processStore(
	ctx context.Context,
	cancel func(),
	wgp *sync.WaitGroup,
	wgs *sync.WaitGroup,
	out chan *store.Store,
) {
	jd.logger.Info("start updating store data")

	for {
		select {
		case <-ctx.Done():
			jd.logger.Info("store data update context done")
			return
		case store, ok := <-out:
			if !ok {
				jd.logger.Info("store notification channel closed")
				wgp.Done()
				return
			}
			addedStr, err := jd.updateDataStores(ctx, store)
			if addedStr == nil || err != nil {
				jd.logger.Error("error processing store data", zap.Error(err), zap.Any("store", store))
				// cancel()
			}
			wgs.Done()
		}
	}
}

func (jd *JSONDataLoader) updateDataStores(ctx context.Context, s *store.Store) (*store.Store, error) {
	// time.Sleep(10 * time.Millisecond)
	if s.Org == "" {
		fileName := ctx.Value(constants.FileNameContextKey).(string)
		s.Org = fileName[0:strings.Index(fileName, ".")]
	}
	return jd.stores.AddStore(ctx, s)
}
