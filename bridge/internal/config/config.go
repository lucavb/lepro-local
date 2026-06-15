package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Device DeviceConfig `toml:"device"`
	Bridge BridgeConfig `toml:"bridge"`
}

type DeviceConfig struct {
	ID           string `toml:"id"`
	FriendlyName string `toml:"friendly_name"`
}

type BridgeConfig struct {
	HomeBroker    string `toml:"home_broker"`
	LeproBroker   string `toml:"lepro_broker"`
	CAFile        string `toml:"ca_file"`
	TLSServerName string `toml:"tls_server_name"`
	TLSInsecure   bool   `toml:"tls_insecure"`
	CertFile      string `toml:"cert_file"`
	KeyFile       string `toml:"key_file"`
}

func Load(path string) (Config, error) {
	var cfg Config
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			if !os.IsNotExist(err) {
				return Config{}, err
			}
		} else if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return Config{}, err
		}
	}
	cfg.applyEnv()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func LoadEnv() (Config, error) {
	var cfg Config
	cfg.applyEnv()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) applyEnv() {
	c.Device.ID = firstNonEmpty(os.Getenv("LEPRO_DEVICE_ID"), c.Device.ID)
	c.Device.FriendlyName = firstNonEmpty(os.Getenv("LEPRO_DEVICE_FRIENDLY_NAME"), c.Device.FriendlyName)

	c.Bridge.HomeBroker = firstNonEmpty(os.Getenv("LEPRO_HOME_BROKER"), os.Getenv("LEPRO_BRIDGE_HOME_BROKER"), c.Bridge.HomeBroker)
	c.Bridge.LeproBroker = firstNonEmpty(os.Getenv("LEPRO_LEPRO_BROKER"), os.Getenv("LEPRO_BRIDGE_LEPRO_BROKER"), c.Bridge.LeproBroker)
	c.Bridge.CAFile = firstNonEmpty(os.Getenv("LEPRO_CA_FILE"), os.Getenv("LEPRO_BRIDGE_CA_FILE"), c.Bridge.CAFile)
	c.Bridge.TLSServerName = firstNonEmpty(os.Getenv("LEPRO_TLS_SERVER_NAME"), os.Getenv("LEPRO_BRIDGE_TLS_SERVER_NAME"), c.Bridge.TLSServerName)
	c.Bridge.CertFile = firstNonEmpty(os.Getenv("LEPRO_TLS_CERT_FILE"), os.Getenv("LEPRO_BRIDGE_TLS_CERT_FILE"), c.Bridge.CertFile)
	c.Bridge.KeyFile = firstNonEmpty(os.Getenv("LEPRO_TLS_KEY_FILE"), os.Getenv("LEPRO_BRIDGE_TLS_KEY_FILE"), c.Bridge.KeyFile)
	if v, ok := lookupBoolEnv("LEPRO_TLS_INSECURE", "LEPRO_BRIDGE_TLS_INSECURE"); ok {
		c.Bridge.TLSInsecure = v
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func lookupBoolEnv(keys ...string) (bool, bool) {
	for _, key := range keys {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}
		switch strings.ToLower(raw) {
		case "1", "true", "yes", "on":
			return true, true
		case "0", "false", "no", "off":
			return false, true
		}
	}
	return false, false
}

func (c Config) Validate() error {
	if c.Device.ID == "" {
		return fmt.Errorf("device.id is required")
	}
	if c.Device.FriendlyName == "" {
		return fmt.Errorf("device.friendly_name is required")
	}
	if c.Bridge.HomeBroker == "" {
		return fmt.Errorf("bridge.home_broker is required")
	}
	if c.Bridge.LeproBroker == "" {
		return fmt.Errorf("bridge.lepro_broker is required")
	}
	if _, err := parseBrokerURL(c.Bridge.HomeBroker); err != nil {
		return fmt.Errorf("bridge.home_broker: %w", err)
	}
	if _, err := parseBrokerURL(c.Bridge.LeproBroker); err != nil {
		return fmt.Errorf("bridge.lepro_broker: %w", err)
	}
	return nil
}

func parseBrokerURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "mqtt", "mqtts", "tcp", "tls":
		return u, nil
	default:
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
}
