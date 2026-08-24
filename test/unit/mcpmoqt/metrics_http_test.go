package mcpmoqt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mcpmoqt "github.com/mcp-moqt/mcp-moqt-transport/pkg/moqttransport"
	"github.com/stretchr/testify/require"
)

func TestMetricsHTTPServer_HandleMetrics(t *testing.T) {
	m := mcpmoqt.NewMetrics()
	m.RecordMessageSent(100)
	m.RecordMessageReceived(50)
	m.RecordDataTrackMessage(32)

	h := mcpmoqt.MetricsHandler(m)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	text := string(body)
	require.Contains(t, text, "mcp_moqt_messages_sent_total")
	require.Contains(t, text, "mcp_moqt_data_track_messages_total")
	require.Contains(t, text, "mcp_moqt_throughput_out_bytes_per_sec")
}

func TestMetricsHandler_DefaultMetrics(t *testing.T) {
	h := mcpmoqt.MetricsHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, strings.Contains(w.Body.String(), "mcp_moqt_uptime_seconds"))
}

func TestMetrics_SnapshotAndDataTrack(t *testing.T) {
	m := mcpmoqt.NewMetrics()
	m.RecordDataTrackMessage(128)
	snap := m.Snapshot()
	require.Equal(t, int64(1), snap.DataTrackMessages)
	require.Equal(t, int64(128), snap.DataTrackBytes)
}

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

func TestNewMOQTServerTransport_WithTLSCertFiles(t *testing.T) {
	dir := t.TempDir()
	cert, key := writeTestCertPair(t, dir)
	srv, err := mcpmoqt.NewMOQTServerTransport(
		mcpmoqt.WithAddr("127.0.0.1:0"),
		mcpmoqt.WithTLSCertFiles(cert, key),
	)
	require.NoError(t, err)
	require.NotNil(t, srv)
	require.NoError(t, srv.Listen())
	require.NoError(t, srv.Close())
}

func TestNewMOQTServerTransport_ListenDefaultTLS(t *testing.T) {
	srv, err := mcpmoqt.NewMOQTServerTransport(mcpmoqt.WithAddr("127.0.0.1:0"))
	require.NoError(t, err)
	defer srv.Close()
	require.NoError(t, srv.Listen())
}
