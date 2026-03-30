package config

import (
	"errors"
	"flag"
)

type Config struct {
	Username    string
	Port        int
	PeerAddress string
}

func Parse() (*Config, error) {
	cfg := &Config{}
	flag.StringVar(&cfg.Username, "username", "", "your display name")
	flag.IntVar(&cfg.Port, "port", 0, "port to listen on (server mode)")
	flag.StringVar(&cfg.PeerAddress, "peer", "", "peer address to connect to (client mode)")
	flag.Parse()
	return cfg, cfg.Validate()
}

func (c *Config) IsServerMode() bool {
	return c.PeerAddress == ""
}

func (c *Config) Validate() error {
	if c.Username == "" {
		return errors.New("--username is required")
	}
	if c.PeerAddress == "" && c.Port == 0 {
		return errors.New("either --port (server) or --peer (client) must be specified")
	}
	return nil
}
