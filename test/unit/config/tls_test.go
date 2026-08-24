package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	mcpconfig "github.com/mcp-moqt/mcp-moqt-transport/pkg/config"
	"github.com/stretchr/testify/require"
)

func writeTestCertPair(t *testing.T, dir string) (certFile, keyFile string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certFile = filepath.Join(dir, "server.crt")
	keyFile = filepath.Join(dir, "server.key")
	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), 0600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}), 0600))
	return certFile, keyFile
}

func TestTLSConfig_LoadServerTLS(t *testing.T) {
	dir := t.TempDir()
	cert, key := writeTestCertPair(t, dir)
	cfg := mcpconfig.TLSConfig{CertFile: cert, KeyFile: key}
	tlsCfg, err := cfg.LoadServerTLS([]string{"moq-00"})
	require.NoError(t, err)
	require.Len(t, tlsCfg.Certificates, 1)
	require.Equal(t, []string{"moq-00"}, tlsCfg.NextProtos)
}

func TestValidateCertPair(t *testing.T) {
	dir := t.TempDir()
	cert, key := writeTestCertPair(t, dir)
	require.NoError(t, mcpconfig.ValidateCertPair(cert, key))
}

func TestConfig_Validate_TLSPaths(t *testing.T) {
	dir := t.TempDir()
	cert, key := writeTestCertPair(t, dir)
	cfg := &mcpconfig.Config{
		Addr: "127.0.0.1:8080",
		ALPN: []string{"moq-00"},
		TLS:  mcpconfig.TLSConfig{CertFile: cert, KeyFile: key},
	}
	require.NoError(t, cfg.Validate())
}
