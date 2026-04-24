package stores

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	geocl "github.com/comfforts/comff-geo-client"
	geo_v1 "github.com/comfforts/comff-geo/api/geo/v1"
	"github.com/comfforts/logger"

	indom "github.com/comfforts/comff-stores/internal/domain/infra"
	stdom "github.com/comfforts/comff-stores/internal/domain/stores"
	"github.com/comfforts/comff-stores/internal/infra/observability"
)

const DEFAULT_SEARCH_RADIUS_METERS = 5000

const (
	MISSING_REQUIRED_FIELD = "missing required field"
	INVALID_ADDRESS_ID     = "invalid address ID"
	INVALID_LAT_LON        = "invalid latitude/longitude"
	INVALID_ADDRESS_STR    = "invalid address string"
)

var (
	ErrMissingRequiredField = errors.New(MISSING_REQUIRED_FIELD)
	ErrInvalidAddressId     = errors.New(INVALID_ADDRESS_ID)
	ErrInvalidLatLon        = errors.New(INVALID_LAT_LON)
	ErrInvalidAddressStr    = errors.New(INVALID_ADDRESS_STR)
)

type StoresServiceConfig struct {
	MongoConfig indom.StoreConfig
}

func NewStoresServiceConfig(
	mongoCfg indom.StoreConfig,
) (*StoresServiceConfig, error) {
	return &StoresServiceConfig{
		MongoConfig: mongoCfg,
	}, nil
}

type storesService struct {
	metrics    observability.Metrics
	storesRepo stdom.StoresRepo
	geoClient  geocl.Client
}

func NewStoresService(ctx context.Context, sr stdom.StoresRepo, gc geocl.Client, mt observability.Metrics) (*storesService, error) {
	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}

	l.Info("initialized stores service")
	return &storesService{
		metrics:    mt,
		storesRepo: sr, // Initialize with actual storesRepo when available
		geoClient:  gc,
	}, nil
}

func (ss *storesService) AddStore(ctx context.Context, st *stdom.AddStoreParams) (string, error) {
	ctx, span := startSpan(ctx, "stores.service.add")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("adding store")

	if st == nil || st.AddressId == "" || st.Name == "" || st.Org == "" {
		finishSpan(span, ErrMissingRequiredField)
		return "", ErrMissingRequiredField
	}

	if _, err := ss.geoClient.GetGeoLocation(ctx, &geo_v1.GeoLocationRequest{
		AddressId: st.AddressId,
	}); err != nil {
		l.Error("error validating address ID with geo service", "address_id", st.AddressId, "error", err.Error())
		finishSpan(span, ErrInvalidAddressId)
		return "", ErrInvalidAddressId
	}

	id, err := ss.storesRepo.AddStore(ctx, &stdom.Store{
		Name:      st.Name,
		Org:       st.Org,
		AddressId: st.AddressId,
	})
	if err != nil {
		l.Error("error adding store to repository", "error", err.Error())
		finishSpan(span, err)
		return "", err
	}

	return id, nil
}

func (ss *storesService) GetStore(ctx context.Context, id string) (*stdom.Store, error) {
	ctx, span := startSpan(ctx, "stores.service.get")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("getting store")

	if id == "" {
		finishSpan(span, ErrMissingRequiredField)
		return nil, ErrMissingRequiredField
	}

	store, err := ss.storesRepo.GetStore(ctx, id)
	if err != nil {
		finishSpan(span, err)
		return nil, err
	}
	return store, nil
}

func (ss *storesService) UpdateStore(ctx context.Context, id string, params *stdom.UpdateStoreParams) error {
	ctx, span := startSpan(ctx, "stores.service.update")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("updating store")

	if id == "" {
		finishSpan(span, ErrMissingRequiredField)
		return ErrMissingRequiredField
	}

	if params == nil || (params.Name == "" && params.Org == "" && params.AddressId == "") {
		finishSpan(span, ErrMissingRequiredField)
		return ErrMissingRequiredField
	}

	if err := ss.storesRepo.UpdateStore(ctx, id, &stdom.UpdateStoreQuery{
		Name:      params.Name,
		Org:       params.Org,
		AddressId: params.AddressId,
	}); err != nil {
		l.Error("error updating store in repository", "error", err.Error())
		finishSpan(span, err)
		return err
	}
	return nil
}

