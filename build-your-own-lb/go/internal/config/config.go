package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all runtime settings for the load balancer
type Config struct {
	Port    int
	Profile string // dev OR prod
	LBType  string
	LBAlgo  string
}

// Load parses CLI flags and returns a validated Config
func Load() (*Config, error) {
	cfg := &Config{}

	// Create a FlagSet to parse flags passed after the binary
	fs := flag.NewFlagSet("loadbalancer", flag.ContinueOnError)

	fs.IntVar(&cfg.Port, "port", 8080, "Port to listen on")
	fs.StringVar(&cfg.Profile, "profile", "dev", "Runtime profile (dev, prod)")
	fs.StringVar(&cfg.LBType, "type", "application", "Type of LB (application, network)")
	fs.StringVar(&cfg.LBAlgo, "algo", "roundrobin", "LB algorithm (roundrobin, leastconn)")

	// Parse arguments (os.Args[1:] are the args after executable name)
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// Validate inputs before returning
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate ensures user input meets requirements
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	if c.Profile != "dev" && c.Profile != "prod" {
		return fmt.Errorf("profile must be dev OR prod, got %s", c.Profile)
	}

	if c.LBType != "application" && c.LBType != "network" {
		return fmt.Errorf("unsupported type %q (allowed: application, network)", c.LBType)
	}

	if c.LBAlgo != "roundrobin" && c.LBAlgo != "leastconn" {
		return fmt.Errorf("unsupported algorithm %q", c.LBAlgo)
	}

	return nil
}
