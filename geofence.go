package geofence

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/qioalice/ipstack"
)

// Geofence holds a Geofenced IP config
type Geofence struct {
	Cache         *cache.Cache
	CacheCreated  bool
	IPStackClient *ipstack.Client
	Latitude      float32
	Longitude     float32
	Sensitivity   int
}

// formatCoordinates converts decimal points to size of sensitivity and givens back a string for comparison
func formatCoordinates(sensitivity int, location float32) string {
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

// New creates a new geofence for the IP address specified.
// Use "" as the ip address to geofence the machine your application is running on
// Token comes from ipstack.com
// Sensitivity is for proximity:
// 0 - 111 km
// 1 - 11.1 km
// 2 - 1.11 km
// 3 111 meters
// 4 11.1 meters
// 5 1.11 meters
func New(ipAddress, ipStackAPIToken string, sensitivity int) (*Geofence, error) {
	// Create new client for ipstack.com
	ipStackClient, err := ipstack.New(ipStackAPIToken)
	if err != nil {
		return nil, err
	}

	// Ensure sensitivity is between 1 - 5
	err = validateSensitivity(sensitivity)
	if err != nil {
		return nil, err
	}

	// New Geofence object
	geofence := &Geofence{
		IPStackClient: ipStackClient,
		Sensitivity:   sensitivity,
	}

	// If no ip address passed, get current device location details
	if ipAddress == "" {
		currentHostLocation, err := geofence.IPStackClient.Me()
		if err != nil {
			return nil, err
		}
		geofence.Latitude = currentHostLocation.Latitide
		geofence.Longitude = currentHostLocation.Longitude
		// If address is passed, fetch details for it
	} else {
		err = validateIPAddress(ipAddress)
		if err != nil {
			return nil, err
		}
		remoteHostLocation, err := geofence.IPStackClient.IP(ipAddress)
		if err != nil {
			return nil, err
		}
		geofence.Latitude = remoteHostLocation.Latitide
		geofence.Longitude = remoteHostLocation.Longitude
	}
	return geofence, nil
}

// CreateCache creates a new cache for IP address lookups to reduce ipstack.com calls/improve performance
// Accepts a duration to keep items in cache. Use -1 to keep items in memory indefinitely
func (g *Geofence) CreateCache(duration time.Duration) {
	g.Cache = cache.New(duration, duration)
	g.CacheCreated = true
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
	ipAddressLookupDetails, err := g.IPStackClient.IP(ipAddress)
	if err != nil {
		return false, err
	}
	// Format our IP coordinates and the clients
	currentLat := formatCoordinates(g.Sensitivity, g.Latitude)
	currentLong := formatCoordinates(g.Sensitivity, g.Longitude)
	clientLat := formatCoordinates(g.Sensitivity, ipAddressLookupDetails.Latitide)
	clientLong := formatCoordinates(g.Sensitivity, ipAddressLookupDetails.Longitude)
	// Compare coordinates
	isNear := currentLat == clientLat && currentLong == clientLong
	// Insert ip address and it's status into the cache if user instantiated a cache
	if g.CacheCreated {
		g.Cache.Set(ipAddress, isNear, cache.DefaultExpiration)
	}
	return isNear, nil
}
