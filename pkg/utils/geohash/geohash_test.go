package geohash

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/comfforts/geocode"
)

// At highest resolution, hash should differ for points where
// either latitude is apart by atleast 0.044
// or longitude is apart by atleast 0.089
const (
	LATITUDE_RESOLUTION  = 0.044
	LONGITUDE_RESOLUTION = 0.088
)

func TestEncodingResolution(t *testing.T) {
	points := []geocode.Point{
		{Latitude: 0.133333, Longitude: 117.500000},
		{Latitude: -33.918861, Longitude: 18.423300},
		{Latitude: 38.294788, Longitude: -122.461510},
		{Latitude: 28.644800, Longitude: 77.216721},
	}

	for _, point := range points {
		hash, _ := Encode(point.Latitude, point.Longitude, 12)
		bound, err := Decode(hash)
		require.NoError(t, err)

		require.Equal(t, true, math.Abs(bound.Latitude.Max-bound.Latitude.Min) < LATITUDE_RESOLUTION)
		require.Equal(t, true, math.Abs(bound.Longitude.Max-bound.Longitude.Min) < LONGITUDE_RESOLUTION)
	}
}
func TestEncodingResolutionChange(t *testing.T) {
	point := geocode.Point{Latitude: 42.713456, Longitude: -79.819675}

	res := map[string][]geocode.Point{}

	n := 0
	cn := 0
	for n < 10 {
		// get hash
		hash, _ := Encode(point.Latitude, point.Longitude, 12)
		_, ok := res[hash]
		if !ok {
			res[hash] = []geocode.Point{}
		}
		res[hash] = append(res[hash], point)
		cn++

		// move latitude and get hash
		hash, _ = Encode(point.Latitude+0.045, point.Longitude, 12)
		_, ok = res[hash]
		if !ok {
			res[hash] = []geocode.Point{}
		}
		res[hash] = append(res[hash], geocode.Point{
			Latitude:  point.Latitude + LATITUDE_RESOLUTION,
			Longitude: point.Longitude,
		})
		cn++

		// move longitude and get hash
		hash, _ = Encode(point.Latitude, point.Longitude+0.09, 12)
		_, ok = res[hash]
		if !ok {
			res[hash] = []geocode.Point{}
		}
		res[hash] = append(res[hash], geocode.Point{
			Latitude:  point.Latitude,
			Longitude: point.Longitude + LATITUDE_RESOLUTION,
		})
		cn++

		point.Latitude = point.Latitude + 0.045
		point.Longitude = point.Longitude + LONGITUDE_RESOLUTION
		n++
	}
	require.Equal(t, cn, len(res))
}
