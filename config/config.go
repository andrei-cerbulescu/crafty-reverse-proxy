package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type OthersType struct {
	ExternalPort string `yaml:"external_port"`
	ExternalIp   string `yaml:"external_ip"`
	InternalIp   string `yaml:"internal_ip"`
	InternalPort string `yaml:"internal_port"`
	Protocol     string `yaml:"protocol"`
}

type ServerType struct {
	ExternalPort string       `yaml:"external_port"`
	ExternalIp   string       `yaml:"external_ip"`
	InternalIp   string       `yaml:"internal_ip"`
	InternalPort string       `yaml:"internal_port"`
	Protocol     string       `yaml:"protocol"`
	Others       []OthersType `yaml:"others"`
}

type Config struct {
	ApiUrl       string        `yaml:"api_url"`
	Username     string        `yaml:"username"`
	Password     string        `yaml:"password"`
	Timeout      time.Duration `yaml:"timeout"`
	AutoShutdown bool          `yaml:"auto_shutdown"`
	Addresses    []ServerType  `yaml:"addresses"`
}

// NewConfig returns default config
func NewConfig() Config {
	return Config{
		ApiUrl:       "https://crafty:8443",
		Username:     "admin",
		Password:     "password",
		Timeout:      time.Minute * 5,
		AutoShutdown: false,
		Addresses: []ServerType{
			{
				ExternalPort: "3120",
				ExternalIp:   "craftyreverseproxy",
				InternalIp:   "crafty",
				InternalPort: "25565",
				Protocol:     "tcp",
			},
		},
	}
}

// Load loads a config from given path.
// It will create default config if there is no config file.
func (c *Config) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			defaultConfig := NewConfig()
			data, marshalErr := yaml.Marshal(defaultConfig)
			if marshalErr != nil {
				return fmt.Errorf("failed to marshal default config: %w", marshalErr)
			}

			writeErr := os.WriteFile(path, data, 0644)
			if writeErr != nil {
				return fmt.Errorf("failed to write default config file: %w", writeErr)
			}

			return fmt.Errorf("config file not found — created default at %s", path)
		}

		return fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("could not read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("could not parse yaml config: %w", err)
	}

	return nil
}
