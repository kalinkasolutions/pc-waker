package main

import (
	"log"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort      = 8080
	defaultBroadcast = "255.255.255.255"
	defaultWolPort   = 9
	defaultCfgPath   = "/config/config.yaml"
)

// Config is the top-level structure of the mounted config.yaml.
type Config struct {
	Port      int    `yaml:"port"`
	Broadcast string `yaml:"broadcast"`
	WolPort   int    `yaml:"wol_port"`
	Hosts     []Host `yaml:"hosts"`
}

// Host is a single wakeable machine.
type Host struct {
	Name      string `yaml:"name"`
	MAC       string `yaml:"mac"`
	Broadcast string `yaml:"broadcast"` // optional per-host override

	// hw is the parsed MAC, populated at load time.
	hw net.HardwareAddr `yaml:"-"`
}

// loadConfig reads and validates the YAML config, applying defaults and
// skipping (with a log line) any host whose MAC does not parse.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.Broadcast == "" {
		cfg.Broadcast = defaultBroadcast
	}
	if cfg.WolPort == 0 {
		cfg.WolPort = defaultWolPort
	}

	valid := cfg.Hosts[:0]
	for _, h := range cfg.Hosts {
		hw, err := net.ParseMAC(h.MAC)
		if err != nil {
			log.Printf("skipping host %q: invalid MAC %q: %v", h.Name, h.MAC, err)
			continue
		}
		h.hw = hw
		if h.Broadcast == "" {
			h.Broadcast = cfg.Broadcast
		}
		valid = append(valid, h)
	}
	cfg.Hosts = valid

	return &cfg, nil
}
