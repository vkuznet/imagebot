package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Configuration stores server configuration parameters
type Configuration struct {
	Port          int      `json:"port"`          // server port number
	Base          string   `json:"base"`          // base URL
	Verbose       int      `json:"verbose"`       // verbose output
	ServerCrt     string   `json:"serverCrt"`     // path to server crt file
	ServerKey     string   `json:"serverKey"`     // path to server key file
	Secret        string   `json:"secret"`        // secret passphrase for encoding/decoding tokens
	Namespaces    []string `json:"namespaces"`    // list allowed namespaces
	Services      []string `json:"services"`      // list allowed services
	Repositories  []string `json:"repositories"`  // list allowed repositories
	UTC           bool     `json:"utc"`           // use UTC for logging or not
	Token         string   `json:"token"`         // authorization token
	MonitRecord   bool     `json:"monitRecord"`   // print on stdout monit record
	TokenInterval int64    `json:"tokenInterval"` // token validity interval in seconds
}

// Config variable represents configuration object
var Config Configuration

// helper function to parse configuration
func parseConfig(configFile string) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to read", err)
		return err
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		log.Println("Unable to parse", err)
		return err
	}
	if Config.TokenInterval == 0 {
		Config.TokenInterval = 60 // default 1 minute
	}
	return nil
}
