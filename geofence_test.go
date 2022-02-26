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
	tests := []struct {
		expected error
		input    string
	}{
		{
			input:    "8.8.8.8",
			expected: nil,
		},
		{
			input:    "8.8.88",
			expected: ErrInvalidIPAddress,
		},
		{
			input:    "2001:db8:3333:4444:5555:6666:7777:8888",
			expected: nil,
		},
		{
			input:    "2001:db8:3333:4444:5555:6666:7777:88888",
			expected: ErrInvalidIPAddress,
		},
	}
	for _, test := range tests {
		actual := validateIPAddress(test.input)
		if test.expected != nil {
			assert.ErrorIs(t, actual, test.expected)
		} else {
			if !errors.Is(actual, test.expected) {
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
