# go-geofence

![GitHub tag (latest semver)](https://img.shields.io/github/v/tag/circa10a/go-geofence?style=plastic)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/circa10a/go-geofence)](https://pkg.go.dev/github.com/circa10a/go-geofence?tab=overview)
[![Go Report Card](https://goreportcard.com/badge/github.com/circa10a/go-geofence)](https://goreportcard.com/report/github.com/circa10a/go-geofence)

A small library to detect if an IP address is close to yours or another of your choosing using https://ipbase.com/

## Usage

First you will need a free API Token from [ipbase.com](https://ipbase.com/)

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
		// ipbase.com API token
		Token: "YOUR_IPBASE_API_TOKEN",
		// Maximum radius of the geofence in kilometers, only clients less than or equal to this distance will return true with IsIPAddressNear()
		// 1 kilometer
		Radius: 1.0,
		// Allow 192.X, 172.X, 10.X and loopback addresses
		AllowPrivateIPAddresses: true,
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

## Caching

To cache keys indefinitely, set `CacheTTL: -1`

### Local (in-memory)

By default, the library will use an in-memory cache that will be used to reduce the number of calls to ipbase.com and increase performance. If no `CacheTTL` value is set (`0`), the in-memory cache is disabled.

### Persistent

If you need a persistent cache to live outside of your application, [Redis](https://redis.io/) is supported by this library. To have the library cache address proximity using a Redis instance, simply provide a `geofence.RedisOptions` struct to `geofence.Config.RedisOptions`. If `RedisOptions` is configured, the in-memory cache will not be used.

> Note: Only Redis 7 is currently supported at the time of this writing.

#### Example Redis Usage

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/circa10a/go-geofence"
	"github.com/go-redis/redis/v9"
)

func main() {
	geofence, err := geofence.New(&geofence.Config{
		IPAddress: "",
		Token: "YOUR_IPBASE_API_TOKEN",
		Radius: 1.0,
		AllowPrivateIPAddresses: true,
		CacheTTL: 7 * (24 * time.Hour), // 1 week
		// Use Redis for caching
		RedisOptions: &geofence.RedisOptions{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		},
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
