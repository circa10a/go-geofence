# go-geofence

![GitHub tag (latest semver)](https://img.shields.io/github/v/tag/circa10a/go-geofence?style=plastic)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/circa10a/go-geofence)](https://pkg.go.dev/github.com/circa10a/go-geofence?tab=overview)
[![Go Report Card](https://goreportcard.com/badge/github.com/circa10a/go-geofence)](https://goreportcard.com/report/github.com/circa10a/go-geofence)

A small library to detect if an IP address is close to yours or another of your choosing using https://freegeoip.app/

## Usage

First you will need a free API Token from [freegeoip.app](https://freegeoip.app/)

```bash
go get github.com/circa10a/go-geofence
```

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/circa10a/go-geofence"
)

func main() {
	geofence, err := geofence.New(&geofence.Config{
		// Empty string to geofence your current public IP address, or you can monitor a remote address by supplying it as the first parameter
		IPAddress: "",
		// freegeoip.app API token
		Token: "YOUR_FREEGEOIP_API_TOKEN",
		// Maximum radius of the geofence in kilometers, only clients less than or equal to this distance will return true with isAddressNearby
		// 1 kilometer
		Radius: 1.0,
		// How long to cache if any ip address is nearby
		CacheTTL: 7 * (24 * time.Hour), // 1 week
	})
	if err != nil {
		log.Fatal(err)
	}
	isAddressNearby, err := geofence.IsIPAddressNear("8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}
	// Address nearby: false
	fmt.Println("Address nearby: ", isAddressNearby)
}
```

### Private IP Addresses

Private IP's will always result in `false` since their coordinates will come back as `0.0000`. You can work around these like so:

```go
package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/circa10a/go-geofence"
)

func main() {
	// Provide an IP Address to test with
	ipAddress := "192.168.1.100"
	geofence, err := geofence.New(&geofence.Config{
		// Empty string to geofence your current public IP address, or you can monitor a remote address by supplying it as the first parameter
		IPAddress: ipAddress,
		// freegeoip.app API token
		Token: "YOUR_FREEGEOIP_API_TOKEN",
		// Maximum radius of the geofence in kilometers, only clients less than or equal to this distance will return true with isAddressNearby
		// 1 kilometer
		Radius: 1.0,
		// How long to cache if any ip address is nearby
		CacheTTL: 7 * (24 * time.Hour), // 1 week
	})
	if err != nil {
		log.Fatal(err)
	}
	// Skip Private IP analysis as it will always be false
	ip := net.ParseIP(ipAddress)
	if !ip.IsPrivate() || ip.IsLoopback() {
		isAddressNearby, err := geofence.IsIPAddressNear(ipAddress)
		if err != nil {
			log.Fatal(err)
		}
		// Address nearby: false
		fmt.Println("Address nearby: ", isAddressNearby)
	}
}
```
