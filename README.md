# go-geofence

![GitHub tag (latest semver)](https://img.shields.io/github/v/tag/circa10a/go-geofence?style=plastic)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/circa10a/go-geofence)](https://pkg.go.dev/github.com/circa10a/go-geofence?tab=overview)
[![Go Report Card](https://goreportcard.com/badge/github.com/circa10a/go-geofence)](https://goreportcard.com/report/github.com/circa10a/go-geofence)

A small library to detect if an IP address is close to yours or another of your choosing using https://freegeoip.app/

## Usage

First you will need a free API Token from [freegeoip.app]https://freegeoip.app/)

```bash
go get github.com/circa10a/go-geofence
```

```go
package main

import (
	"fmt"
	"log"

	"github.com/circa10a/go-geofence"
)

func main() {
	// Empty string to geofence your current public IP address, or you can monitor a remote address by supplying it as the first parameter
	// freegeoip.app API token
	// Sensitivity
	// 0 - 111 km
	// 1 - 11.1 km
	// 2 - 1.11 km
	// 3 111 meters
	// 4 11.1 meters
	// 5 1.11 meters
	geofence, err := geofence.New("", "YOUR_FREEGEOIP_API_TOKEN", 3)
	if err != nil {
		log.Fatal(err)
	}
	// Create cache that holds status in memory until application is restarted
	geofence.CreateCache(-1)
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
	"strings"

	"github.com/circa10a/go-geofence"
)

func main() {
	// Provide an IP Address to test with
	ipAddress := "192.168.1.100"
	geofence, err := geofence.New("", "YOUR_FREEGEOIP_API_TOKEN", 3)
	if err != nil {
		log.Fatal(err)
	}
	// Create cache that holds status in memory until application is restarted
	geofence.CreateCache(-1)
	if !strings.HasPrefix(ipAddress, "192.") && !strings.HasPrefix(ipAddress, "172.") && !strings.HasPrefix(ipAddress, "10.") {
		isAddressNearby, err := geofence.IsIPAddressNear(ipAddress)
		if err != nil {
			log.Fatal(err)
		}
		// Address nearby: false
		fmt.Println("Address nearby: ", isAddressNearby)
	}
}
```
