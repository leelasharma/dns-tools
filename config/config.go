package config

import (
	"fmt"
	"io/ioutil"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

var (
	regexHostname = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])\.$`)
)

// ManagedZoneDefaults holds the default configuration fo  managed zones
type ManagedZoneDefaults struct {
	TTL int
	//	SOA  ManagedZoneSOAConfig
}

// ManagedZoneConfig holds a managed zone's configuration
type ManagedZoneConfig struct {
	FQDN string
	TTL  int
	//	SOA  ManagedZoneSOAConfig
}

/*
// ManagedZoneSOAConfig holds the Start Of Authority configuration of a
// managed zone
type ManagedZoneSOAConfig struct {
	TTL     int32 <-- must always be 0 or positive, bit only holds values up to 2^32-1 according to RFC 2181
	Refresh int
	Retry   int
	Expire  int
	NegTTL  int32
}
*/

// Config holds the dns-tools configuration
type Config struct {
	ZoneDataDirectory string
	Defaults          ManagedZoneDefaults
	ManagedZones      []ManagedZoneConfig
}

// YAMLConfig holds exactly one dns-tools configuration
type YAMLConfig struct {
	Config Config
}

// New creates a new database instance
func New(fname string) (*Config, error) {
	yamlConfig := YAMLConfig{}
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	err = yaml.UnmarshalStrict(data, &yamlConfig)
	if err != nil {
		return nil, err
	}
	config := yamlConfig.Config

	// verify defaults
	err = checkTTL(config.Defaults.TTL)
	if err != nil {
		return nil, fmt.Errorf("defaults: %v", err)
	}
	// checkSOA()

	// verify individual managed zones
	seen := make(map[string]bool)
	for idx := range config.ManagedZones {
		mz := &config.ManagedZones[idx] // uses actual data, not local copy
		// apply defaults to unset values
		if mz.TTL == 0 {
			mz.TTL = config.Defaults.TTL
		}
		// check name
		err = checkFQDN(mz.FQDN)
		if err != nil {
			return nil, fmt.Errorf("managed zone `%v`: %v", mz.FQDN, err)
		}
		// check managed zone default TTL
		err = checkTTL(mz.TTL)
		if err != nil {
			return nil, fmt.Errorf("managed zone `%v`: %v", mz.FQDN, err)
		}
		// check for duplicate zones
		if _, ok := seen[mz.FQDN]; ok {
			return nil, fmt.Errorf("managed zone `%v`: duplicate entry", mz.FQDN)
		}
		seen[mz.FQDN] = true
	}
	return &config, nil
}

func checkFQDN(fqdn string) error {
	if !regexHostname.MatchString(fqdn) {
		return fmt.Errorf("invalid FQDN: %v", fqdn)
	}
	return nil
}

func checkTTL(ttl int) error {
	if ttl < 1 || ttl > 2147483647 {
		return fmt.Errorf("invalid TTL: %v", ttl)
	}
	return nil
}
