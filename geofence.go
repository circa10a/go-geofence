package geofence

import (
	"errors"
	"net"
	"time"

	"github.com/EpicStep/go-simple-geo/v2/geo"
	"github.com/circa10a/go-geofence/cache"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/context"
)

const (
	ipBaseBaseURL = "https://api.ipbase.com/v2"
)

// Config holds the user configuration to setup a new geofence
type Config struct {
	RedisOptions            *cache.RedisOptions
	IPAddress               string
	Token                   string
	Radius                  float64
	CacheTTL                time.Duration
	AllowPrivateIPAddresses bool
}

// Geofence holds a ipbase.com client, redis client, in-memory cache and user supplied config
type Geofence struct {
	cache        cache.Cache
	ipbaseClient *resty.Client
	ctx          context.Context
	Config       Config
	Latitude     float64
	Longitude    float64
}

// ipbaseResponse is the json response from ipbase.com
type ipbaseResponse struct {
	Data data `json:"data"`
}

// IPBaseError is the json response when there is an error from ipbase.com
type IPBaseError struct {
	Message string `json:"message"`
}

func (e *IPBaseError) Error() string {
	return e.Message
}

// ErrInvalidIPAddress is the error raised when an invalid IP address is provided
var ErrInvalidIPAddress = errors.New("invalid IP address provided")

// validateIPAddress ensures valid ip address
func validateIPAddress(ipAddress string) error {
	if net.ParseIP(ipAddress) == nil {
		return ErrInvalidIPAddress
	}
	return nil
}

// getIPGeoData fetches geolocation data for specified IP address from https://ipbase.com
func (g *Geofence) getIPGeoData(ipAddress string) (*ipbaseResponse, error) {
	response := &ipbaseResponse{}
	ipbaseError := &IPBaseError{}

	resp, err := g.ipbaseClient.R().
		SetHeader("Accept", "application/json").
		SetQueryParam("apikey", g.Config.Token).
		SetQueryParam("ip", ipAddress).
		SetResult(response).
		SetError(ipbaseError).
		Get("/info")
	if err != nil {
		return response, err
	}

	// If api gives back status code >399, report error to user
	if resp.IsError() {
		return response, ipbaseError
	}

	return resp.Result().(*ipbaseResponse), nil
}

// New creates a new geofence for the IP address specified.
// Use "" as the ip address to geofence the machine your application is running on
// Token comes from https://ipbase.com/
func New(c *Config) (*Geofence, error) {
	// Create new client for ipbase.com
	ipbaseClient := resty.New().SetBaseURL(ipBaseBaseURL)

	// New Geofence object
	geofence := &Geofence{
		Config:       *c,
		ipbaseClient: ipbaseClient,
		ctx:          context.Background(),
	}

	// Set up redis client if options are provided
	// else we create a local in-memory cache
	if c.RedisOptions != nil {
		c.RedisOptions.TTL = c.CacheTTL
		if c.CacheTTL < 0 {
			c.RedisOptions.TTL = 0
		}
		geofence.cache = cache.NewRedisCache(c.RedisOptions)
	} else {
		geofence.cache = cache.NewMemoryCache(&cache.MemoryOptions{
			TTL: c.CacheTTL,
		})
	}

	// Get current location of specified IP address
	// If empty string, use public IP of device running this
	// or use location of the specified IP
	ipAddressLookupDetails, err := geofence.getIPGeoData(c.IPAddress)
	if err != nil {
		return geofence, err
	}

	// Set the location of our geofence to compare against looked up IP's
	geofence.Latitude = ipAddressLookupDetails.Data.Location.Latitude
	geofence.Longitude = ipAddressLookupDetails.Data.Location.Longitude

	return geofence, nil
}

// IsIPAddressNear returns true if the specified address is within proximity
func (g *Geofence) IsIPAddressNear(ipAddress string) (bool, error) {
	// Ensure IP is valid first
	err := validateIPAddress(ipAddress)
	if err != nil {
		return false, err
	}

	if g.Config.AllowPrivateIPAddresses {
		ip := net.ParseIP(ipAddress)
		if ip.IsPrivate() || ip.IsLoopback() {
			return true, nil
		}
	}

	// Check if ipaddress has been looked up before and is in cache
	isIPAddressNear, found, err := g.cache.Get(g.ctx, ipAddress)
	if err != nil {
		return false, err
	}

	if found {
		return isIPAddressNear, nil
	}

	// If not in cache, lookup IP and compare
	ipAddressLookupDetails, err := g.getIPGeoData(ipAddress)
	if err != nil {
		return false, err
	}

	// Format our IP coordinates and the clients
	currentCoordinates := geo.NewCoordinatesFromDegrees(g.Latitude, g.Longitude)
	clientCoordinates := geo.NewCoordinatesFromDegrees(ipAddressLookupDetails.Data.Location.Latitude, ipAddressLookupDetails.Data.Location.Longitude)

	// Get distance in kilometers
	distance := currentCoordinates.Distance(clientCoordinates)

	// Compare coordinates
	// Distance must be less than or equal to the configured radius to be near
	isNear := distance <= g.Config.Radius

	err = g.cache.Set(g.ctx, ipAddress, isNear)
	if err != nil {
		return false, err
	}

	return isNear, nil
}
