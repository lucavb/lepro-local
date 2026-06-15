package config

import (
	"path/filepath"
	"testing"
)

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "..", "lepro-lab.toml.example"))
	if err != nil {
		t.Fatalf("load example: %v", err)
	}
	if cfg.Device.ID != "3619294555" {
		t.Fatalf("unexpected device id: %s", cfg.Device.ID)
	}
	if cfg.Device.FriendlyName != "patio-lights" {
		t.Fatalf("unexpected friendly name: %s", cfg.Device.FriendlyName)
	}
}

func TestLoadEnvOnly(t *testing.T) {
	t.Setenv("LEPRO_DEVICE_ID", "3619294555")
	t.Setenv("LEPRO_DEVICE_FRIENDLY_NAME", "patio-lights")
	t.Setenv("LEPRO_HOME_BROKER", "mqtt://127.0.0.1:1883")
	t.Setenv("LEPRO_LEPRO_BROKER", "mqtts://mqtt.example.home:8883")
	t.Setenv("LEPRO_CA_FILE", filepath.Join("testdata", "ca.pem"))
	t.Setenv("LEPRO_TLS_SERVER_NAME", "mqtt.example.home")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load env-only: %v", err)
	}
	if cfg.Device.ID != "3619294555" || cfg.Device.FriendlyName != "patio-lights" {
		t.Fatalf("unexpected env-only device config: %+v", cfg.Device)
	}
	if cfg.Bridge.HomeBroker != "mqtt://127.0.0.1:1883" {
		t.Fatalf("unexpected home broker: %s", cfg.Bridge.HomeBroker)
	}
}

func TestTLSConfigModes(t *testing.T) {
	cfg := BridgeConfig{
		CAFile:        filepath.Join("testdata", "ca.pem"),
		TLSServerName: "mqtt.example.home",
	}
	tlsCfg, err := cfg.TLSConfig()
	if err != nil {
		t.Fatalf("tls config: %v", err)
	}
	if tlsCfg.ServerName != "mqtt.example.home" {
		t.Fatalf("server name not applied")
	}
	if tlsCfg.RootCAs == nil {
		t.Fatalf("expected root CAs")
	}

	insecureCfg, err := (BridgeConfig{TLSInsecure: true}).TLSConfig()
	if err != nil {
		t.Fatalf("insecure tls config: %v", err)
	}
	if !insecureCfg.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify")
	}
}
