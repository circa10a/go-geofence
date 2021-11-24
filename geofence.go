package geofence

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"
)

const (
	freeGeoIPBaseURL = "https://api.freegeoip.app/json"
)

// Geofence holds a Geofenced IP config
type Geofence struct {
	Cache           *cache.Cache
	FreeGeoIPClient *resty.Client
	token           string
	Sensitivity     int
	Latitude        float64
	Longitude       float64
	CacheCreated    bool
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

// formatCoordinates converts decimal points to size of sensitivity and givens back a string for comparison
func formatCoordinates(sensitivity int, location float64) string {
	return fmt.Sprintf("%*.*f", 0, sensitivity, location)
}

// ErrInvalidSensitivity is the error raised when sensitivity is less than 0 or more than 5
var ErrInvalidSensitivity = errors.New("invalid sensitivity. value must be between 0 - 5")

// validateSensitivity ensures valid value between 0 - 5
func validateSensitivity(sensitivity int) error {
	if sensitivity < 0 || sensitivity > 5 {
		return ErrInvalidSensitivity
	}
	return nil
}

// ErrInvalidIPAddress is the error raised when an invalid IP address is provided
var ErrInvalidIPAddress = errors.New("invalid IPv4 address provided")

// validateIPAddress ensures valid ipv4 address
func validateIPAddress(ipAddress string) error {
	if net.ParseIP(ipAddress) == nil {
		return ErrInvalidIPAddress
	}
	return nil
}

// getIPGeoData fetches geolocation data for specified IP address from https://freegeoip.app
func (g *Geofence) getIPGeoData(ipAddress string) (*FreeGeoIPResponse, error) {
	resp, err := g.FreeGeoIPClient.R().
		SetHeader("Accept", "application/json").
		SetQueryParam("apikey", g.token).
		SetResult(&FreeGeoIPResponse{}).
		SetError(&FreeGeoIPError{}).
		Get(ipAddress)
	if err != nil {
		return &FreeGeoIPResponse{}, err
	}
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
func New(ipAddress, freeGeoIPAPIToken string, sensitivity int) (*Geofence, error) {
	// Create new client for freegeoip.app
	freeGeoIPClient := resty.New().SetBaseURL(freeGeoIPBaseURL)

	// Ensure sensitivity is between 1 - 5
	err := validateSensitivity(sensitivity)
	if err != nil {
		return nil, err
	}

	// New Geofence object
	geofence := &Geofence{
		FreeGeoIPClient: freeGeoIPClient,
		Sensitivity:     sensitivity,
	}

	// Hold token
	geofence.token = freeGeoIPAPIToken

	// Get current location of specified IP address
	// If empty string, use public IP of device running this
	// Or use location of the specified IP
	ipAddressLookupDetails, err := geofence.getIPGeoData(ipAddress)
	if err != nil {
		return nil, err
	}

	// Set the location of our geofence to compare against looked up IP's
	geofence.Latitude = ipAddressLookupDetails.Latitude
	geofence.Longitude = ipAddressLookupDetails.Longitude

	return geofence, nil
}

// CreateCache creates a new cache for IP address lookups to reduce calls/improve performance
// Accepts a duration to keep items in cache. Use -1 to keep items in memory indefinitely
func (g *Geofence) CreateCache(duration time.Duration) {
	if !g.CacheCreated {
		g.Cache = cache.New(duration, duration)
		g.CacheCreated = true
	}
}

// IsIPAddressNear returns true if the specified address is within proximity
func (g *Geofence) IsIPAddressNear(ipAddress string) (bool, error) {
	// Ensure IP is valid first
	err := validateIPAddress(ipAddress)
	if err != nil {
		return false, err
	}
	// Check if ipaddress has been looked up before and is in cache
	if g.CacheCreated {
		if isIPAddressNear, found := g.Cache.Get(ipAddress); found {
			return isIPAddressNear.(bool), nil
		}
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
	if g.CacheCreated {
		g.Cache.Set(ipAddress, isNear, cache.DefaultExpiration)
	}
	return isNear, nil
}
