package geofence

import (
	"fmt"
	"net"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"
)

const (
	freeGeoIPBaseURL                = "https://api.freegeoip.app/json"
	invalidSensitivityErrString     = "invalid sensitivity. value must be between 0 - 5"
	invalidIPAddressString          = "invalid IPv4 address provided"
	deleteExpiredCacheItemsInternal = 10 * time.Minute
)

// Config holds the user configuration to setup a new geofence
type Config struct {
	IPAddress   string
	Token       string
	Sensitivity int
	CacheTTL    time.Duration
}

// Geofence holds a freegeoip.app client, cache and user supplied config
type Geofence struct {
	Cache           *cache.Cache
	FreeGeoIPClient *resty.Client
	Config
	Latitude  float64
	Longitude float64
}

// FreeGeoIPResponse is the json response from freegeoip.app
type FreeGeoIPResponse struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionCode  string  `json:"region_code"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	ZipCode     string  `json:"zip_code"`
	TimeZone    string  `json:"time_zone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
}

// FreeGeoIPError is the json response when there is an error from freegeoip.app
type FreeGeoIPError struct {
	Message string `json:"message"`
}

func (e *FreeGeoIPError) Error() string {
	return e.Message
}

// ErrInvalidSensitivity is the error raised when sensitivity is less than 0 or more than 5
type ErrInvalidSensitivity struct {
	msg string
}

func (e *ErrInvalidSensitivity) Error() string {
	return e.msg
}

// ErrInvalidIPAddress is the error raised when an invalid IP address is provided
type ErrInvalidIPAddress struct {
	msg string
}

func (e *ErrInvalidIPAddress) Error() string {
	return e.msg
}

// validateSensitivity ensures valid value between 0 - 5
func validateSensitivity(sensitivity int) error {
	if sensitivity < 0 || sensitivity > 5 {
		return &ErrInvalidSensitivity{
			msg: invalidSensitivityErrString,
		}
	}
	return nil
}

// validateIPAddress ensures valid ip address
func validateIPAddress(ipAddress string) error {
	ipAddressErr := &ErrInvalidIPAddress{
		msg: invalidIPAddressString,
	}
	if net.ParseIP(ipAddress) == nil {
		return ipAddressErr
	}
	return nil
}

// getIPGeoData fetches geolocation data for specified IP address from https://freegeoip.app
func (g *Geofence) getIPGeoData(ipAddress string) (*FreeGeoIPResponse, error) {
	resp, err := g.FreeGeoIPClient.R().
		SetHeader("Accept", "application/json").
		SetQueryParam("apikey", g.Token).
		SetResult(&FreeGeoIPResponse{}).
		SetError(&FreeGeoIPError{}).
		Get(ipAddress)
	if err != nil {
		return &FreeGeoIPResponse{}, err
	}

	// If api gives back status code >399, report error to user
	if resp.IsError() {
		return &FreeGeoIPResponse{}, resp.Error().(*FreeGeoIPError)
	}

	return resp.Result().(*FreeGeoIPResponse), nil
}

// New creates a new geofence for the IP address specified.
// Use "" as the ip address to geofence the machine your application is running on
// Token comes from https://freegeoip.app/
// Sensitivity is for proximity:
// 0 - 111 km
// 1 - 11.1 km
// 2 - 1.11 km
// 3 111 meters
// 4 11.1 meters
// 5 1.11 meters
func New(c *Config) (*Geofence, error) {
	// Create new client for freegeoip.app
	freeGeoIPClient := resty.New().SetBaseURL(freeGeoIPBaseURL)

	// Ensure sensitivity is between 1 - 5
	err := validateSensitivity(c.Sensitivity)
	if err != nil {
		return nil, err
	}

	// New Geofence object
	geofence := &Geofence{
		Config:          *c,
		FreeGeoIPClient: freeGeoIPClient,
		Cache:           cache.New(c.CacheTTL, deleteExpiredCacheItemsInternal),
	}

	// Get current location of specified IP address
	// If empty string, use public IP of device running this
	// Or use location of the specified IP
	ipAddressLookupDetails, err := geofence.getIPGeoData(c.IPAddress)
	if err != nil {
		return geofence, err
	}

	// Set the location of our geofence to compare against looked up IP's
	geofence.Latitude = ipAddressLookupDetails.Latitude
	geofence.Longitude = ipAddressLookupDetails.Longitude

	return geofence, nil
}

// formatCoordinates converts decimal points to size of sensitivity and givens back a string for comparison
func formatCoordinates(sensitivity int, location float64) string {
	return fmt.Sprintf("%*.*f", 0, sensitivity, location)
}

// IsIPAddressNear returns true if the specified address is within proximity
func (g *Geofence) IsIPAddressNear(ipAddress string) (bool, error) {
	// Ensure IP is valid first
	err := validateIPAddress(ipAddress)
	if err != nil {
		return false, err
	}

	// Check if ipaddress has been looked up before and is in cache
	if isIPAddressNear, found := g.Cache.Get(ipAddress); found {
		return isIPAddressNear.(bool), nil
	}

	// If not in cache, lookup IP and compare
	ipAddressLookupDetails, err := g.getIPGeoData(ipAddress)
	if err != nil {
		return false, err
	}

	// Format our IP coordinates and the clients
	currentLat := formatCoordinates(g.Sensitivity, g.Latitude)
	currentLong := formatCoordinates(g.Sensitivity, g.Longitude)
	clientLat := formatCoordinates(g.Sensitivity, ipAddressLookupDetails.Latitude)
	clientLong := formatCoordinates(g.Sensitivity, ipAddressLookupDetails.Longitude)

	// Compare coordinates
	isNear := currentLat == clientLat && currentLong == clientLong

	// Insert ip address and it's status into the cache if user instantiated a cache
	g.Cache.Set(ipAddress, isNear, cache.DefaultExpiration)

	return isNear, nil
}
