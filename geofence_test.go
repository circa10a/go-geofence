package geofence

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestValidateIPAddress(t *testing.T) {
	type test struct {
		Expected error
		Input    string
	}
	tests := []test{
		{
			Input:    "8.8.8.8",
			Expected: nil,
		},
		{
			Input:    "8.8.88",
			Expected: &ErrInvalidIPAddress{},
		},
		{
			Input:    "2001:db8:3333:4444:5555:6666:7777:8888",
			Expected: nil,
		},
		{
			Input:    "2001:db8:3333:4444:5555:6666:7777:88888",
			Expected: &ErrInvalidIPAddress{},
		},
	}
	for _, test := range tests {
		actual := validateIPAddress(test.Input)
		if test.Expected != nil {
			assert.EqualErrorf(t, actual, invalidIPAddressString, "Error should be: %v, got: %v", invalidIPAddressString, actual)
		} else {
			if !errors.Is(actual, test.Expected) {
				t.Fail()
			}
		}
	}
}

func TestGeofenceNear(t *testing.T) {
	fakeIPAddress := "8.8.8.8"
	fakeApiToken := "fakeApiToken"
	fakeLatitude := 37.751
	fakeLongitude := -97.822
	fakeRadius := 0.0
	fakeEndpoint := fmt.Sprintf("%s/%s?apikey=%s", freeGeoIPBaseURL, fakeIPAddress, fakeApiToken)

	// new geofence
	geofence, _ := New(&Config{
		IPAddress: fakeIPAddress,
		Token:     fakeApiToken,
		Radius:    fakeRadius,
		CacheTTL:  7 * (24 * time.Hour), // 1 week
	})
	geofence.Latitude = fakeLatitude
	geofence.Longitude = fakeLongitude

	httpmock.ActivateNonDefault(geofence.FreeGeoIPClient.GetClient())
	defer httpmock.DeactivateAndReset()

	// mock json rsponse
	response := &FreeGeoIPResponse{
		IP:          fakeIPAddress,
		CountryCode: "US",
		CountryName: "United States",
		TimeZone:    "America/Chicago",
		Latitude:    fakeLatitude,
		Longitude:   fakeLongitude,
	}

	// mock freegeoip.app response
	httpmock.RegisterResponder("GET", fakeEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, response)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		})

	// check that addresses is nearby
	isAddressNearby, err := geofence.IsIPAddressNear(fakeIPAddress)
	assert.NoError(t, err)
	assert.True(t, isAddressNearby)

	// get count info
	httpmock.GetTotalCallCount()
	// get the amount of calls for the registered responder
	info := httpmock.GetCallCountInfo()
	// Check total calls
	assert.Equal(t, info[fmt.Sprintf("GET %s", fakeEndpoint)], 1)
}

func TestGeofenceNotNear(t *testing.T) {
	fakeIPAddress := "8.8.8.8"
	fakeApiToken := "fakeApiToken"
	fakeLatitude := 37.751
	fakeLongitude := -98.822
	fakeRadius := 0.0
	fakeEndpoint := fmt.Sprintf("%s/%s?apikey=%s", freeGeoIPBaseURL, fakeIPAddress, fakeApiToken)

	// new geofence
	geofence, _ := New(&Config{
		IPAddress: fakeIPAddress,
		Token:     fakeApiToken,
		Radius:    fakeRadius,
		CacheTTL:  7 * (24 * time.Hour), // 1 week
	})
	geofence.Latitude = fakeLatitude + 1
	geofence.Longitude = fakeLongitude + 1

	httpmock.ActivateNonDefault(geofence.FreeGeoIPClient.GetClient())
	defer httpmock.DeactivateAndReset()

	// mock json rsponse
	response := &FreeGeoIPResponse{
		IP:          fakeIPAddress,
		CountryCode: "US",
		CountryName: "United States",
		TimeZone:    "America/Chicago",
		Latitude:    fakeLatitude,
		Longitude:   fakeLongitude,
	}

	// mock freegeoip.app response
	httpmock.RegisterResponder("GET", fakeEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, response)
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		})

	// check that addresses is not nearby
	isAddressNearby, err := geofence.IsIPAddressNear(fakeIPAddress)
	assert.NoError(t, err)
	assert.False(t, isAddressNearby)

	// get count info
	httpmock.GetTotalCallCount()
	// get the amount of calls for the registered responder
	info := httpmock.GetCallCountInfo()
	// Check total calls
	assert.Equal(t, info[fmt.Sprintf("GET %s", fakeEndpoint)], 1)
}
