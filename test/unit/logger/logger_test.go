package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mcp-moqt/mcp-moqt-transport/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf, logger.InfoLevel)
	
	l.Info("test message %s", "hello")
	
	output := buf.String()
	assert.True(t, strings.Contains(output, "[INFO]"))
	assert.True(t, strings.Contains(output, "test message hello"))
}

func TestLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf, logger.InfoLevel)
	
	l.Debug("debug message")
	
	output := buf.String()
	assert.Equal(t, "", output)
}

func TestLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf, logger.InfoLevel)
	
	l.Warn("warning message")
	
	output := buf.String()
	assert.True(t, strings.Contains(output, "[WARN]"))
	assert.True(t, strings.Contains(output, "warning message"))
}

func TestLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf, logger.InfoLevel)
	
	l.Error("error message")
	
	output := buf.String()
	assert.True(t, strings.Contains(output, "[ERROR]"))
	assert.True(t, strings.Contains(output, "error message"))
}

func TestLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf, logger.ErrorLevel)
	
	l.SetLevel(logger.DebugLevel)
	l.Debug("debug after level change")
	
	output := buf.String()
	assert.True(t, strings.Contains(output, "[DEBUG]"))
}
