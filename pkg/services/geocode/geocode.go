package geocode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/errors"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/cache"
	"github.com/comfforts/comff-stores/pkg/utils/geohash"
	"go.uber.org/zap"
)

const (
	ThirtyDays    = 24 * 30 * time.Hour
	OneDay        = 24 * time.Hour
	FiveHours     = 5 * time.Hour
	OneHour       = time.Hour
	ThirtyMinutes = 30 * time.Minute
)

type GeocoderResults struct {
	Results []Result `json:"results"`
	Status  string   `json:"status"`
}

type Result struct {
	AddressComponents []Address `json:"address_components"`
	FormattedAddress  string    `json:"formatted_address"`
	Geometry          Geometry  `json:"geometry"`
	PlaceId           string    `json:"place_id"`
	Types             []string  `json:"types"`
}

// Address store each address is identified by the 'types'
type Address struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

// Geometry store each value in the geometry
type Geometry struct {
	Bounds       Bounds `json:"bounds"`
	Location     LatLng `json:"location"`
	LocationType string `json:"location_type"`
	Viewport     Bounds `json:"viewport"`
}

// Bounds Northeast and Southwest
type Bounds struct {
	Northeast LatLng `json:"northeast"`
	Southwest LatLng `json:"southwest"`
}

// LatLng store the latitude and longitude
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type GeoCodeService struct {
	host   string
	path   string
	config config.GeoCodeServiceConfig
	logger *logging.AppLogger
	cache  *cache.CacheService
}

func unmarshalLPoint(p interface{}) (interface{}, error) {
	var point geohash.Point
	body, err := json.Marshal(p)
	if err != nil {
		appErr := errors.WrapError(err, cache.ERROR_MARSHALLING_CACHE_OBJECT)
		return point, appErr
	}

	err = json.Unmarshal(body, &point)
	if err != nil {
		appErr := errors.WrapError(err, cache.ERROR_UNMARSHALLING_CACHE_JSON)
		return point, appErr
	}
	return point, nil
}

func NewGeoCodeService(config config.GeoCodeServiceConfig, logger *logging.AppLogger) (*GeoCodeService, error) {
	if config.GeocoderKey == "" || logger == nil {
		return nil, errors.NewAppError(errors.ERROR_MISSING_REQUIRED)
	}

	geocoderSrv := GeoCodeService{
		host:   "https://maps.googleapis.com",
		path:   "/maps/api/geocode/json",
		config: config,
		logger: logger,
	}

	if config.IsCached {
		c, err := cache.NewCacheService("GEOCODE", logger, unmarshalLPoint)
		if err != nil {
			logger.Error("error setting up cache service", zap.Error(err))
			return nil, err
		}
		geocoderSrv.cache = c
	}
	return &geocoderSrv, nil
}

func (g *GeoCodeService) Geocode(ctx context.Context, postalCode, countryCode string) (*geohash.Point, error) {
	if ctx == nil {
		g.logger.Error("context is nil", zap.Error(errors.ErrNilContext))
		return nil, errors.ErrNilContext
	}

	if g.config.IsCached {
		point, exp, err := g.getFromCache(postalCode)
		if err == nil {
			g.logger.Info("returning cached value", zap.String("postalCode", postalCode), zap.Any("exp", exp))
			return point, nil
		} else {
			g.logger.Error("geocoder cache get error", zap.Error(err), zap.String("postalCode", postalCode))
		}
	}

	if countryCode == "" {
		countryCode = "US"
	}

	url := g.postalCodeURL(countryCode, postalCode)
	r, err := http.Get(url)
	if err != nil {
		g.logger.Error("geocoder request error", zap.Error(err), zap.String("postalCode", postalCode))
		return nil, err
	}
	defer r.Body.Close()

	var results GeocoderResults
	err = json.NewDecoder(r.Body).Decode(&results)
	if err != nil || &results == (*GeocoderResults)(nil) || len(results.Results) < 1 {
		g.logger.Error(errors.ERROR_GEOCODING_POSTALCODE, zap.Error(err), zap.String("postalCode", postalCode))
		return nil, errors.NewAppError(errors.ERROR_GEOCODING_POSTALCODE)
	}
	lat, long := results.Results[0].Geometry.Location.Lat, results.Results[0].Geometry.Location.Lng

	geoPoint := geohash.Point{
		Latitude:  lat,
		Longitude: long,
	}

	if g.config.IsCached {
		err = g.setInCache(postalCode, geoPoint)
		if err != nil {
			g.logger.Error("geocoder cache set error", zap.Error(err), zap.String("postalCode", postalCode))
		}
	}

	return &geoPoint, nil
}

func (g *GeoCodeService) Clear() {
	g.logger.Info("cleaning up geo code data structures")
	if g.config.IsCached {
		g.cache.SaveFile()
	}
}

func (g *GeoCodeService) postalCodeURL(countryCode, postalCode string) string {
	return fmt.Sprintf("%s%s?components=country:%s|postal_code:%s&sensor=false&key=%s", g.host, g.path, countryCode, postalCode, g.config.GeocoderKey)
}

func (g *GeoCodeService) getFromCache(postalCode string) (*geohash.Point, time.Time, error) {
	val, exp, err := g.cache.Get(postalCode)
	if err != nil {
		g.logger.Error(cache.ERROR_GET_CACHE, zap.Error(err), zap.String("postalCode", postalCode))
		return nil, exp, err
	}
	point, ok := val.(geohash.Point)
	if !ok {
		g.logger.Error("error getting cache point value", zap.Error(cache.ErrGetCache), zap.String("postalCode", postalCode))
		return nil, exp, cache.ErrGetCache
	}
	return &point, exp, nil
}

func (g *GeoCodeService) setInCache(postalCode string, point geohash.Point) error {
	err := g.cache.Set(postalCode, point, ThirtyDays)
	if err != nil {
		g.logger.Error(cache.ERROR_SET_CACHE, zap.Error(err), zap.String("postalCode", postalCode))
		return err
	}
	return nil
}
