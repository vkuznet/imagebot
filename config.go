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
	RootCAs       string   `json:"rootCAs"`       // server Root CAs path
	ReadTimeout   int      `json:"read_timeout"`  // server read timeout in sec
	WriteTimeout  int      `json:"write_timeout"` // server write timeout in sec
	Secret        string   `json:"secret"`        // secret passphrase for encoding/decoding tokens
	Namespaces    []string `json:"namespaces"`    // list allowed namespaces
	Services      []string `json:"services"`      // list allowed services
	Images        []string `json:"images"`        // list of allowed docker hub images
	UTC           bool     `json:"utc"`           // use UTC for logging or not
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
	// default values
	if Config.Port == 0 {
		Config.Port = 8111
	}
	if Config.TokenInterval == 0 {
		Config.TokenInterval = 60 // default 1 minute
	}
	if Config.ReadTimeout == 0 {
		Config.ReadTimeout = 60
	}
	if Config.WriteTimeout == 0 {
		Config.WriteTimeout = 60
	}
	return nil
}
