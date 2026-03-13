package logger

import (
	"io"
	"log"
	"os"
	"sync"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Logger struct {
	logger *log.Logger
	level  Level
	mu     sync.Mutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

func init() {
	once.Do(func() {
		defaultLogger = New(os.Stdout, InfoLevel)
	})
}

func New(output io.Writer, level Level) *Logger {
	return &Logger{
		logger: log.New(output, "", log.LstdFlags),
		level:  level,
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	prefix := ""
	switch level {
	case DebugLevel:
		prefix = "[DEBUG]"
	case InfoLevel:
		prefix = "[INFO]"
	case WarnLevel:
		prefix = "[WARN]"
	case ErrorLevel:
		prefix = "[ERROR]"
	}

	l.logger.Printf(prefix+" "+format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

func GetLogger() *Logger {
	return defaultLogger
}
