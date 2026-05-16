package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("LENKER_AGENT_HTTP_ADDR", "")
	t.Setenv("LENKER_AGENT_HEARTBEAT_INTERVAL", "")
	t.Setenv("LENKER_AGENT_CONFIG_POLL_INTERVAL", "")
	t.Setenv("LENKER_AGENT_XRAY_BIN", "")
	t.Setenv("LENKER_AGENT_TLS_ENABLED", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}
	if cfg.HTTPAddr != ":8090" {
		t.Fatalf("unexpected http addr: %q", cfg.HTTPAddr)
	}
	if cfg.HeartbeatInterval != 30*time.Second {
		t.Fatalf("unexpected heartbeat interval: %s", cfg.HeartbeatInterval)
	}
	if cfg.ConfigPollInterval != 30*time.Second {
		t.Fatalf("unexpected config poll interval: %s", cfg.ConfigPollInterval)
	}
	if cfg.TLSEnabled {
		t.Fatalf("expected tls disabled by default")
	}
}

func TestLoadEnv(t *testing.T) {
	t.Setenv("LENKER_AGENT_HTTP_ADDR", ":9999")
	t.Setenv("LENKER_AGENT_NODE_ID", "node-1")
	t.Setenv("LENKER_AGENT_BOOTSTRAP_TOKEN", "token")
	t.Setenv("LENKER_AGENT_NODE_TOKEN", "node-token")
	t.Setenv("LENKER_AGENT_PANEL_URL", "https://panel.example.com/")
	t.Setenv("LENKER_AGENT_STATE_DIR", "/tmp/lenker")
	t.Setenv("LENKER_AGENT_LOG_LEVEL", "debug")
	t.Setenv("LENKER_AGENT_HEARTBEAT_INTERVAL", "45s")
	t.Setenv("LENKER_AGENT_CONFIG_POLL_INTERVAL", "60s")
	t.Setenv("LENKER_AGENT_XRAY_BIN", "/usr/local/bin/xray")
	t.Setenv("LENKER_AGENT_TLS_ENABLED", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}
	if cfg.HTTPAddr != ":9999" || cfg.NodeID != "node-1" || cfg.BootstrapToken != "token" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if cfg.NodeToken != "node-token" {
		t.Fatalf("unexpected node token")
	}
	if cfg.PanelURL != "https://panel.example.com" {
		t.Fatalf("expected panel url trim, got %q", cfg.PanelURL)
	}
	if cfg.HeartbeatInterval != 45*time.Second {
		t.Fatalf("unexpected heartbeat interval: %s", cfg.HeartbeatInterval)
	}
	if cfg.ConfigPollInterval != 60*time.Second {
		t.Fatalf("unexpected config poll interval: %s", cfg.ConfigPollInterval)
	}
	if cfg.XrayBin != "/usr/local/bin/xray" {
		t.Fatalf("unexpected xray bin: %q", cfg.XrayBin)
	}
	if !cfg.TLSEnabled {
		t.Fatalf("expected tls enabled")
	}
}

func TestLoadInvalidHeartbeat(t *testing.T) {
	t.Setenv("LENKER_AGENT_HEARTBEAT_INTERVAL", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatalf("expected invalid heartbeat interval error")
	}
}

func TestLoadInvalidConfigPollInterval(t *testing.T) {
	t.Setenv("LENKER_AGENT_CONFIG_POLL_INTERVAL", "0s")

	if _, err := Load(); err == nil {
		t.Fatalf("expected invalid config poll interval error")
	}
}
