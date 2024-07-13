// Look.
// I know people hate types.go and want to keep the structs
// but damn these are some exaustive types and it flooded the primary package logic

package geofence

type rangeType struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type connection struct {
	Organization string `json:"organization"`
	Isp          string `json:"isp"`
	Range        string `json:"range"`
	Asn          int    `json:"asn"`
}

type continent struct {
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
	WikidataID     string `json:"wikidata_id"`
	Code           int    `json:"code"`
	GeonamesID     int    `json:"geonames_id"`
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
	Fips              string       `json:"fips"`
	Alpha3            string       `json:"alpha3"`
	WikidataID        string       `json:"wikidata_id"`
	HascID            string       `json:"hasc_id"`
	Emoji             string       `json:"emoji"`
	Ioc               string       `json:"ioc"`
	Alpha2            string       `json:"alpha2"`
	Name              string       `json:"name"`
	NameTranslated    string       `json:"name_translated"`
	Languages         []languages  `json:"languages"`
	Timezones         []string     `json:"timezones"`
	Currencies        []currencies `json:"currencies"`
	CallingCodes      []string     `json:"calling_codes"`
	GeonamesID        int          `json:"geonames_id"`
	IsInEuropeanUnion bool         `json:"is_in_european_union"`
}

type city struct {
	Alpha2         any    `json:"alpha2"`
	HascID         any    `json:"hasc_id"`
	Fips           string `json:"fips"`
	WikidataID     string `json:"wikidata_id"`
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
	GeonamesID     int    `json:"geonames_id"`
}

type region struct {
	Fips           string `json:"fips"`
	Alpha2         string `json:"alpha2"`
	HascID         string `json:"hasc_id"`
	WikidataID     string `json:"wikidata_id"`
	Name           string `json:"name"`
	NameTranslated string `json:"name_translated"`
	GeonamesID     int    `json:"geonames_id"`
}

type location struct {
	Zip        string    `json:"zip"`
	City       city      `json:"city"`
	Region     region    `json:"region"`
	Continent  continent `json:"continent"`
	Country    country   `json:"country"`
	GeonamesID int       `json:"geonames_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
}

type timezone struct {
	ID               string `json:"id"`
	CurrentTime      string `json:"current_time"`
	Code             string `json:"code"`
	IsDaylightSaving bool   `json:"is_daylight_saving"`
	GmtOffset        int    `json:"gmt_offset"`
}

type security struct {
	IsAnonymous     bool `json:"is_anonymous"`
	IsDatacenter    bool `json:"is_datacenter"`
	IsVpn           bool `json:"is_vpn"`
	IsBot           bool `json:"is_bot"`
	IsAbuser        bool `json:"is_abuser"`
	IsKnownAttacker bool `json:"is_known_attacker"`
	IsProxy         bool `json:"is_proxy"`
	IsSpam          bool `json:"is_spam"`
	IsTor           bool `json:"is_tor"`
	IsIcloudRelay   bool `json:"is_icloud_relay"`
	ThreatScore     int  `json:"threat_score"`
}

type domains struct {
	Domains []string `json:"domains"`
	Count   int      `json:"count"`
}

type data struct {
	RangeType  rangeType  `json:"range_type"`
	IP         string     `json:"ip"`
	Hostname   string     `json:"hostname"`
	Type       string     `json:"type"`
	Connection connection `json:"connection"`
	Tlds       []string   `json:"tlds"`
	Timezone   timezone   `json:"timezone"`
	Domains    domains    `json:"domains"`
	Location   location   `json:"location"`
	Security   security   `json:"security"`
}
