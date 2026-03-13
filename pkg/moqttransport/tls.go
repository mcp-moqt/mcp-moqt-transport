package mcpmoqt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

// generateSelfSignedCert generates a self-signed certificate for testing.
func generateSelfSignedCert() (*tls.Config, error) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	// Encode certificate and private key
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	// Load certificate
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	// Create TLS config
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"moq-00"},
	}, nil
}

// defaultTLSConfig returns the default TLS config for the given role.
func defaultTLSConfig(role transportRole, alpn []string) (*tls.Config, error) {
	switch role {
	case roleServer:
		// Server: use self-signed certificate
		cfg, err := generateSelfSignedCert()
		if err != nil {
			return nil, err
		}
		cfg.NextProtos = alpn
		return cfg, nil
	case roleClient:
		// Client: skip verification for development
		return &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         alpn,
		}, nil
	default:
		return nil, nil
	}
}
