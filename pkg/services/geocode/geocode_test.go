package geocode

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/comfforts/comff-stores/pkg/config"
	"github.com/comfforts/comff-stores/pkg/logging"
	"github.com/comfforts/comff-stores/pkg/services/filestorage"
)

func TestGeocoder(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client *GeoCodeService,
	){
		"request gecoding succeeds": testGeocode,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, teardown := setupTest(t)
			defer teardown()
			fn(t, client)
		})
	}
}

func setupTest(t *testing.T) (
	client *GeoCodeService,
	teardown func(),
) {
	t.Helper()

	logger := zaptest.NewLogger(t)
	appLogger := logging.NewAppLogger(logger, nil)

	appCfg, err := config.GetAppConfig("test-config.json", appLogger)
	require.NoError(t, err)

	cscCfg := appCfg.Services.CloudStorageClientCfg
	csc, err := filestorage.NewCloudStorageClient(cscCfg, appLogger)
	require.NoError(t, err)

	gscCfg := appCfg.Services.GeoCodeCfg
	gsc, err := NewGeoCodeService(gscCfg, csc, appLogger)
	require.NoError(t, err)

	return gsc, func() {
		t.Log(" TestGeocoder ended")

		err := os.RemoveAll(gscCfg.DataPath)
		require.NoError(t, err)
	}
}

func testGeocode(t *testing.T, client *GeoCodeService) {
	postalCode := "92612"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pt, err := client.Geocode(ctx, postalCode, "")
	require.NoError(t, err)
	require.Equal(t, "33.66", fmt.Sprintf("%0.2f", pt.Latitude))
	require.Equal(t, "-117.83", fmt.Sprintf("%0.2f", pt.Longitude))
}
