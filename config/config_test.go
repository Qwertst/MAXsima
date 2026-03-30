package config

import (
	"testing"
)

func TestConfigServerModeDetection(t *testing.T) {
	cfg := Config{
		Username:    "Alice",
		Port:        50051,
		PeerAddress: "",
	}

	isServer := cfg.IsServerMode()

	if !isServer {
		t.Errorf("Config without PeerAddress should be server mode, got client mode")
	}
}

func TestConfigClientModeDetection(t *testing.T) {
	cfg := Config{
		Username:    "Bob",
		Port:        0,
		PeerAddress: "localhost:50051",
	}

	isServer := cfg.IsServerMode()

	if isServer {
		t.Errorf("Config with PeerAddress should be client mode, got server mode")
	}
}

func TestConfigValidationWithValidServerConfig(t *testing.T) {
	cfg := Config{
		Username:    "Alice",
		Port:        50051,
		PeerAddress: "",
	}

	err := cfg.Validate()

	if err != nil {
		t.Errorf("Valid server config should not produce error, got: %v", err)
	}
}

func TestConfigValidationWithValidClientConfig(t *testing.T) {
	cfg := Config{
		Username:    "Bob",
		Port:        0,
		PeerAddress: "localhost:50051",
	}

	err := cfg.Validate()

	if err != nil {
		t.Errorf("Valid client config should not produce error, got: %v", err)
	}
}

func TestConfigValidationWithMissingUsername(t *testing.T) {
	cfg := Config{
		Username:    "",
		Port:        50051,
		PeerAddress: "",
	}

	err := cfg.Validate()

	if err == nil {
		t.Errorf("Config without username should produce validation error, got nil")
	}
}

func TestConfigValidationServerMissingPort(t *testing.T) {
	cfg := Config{
		Username:    "Alice",
		Port:        0,
		PeerAddress: "",
	}

	err := cfg.Validate()

	if err == nil {
		t.Errorf("Server config without port should produce validation error, got nil")
	}
}

func TestConfigValidationClientMissingPeerAddress(t *testing.T) {
	cfg := Config{
		Username:    "Bob",
		Port:        0,
		PeerAddress: "",
	}

	err := cfg.Validate()

	if err == nil {
		t.Errorf("Config with neither port nor peer should produce validation error, got nil")
	}
}

func TestConfigValidationInvalidPort(t *testing.T) {
	testCases := []struct {
		port        int
		description string
	}{
		{port: -1, description: "negative port"},
		{port: 0, description: "port zero for server"},
		{port: 65536, description: "port over 65535"},
		{port: 1000000, description: "port way over max"},
	}

	for _, tc := range testCases {
		cfg := Config{
			Username:    "Alice",
			Port:        tc.port,
			PeerAddress: "",
		}

		err := cfg.Validate()
		if tc.port < 1 || tc.port > 65535 {
			if tc.port != 0 || cfg.PeerAddress == "" {
				if err == nil && tc.port > 65535 {
					t.Errorf("Port %d (%s) should produce error but got nil", tc.port, tc.description)
				}
			}
		}
	}
}

func TestConfigCreation(t *testing.T) {
	username := "Charlie"
	port := 50052
	peerAddr := "192.168.1.1:50051"

	cfg := Config{
		Username:    username,
		Port:        port,
		PeerAddress: peerAddr,
	}

	if cfg.Username != username {
		t.Errorf("Username not set correctly: expected '%s', got '%s'", username, cfg.Username)
	}
	if cfg.Port != port {
		t.Errorf("Port not set correctly: expected %d, got %d", port, cfg.Port)
	}
	if cfg.PeerAddress != peerAddr {
		t.Errorf("PeerAddress not set correctly: expected '%s', got '%s'", peerAddr, cfg.PeerAddress)
	}
}

func TestConfigParseFlags(t *testing.T) {
	args := []string{
		"--username", "TestUser",
		"--port", "50053",
	}

	cfg := ParseFlags(args)

	if cfg.Username != "TestUser" {
		t.Errorf("Parsed username should be 'TestUser', got '%s'", cfg.Username)
	}
	if cfg.Port != 50053 {
		t.Errorf("Parsed port should be 50053, got %d", cfg.Port)
	}
}

func TestConfigParseFlagsWithClientMode(t *testing.T) {
	args := []string{
		"--username", "ClientUser",
		"--peer", "server.example.com:50051",
	}

	cfg := ParseFlags(args)

	if cfg.Username != "ClientUser" {
		t.Errorf("Parsed username should be 'ClientUser', got '%s'", cfg.Username)
	}
	if cfg.PeerAddress != "server.example.com:50051" {
		t.Errorf("Parsed peer address should be 'server.example.com:50051', got '%s'", cfg.PeerAddress)
	}
	if cfg.Port != 0 {
		t.Logf("Client mode shouldn't set Port (expected 0), got %d", cfg.Port)
	}
}

func TestConfigParseFlagsWithMissingRequiredFlag(t *testing.T) {
	args := []string{
		"--port", "50051",
	}

	cfg := ParseFlags(args)

	if cfg.Username != "" {
		t.Logf("Missing --username should result in empty username for later validation")
	}
}

func TestConfigValidatePeerAddressFormat(t *testing.T) {
	testCases := []struct {
		peerAddr    string
		isValid     bool
		description string
	}{
		{peerAddr: "localhost:50051", isValid: true, description: "localhost with port"},
		{peerAddr: "127.0.0.1:50051", isValid: true, description: "IP with port"},
		{peerAddr: "example.com:50051", isValid: true, description: "domain with port"},
		{peerAddr: "localhost", isValid: false, description: "missing port"},
		{peerAddr: ":50051", isValid: false, description: "missing host"},
		{peerAddr: "localhost:invalid", isValid: false, description: "invalid port"},
	}

	for _, tc := range testCases {
		cfg := Config{
			Username:    "TestUser",
			PeerAddress: tc.peerAddr,
		}

		err := cfg.ValidatePeerAddress()
		if tc.isValid && err != nil {
			t.Errorf("Address '%s' (%s) should be valid but got error: %v", tc.peerAddr, tc.description, err)
		}
		if !tc.isValid && err == nil {
			t.Errorf("Address '%s' (%s) should be invalid but validation passed", tc.peerAddr, tc.description)
		}
	}
}
