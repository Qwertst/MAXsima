package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
)

// Config holds the application configuration parsed from CLI flags.
type Config struct {
	Username    string
	Port        int
	PeerAddress string
}

// Parse parses CLI flags from os.Args and validates the resulting config.
func Parse() (*Config, error) {
	cfg := &Config{}
	flag.StringVar(&cfg.Username, "username", "", "your display name")
	flag.IntVar(&cfg.Port, "port", 0, "port to listen on (server mode)")
	flag.StringVar(&cfg.PeerAddress, "peer", "", "peer address to connect to (client mode)")
	flag.Parse()
	return cfg, cfg.Validate()
}

// ParseFlags parses the given args slice (without program name) and returns a Config.
// It does not validate — call Validate() separately if needed.
func ParseFlags(args []string) *Config {
	cfg := &Config{}
	fs := flag.NewFlagSet("maxsima", flag.ContinueOnError)
	fs.StringVar(&cfg.Username, "username", "", "your display name")
	fs.IntVar(&cfg.Port, "port", 0, "port to listen on (server mode)")
	fs.StringVar(&cfg.PeerAddress, "peer", "", "peer address to connect to (client mode)")
	_ = fs.Parse(args)
	return cfg
}

// IsServerMode returns true when no peer address is configured.
func (c *Config) IsServerMode() bool {
	return c.PeerAddress == ""
}

// Validate checks that the config is valid for starting the application.
func (c *Config) Validate() error {
	if c.Username == "" {
		return errors.New("--username is required")
	}
	if c.PeerAddress == "" && c.Port == 0 {
		return errors.New("either --port (server) or --peer (client) must be specified")
	}
	if c.PeerAddress == "" && (c.Port < 1 || c.Port > 65535) {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", c.Port)
	}
	return nil
}

// ValidatePeerAddress validates the format of PeerAddress (host:port).
func (c *Config) ValidatePeerAddress() error {
	if c.PeerAddress == "" {
		return errors.New("peer address is empty")
	}
	host, portStr, err := net.SplitHostPort(c.PeerAddress)
	if err != nil {
		return fmt.Errorf("invalid peer address %q: %w", c.PeerAddress, err)
	}
	if host == "" {
		return fmt.Errorf("invalid peer address %q: missing host", c.PeerAddress)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid peer address %q: invalid port %q", c.PeerAddress, portStr)
	}
	return nil
}
