package config

import (
	"encoding/json"
	"io"
	"os"
)

type OthersType struct {
	ExternalPort string `json:"external_port"`
	ExternalIp   string `json:"external_ip"`
	InternalIp   string `json:"internal_ip"`
	InternalPort string `json:"internal_port"`
	Protocol     string `json:"protocol"`
}

type ServerType struct {
	ExternalPort string       `json:"external_port"`
	ExternalIp   string       `json:"external_ip"`
	InternalIp   string       `json:"internal_ip"`
	InternalPort string       `json:"internal_port"`
	Protocol     string       `json:"protocol"`
	Others       []OthersType `json:"others"`
}

type Config struct {
	ApiUrl       string       `json:"api_url"`
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	Timeout      int          `json:"timeout"`
	AutoShutdown bool         `json:"auto_shutdown"`
	Addresses    []ServerType `json:"addresses"`
}

func loadConfig() Config {
	file, err := os.Open("./config.json")
	if err != nil {
		_, err = os.Create("./config.json")
		if err != nil {
			panic("Could not open config\n")
		}

		err = os.Chmod("./config.json", 0755)
		if err != nil {
			panic("Created a file but could not chmod it\n")
		}

		panic("Could not open config\nCreated one config file before exiting\n")
	}

	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic("Could not read config\n")
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)

	if err != nil {
		panic("Could not parse config\n")
	}

	return config
}

var singleConfigInstance *Config

func GetConfig() *Config {
	if singleConfigInstance == nil {
		var config = loadConfig()
		singleConfigInstance = &config
	}

	return singleConfigInstance
}
