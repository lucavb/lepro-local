package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func (b BridgeConfig) TLSConfig() (*tls.Config, error) {
	if b.TLSInsecure {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}

	cfg := &tls.Config{}
	if b.TLSServerName != "" {
		cfg.ServerName = b.TLSServerName
	}
	if b.CAFile != "" {
		pem, err := os.ReadFile(b.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read ca file: %w", err)
		}
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if ok := pool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("append ca file: no certificates found")
		}
		cfg.RootCAs = pool
	}
	if b.CertFile != "" || b.KeyFile != "" {
		if b.CertFile == "" || b.KeyFile == "" {
			return nil, fmt.Errorf("both cert_file and key_file are required together")
		}
		cert, err := tls.LoadX509KeyPair(b.CertFile, b.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load client cert: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	}
	return cfg, nil
}
