package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr           string
	NodeID             string
	BootstrapToken     string
	NodeToken          string
	PanelURL           string
	StateDir           string
	XrayBin            string
	LogLevel           string
	HeartbeatInterval  time.Duration
	ConfigPollInterval time.Duration
	TLSEnabled         bool
}

func Load() (Config, error) {
	heartbeatInterval, err := durationFromEnv("LENKER_AGENT_HEARTBEAT_INTERVAL", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	configPollInterval, err := durationFromEnv("LENKER_AGENT_CONFIG_POLL_INTERVAL", 30*time.Second)
	if err != nil {
		return Config{}, err
	}

	tlsEnabled, err := boolFromEnv("LENKER_AGENT_TLS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:           stringFromEnv("LENKER_AGENT_HTTP_ADDR", ":8090"),
		NodeID:             strings.TrimSpace(os.Getenv("LENKER_AGENT_NODE_ID")),
		BootstrapToken:     strings.TrimSpace(os.Getenv("LENKER_AGENT_BOOTSTRAP_TOKEN")),
		NodeToken:          strings.TrimSpace(os.Getenv("LENKER_AGENT_NODE_TOKEN")),
		PanelURL:           strings.TrimRight(strings.TrimSpace(os.Getenv("LENKER_AGENT_PANEL_URL")), "/"),
		StateDir:           stringFromEnv("LENKER_AGENT_STATE_DIR", ".lenker-node-agent"),
		XrayBin:            strings.TrimSpace(os.Getenv("LENKER_AGENT_XRAY_BIN")),
		LogLevel:           stringFromEnv("LENKER_AGENT_LOG_LEVEL", "info"),
		HeartbeatInterval:  heartbeatInterval,
		ConfigPollInterval: configPollInterval,
		TLSEnabled:         tlsEnabled,
	}

	if cfg.HeartbeatInterval <= 0 {
		return Config{}, errors.New("LENKER_AGENT_HEARTBEAT_INTERVAL must be positive")
	}
	if cfg.ConfigPollInterval <= 0 {
		return Config{}, errors.New("LENKER_AGENT_CONFIG_POLL_INTERVAL must be positive")
	}

	return cfg, nil
}

func stringFromEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration, nil
	}

	seconds, atoiErr := strconv.Atoi(value)
	if atoiErr != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}

func boolFromEnv(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseBool(value)
}
