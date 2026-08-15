package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Monitors MonitorConfig `json:"monitors"`
}

type MonitorConfig struct {
	Odyssey MonitorSettings `json:"odyssey"`
	ProBike MonitorSettings `json:"probike"`
}

type MonitorSettings struct {
	Enabled bool `json:"enabled"`
}

func loadConfig() (Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config.json: %w", err)
	}

	var config Config

	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config.json: %w", err)
	}

	return config, nil
}