func (ss *storesService) DeleteStore(ctx context.Context, id string) error {
	ctx, span := startSpan(ctx, "stores.service.delete")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("deleting store")

	if id == "" {
		finishSpan(span, ErrMissingRequiredField)
		return ErrMissingRequiredField
	}

	err = ss.storesRepo.DeleteStore(ctx, id)
	if err != nil {
		finishSpan(span, err)
		return err
	}
	return nil
}

func (ss *storesService) SearchStores(ctx context.Context, params *stdom.SearchStoreParams) ([]*stdom.Store, error) {
	ctx, span := startSpan(ctx, "stores.service.search")
	defer span.End()

	l, err := logger.LoggerFromContext(ctx)
	if err != nil {
		l = logger.GetSlogLogger()
	}
	l.Debug("searching stores")

	if params == nil || (params.Org == "" && params.Name == "" && params.AddressId == "" && params.AddressStr == "" && (params.Latitude == 0 || params.Longitude == 0)) {
		finishSpan(span, ErrMissingRequiredField)
		return nil, ErrMissingRequiredField
	}

	if params.AddressId == "" {
		if params.AddressStr != "" {
			geoResp, err := ss.geoClient.GeoLocate(ctx, &geo_v1.GeoRequest{
				AddressStr: params.AddressStr,
			})
			if err != nil {
				l.Error("error validating address string with geo service", "address_str", params.AddressStr, "error", err.Error())
				finishSpan(span, ErrInvalidAddressStr)
				return nil, ErrInvalidAddressStr
			}
			params.AddressId = geoResp.GetPoint().GetHash()
		} else if params.Latitude != 0 && params.Longitude != 0 {
			geoResp, err := ss.geoClient.GeoLocate(ctx, &geo_v1.GeoRequest{
				Latitude:  params.Latitude,
				Longitude: params.Longitude,
			})
			if err != nil {
				l.Error("error validating latitude/longitude with geo service", "latitude", params.Latitude, "longitude", params.Longitude, "error", err.Error())
				finishSpan(span, ErrInvalidLatLon)
				return nil, ErrInvalidLatLon
			}
			params.AddressId = geoResp.GetPoint().GetHash()
		}

		if params.AddressId != "" {
			if params.Distance == 0 {
				params.Distance = DEFAULT_SEARCH_RADIUS_METERS
			}

			// shave off hash prefix for requested distance-based search
			hashLen := len(params.AddressId)
			dist := params.Distance / 1000
			switch {
			case dist <= 1:
				if hashLen > 18 {
					params.AddressId = params.AddressId[0:18]
				}
			case dist <= 10:
				if hashLen > 15 {
					params.AddressId = params.AddressId[0:15]
				}
			case dist <= 100:
				if hashLen > 12 {
					params.AddressId = params.AddressId[0:12]
				}
			default:
				if hashLen > 15 {
					params.AddressId = params.AddressId[0:15]
				}
			}
			l.Debug("searching stores around point", "point", params.AddressId)
		}
	}

	searchQry := &stdom.SearchStoreQuery{
		Org:       params.Org,
		Name:      params.Name,
		AddressId: params.AddressId,
	}

	stores, err := ss.storesRepo.SearchStores(ctx, searchQry)
	if err != nil {
		l.Error("error searching stores in repository", "error", err.Error())
		finishSpan(span, err)
		return nil, err
	}
	return stores, nil
}

func startSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return otel.Tracer("stores-service").Start(ctx, name, trace.WithAttributes(attrs...))
}

func finishSpan(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(otelcodes.Error, err.Error())
}
