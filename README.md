# go-geofence

A small library to detect if an IP address is close to yours or another of your choosing using https://ipstack.com/

## Usage

First you will need a free API Token from [ipstack.com](https://ipstack.com/signup/free)

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
	// ipstack.com API token
	// Sensitivity
	// 0 - 111 km
	// 1 - 11.1 km
	// 2 - 1.11 km
	// 3 111 meters
	// 4 11.1 meters
	// 5 1.11 meters
	geofence, err := geofence.New("", "YOUR_IPSTACK.COM_API_TOKEN", 3)
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
