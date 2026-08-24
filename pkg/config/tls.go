package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"
)

// TLSConfig holds TLS certificate paths and client verification options.
type TLSConfig struct {
	CertFile           string `json:"cert_file" yaml:"cert_file"`
	KeyFile            string `json:"key_file" yaml:"key_file"`
	CAFile             string `json:"ca_file" yaml:"ca_file"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// HasServerCert returns true when both cert and key files are configured.
func (t TLSConfig) HasServerCert() bool {
	return t.CertFile != "" && t.KeyFile != ""
}

// LoadServerTLS builds a server-side tls.Config from certificate files.
func (t TLSConfig) LoadServerTLS(alpn []string) (*tls.Config, error) {
	if !t.HasServerCert() {
		return nil, fmt.Errorf("tls cert_file and key_file are required for server TLS")
	}
	cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load tls key pair: %w", err)
	}
	if len(alpn) == 0 {
		alpn = []string{"moq-00"}
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   append([]string(nil), alpn...),
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// LoadClientTLS builds a client-side tls.Config with optional CA pool.
func (t TLSConfig) LoadClientTLS(alpn []string) (*tls.Config, error) {
	if len(alpn) == 0 {
		alpn = []string{"moq-00"}
	}
	cfg := &tls.Config{
		InsecureSkipVerify: t.InsecureSkipVerify,
		NextProtos:         append([]string(nil), alpn...),
		MinVersion:         tls.VersionTLS12,
	}
	if t.CAFile != "" {
		pem, err := os.ReadFile(t.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read ca file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("parse ca file: no certificates found")
		}
		cfg.RootCAs = pool
		cfg.InsecureSkipVerify = false
	}
	return cfg, nil
}

// ValidateCertPair loads the certificate pair and checks expiry (warn-only helper for doctor).
func ValidateCertPair(certFile, keyFile string) error {
	if certFile == "" || keyFile == "" {
		return fmt.Errorf("cert_file and key_file must both be set")
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("certificate chain is empty")
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("parse leaf certificate: %w", err)
	}
	now := time.Now()
	if now.Before(leaf.NotBefore) {
		return fmt.Errorf("certificate not yet valid (starts %s)", leaf.NotBefore.Format(time.RFC3339))
	}
	if now.After(leaf.NotAfter) {
		return fmt.Errorf("certificate expired at %s", leaf.NotAfter.Format(time.RFC3339))
	}
	return nil
}
