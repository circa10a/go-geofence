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

func TestFormatCoordinates(t *testing.T) {
	type test struct {
		Expected    string
		Sensitivity int
		Input       float64
	}
	tests := []test{
		{
			Input:       -31.12345,
			Expected:    "-31.12345",
			Sensitivity: 5,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.1234",
			Sensitivity: 4,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.123",
			Sensitivity: 3,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.123",
			Sensitivity: 3,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.12",
			Sensitivity: 2,
		},
		{
			Input:       -31.12345,
			Expected:    "-31.1",
			Sensitivity: 1,
		},
		{
			Input:       -31.12345,
			Expected:    "-31",
			Sensitivity: 0,
		},
	}
	for _, test := range tests {
		actual := formatCoordinates(test.Sensitivity, test.Input)
		if test.Expected != actual {
			t.Fail()
		}
	}
}

func TestValidateSensitivity(t *testing.T) {
	type test struct {
		Expected error
		Input    int
	}
	tests := []test{
		{
			Input:    6,
			Expected: &ErrInvalidSensitivity{},
		},
		{
			Input:    5,
			Expected: nil,
		},
		{
			Input:    4,
			Expected: nil,
		},
		{
			Input:    3,
			Expected: nil,
		},
		{
			Input:    2,
			Expected: nil,
		},
		{
			Input:    1,
			Expected: nil,
		},
		{
			Input:    0,
			Expected: nil,
		},
		{
			Input:    -1,
			Expected: &ErrInvalidSensitivity{},
		},
	}
	for _, test := range tests {
		actual := validateSensitivity(test.Input)
		if test.Expected != nil {
			assert.EqualErrorf(t, actual, invalidSensitivityErrString, "Error should be: %v, got: %v", invalidSensitivityErrString, actual)
		} else {
			if !errors.Is(actual, test.Expected) {
				t.Fail()
			}
		}
	}
}

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
	fakeEndpoint := fmt.Sprintf("%s/%s?apikey=%s", freeGeoIPBaseURL, fakeIPAddress, fakeApiToken)

	// new geofence
	geofence, _ := New(&Config{
		IPAddress:   fakeIPAddress,
		Token:       fakeApiToken,
		Sensitivity: 3,                    // 3 is recommended
		CacheTTL:    7 * (24 * time.Hour), // 1 week
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

	isAddressNearby, err := geofence.IsIPAddressNear(fakeIPAddress)
	assert.NoError(t, err)
	assert.True(t, isAddressNearby)
}

func TestGeofenceNotNear(t *testing.T) {
	fakeIPAddress := "8.8.8.8"
	fakeApiToken := "fakeApiToken"
	fakeLatitude := 37.751
	fakeLongitude := -98.822
	fakeEndpoint := fmt.Sprintf("%s/%s?apikey=%s", freeGeoIPBaseURL, fakeIPAddress, fakeApiToken)

	// new geofence
	geofence, _ := New(&Config{
		IPAddress:   fakeIPAddress,
		Token:       fakeApiToken,
		Sensitivity: 3,                    // 3 is recommended
		CacheTTL:    7 * (24 * time.Hour), // 1 week
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

	isAddressNearby, err := geofence.IsIPAddressNear(fakeIPAddress)
	assert.NoError(t, err)
	assert.False(t, isAddressNearby)
}
