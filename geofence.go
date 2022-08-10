package geofence

import (
	"fmt"
	"net"
	"time"

	"github.com/EpicStep/go-simple-geo/v2/geo"
	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"
)

const (
	ipBaseBaseURL                   = "https://api.ipbase.com/v2"
	deleteExpiredCacheItemsInternal = 10 * time.Minute
)

// Config holds the user configuration to setup a new geofence
type Config struct {
	IPAddress               string
	Token                   string
	Radius                  float64
	AllowPrivateIPAddresses bool
	CacheTTL                time.Duration
}

// Geofence holds a ipbase.com client, cache and user supplied config
type Geofence struct {
	Cache        *cache.Cache
	IPBaseClient *resty.Client
	Config
	Latitude  float64
	Longitude float64
}

// ipBaseResponse is the json response from ipbase.com
type ipBaseResponse struct {
	Data data `json:"data"`
}

type timezone struct {
	Id              string `json:"id"`
	CurrentTime     string `json:"current_time"`
	Code            string `json:"code"`
	IDaylightSaving bool   `json:"is_daylight_saving"`
	GmtOffset       int    `json:"gmt_offset"`
}

type connection struct {
	Organization string `json:"organization"`
	Isp          string `json:"isp"`
	Asn          int    `json:"asn"`
}

type continent struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
}

type currencies struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	SymbolNative  string `json:"symbol_native"`
	Code          string `json:"code"`
	NamePlural    string `json:"name_plural"`
	DecimalDigits int    `json:"decimal_digits"`
	Rounding      int    `json:"rounding"`
}

type languages struct {
	Name       string `json:"name"`
	NameNative string `json:"name_native"`
}
type country struct {
	Alpha2            string       `json:"alpha2"`
	Alpha3            string       `json:"alpha3"`
	CallingCodes      []string     `json:"calling_codes"`
	Currencies        []currencies `json:"currencies"`
	Emoji             string       `json:"emoji"`
	Ioc               string       `json:"ioc"`
	Languages         []languages  `json:"languages"`
	Name              string       `json:"name"`
	NameTranslated    string       `json:"name_translated"`
	Timezones         []string     `json:"timezones"`
	IsInEuropeanUnion bool         `json:"is_in_european_union"`
}

type city struct {
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
}

type region struct {
	Fips           interface{} `json:"fips"`
	Alpha2         interface{} `json:"alpha2"`
	Name           string      `json:"name"`
	NameTranslated string      `json:"name_translated"`
}

type location struct {
	GeonamesID interface{} `json:"geonames_id"`
	Region     region      `json:"region"`
	Continent  continent   `json:"continent"`
	City       city        `json:"city"`
	Zip        string      `json:"zip"`
	Country    country     `json:"country"`
	Latitude   float64     `json:"latitude"`
	Longitude  float64     `json:"longitude"`
}

type data struct {
	Timezone   timezone   `json:"timezone"`
	IP         string     `json:"ip"`
	Type       string     `json:"type"`
	Connection connection `json:"connection"`
	Location   location   `json:"location"`
}

// IPBaseError is the json response when there is an error from ipbase.com
type IPBaseError struct {
	Message string `json:"message"`
}

func (e *IPBaseError) Error() string {
	return e.Message
}

// ErrInvalidIPAddress is the error raised when an invalid IP address is provided
var ErrInvalidIPAddress = fmt.Errorf("invalid IP address provided")

// validateIPAddress ensures valid ip address
func validateIPAddress(ipAddress string) error {
	if net.ParseIP(ipAddress) == nil {
		return ErrInvalidIPAddress
	}
	return nil
}

// getIPGeoData fetches geolocation data for specified IP address from https://ipbase.com
func (g *Geofence) getIPGeoData(ipAddress string) (*ipBaseResponse, error) {
	response := &ipBaseResponse{}
	ipBaseError := &IPBaseError{}

	resp, err := g.IPBaseClient.R().
		SetHeader("Accept", "application/json").
		SetQueryParam("apikey", g.Token).
		SetQueryParam("ip", ipAddress).
		SetResult(response).
		SetError(ipBaseError).
		Get("/info")
	if err != nil {
		return response, err
	}

	// If api gives back status code >399, report error to user
	if resp.IsError() {
		return response, ipBaseError
	}

	return resp.Result().(*ipBaseResponse), nil
}

// New creates a new geofence for the IP address specified.
// Use "" as the ip address to geofence the machine your application is running on
// Token comes from https://ipbase.com/
func New(c *Config) (*Geofence, error) {
	// Create new client for ipbase.com
	IPBaseClient := resty.New().SetBaseURL(ipBaseBaseURL)

	// New Geofence object
	geofence := &Geofence{
		Config:       *c,
		IPBaseClient: IPBaseClient,
		Cache:        cache.New(c.CacheTTL, deleteExpiredCacheItemsInternal),
	}

	// Get current location of specified IP address
	// If empty string, use public IP of device running this
	// Or use location of the specified IP
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

	if g.AllowPrivateIPAddresses {
		ip := net.ParseIP(ipAddress)
		if ip.IsPrivate() || ip.IsLoopback() {
			return true, nil
		}
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
	currentCoordinates := geo.NewCoordinatesFromDegrees(g.Latitude, g.Longitude)
	clientCoordinates := geo.NewCoordinatesFromDegrees(ipAddressLookupDetails.Data.Location.Latitude, ipAddressLookupDetails.Data.Location.Longitude)

	// Get distance in kilometers
	distance := currentCoordinates.Distance(clientCoordinates)

	// Compare coordinates
	// distance must be less than or equal to the configured radius to be near
	isNear := distance <= g.Radius

	// Insert ip address and it's status into the cache if user instantiated a cache
	g.Cache.Set(ipAddress, isNear, cache.DefaultExpiration)

	return isNear, nil
}
