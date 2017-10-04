// Package influx provides helper functions for parsing InfluxDB client
// configurations
package influx

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config holds a Influx client configuration
type Config struct {
	Server   string `json:"server"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Load parses a InfluxDB client configuration from a simple JSON file
func (conf *Config) Load(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("read Influx config file: %v", err)
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		return fmt.Errorf("parse Influx config file: %v", err)
	}
	return nil
}
