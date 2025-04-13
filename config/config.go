package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ApiUrl       string        `yaml:"api_url"`
	Username     string        `yaml:"username"`
	Password     string        `yaml:"password"`
	LogLevel     string        `yaml:"log_level"`
	Timeout      time.Duration `yaml:"timeout"`
	AutoShutdown bool          `yaml:"auto_shutdown"`
	Addresses    []ServerType  `yaml:"addresses"`
}

type ServerType struct {
	Protocol   string `yaml:"protocol"`
	Listener   Host   `yaml:"listener"`
	CraftyHost Host   `yaml:"crafty_host"`
}

type Host struct {
	Addr string `yaml:"addr"`
	Port int    `yaml:"port"`
}

// NewConfig returns default config
func NewConfig() Config {
	return Config{
		ApiUrl:       "https://crafty:8443",
		Username:     "admin",
		Password:     "password",
		LogLevel:     "INFO",
		Timeout:      time.Minute * 5,
		AutoShutdown: true,
		Addresses: []ServerType{
			{
				Protocol: "tcp",
				Listener: Host{
					Addr: "127.0.0.1",
					Port: 25565,
				},
				CraftyHost: Host{
					Addr: "crafty",
					Port: 25565,
				},
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

			log.Printf("config file not found — created default at %s\n", path)
			return nil
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
