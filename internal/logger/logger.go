package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	
	// Set log level
	Logger.SetLevel(logrus.InfoLevel)
	
	// Set formatter
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	// Set output to both console and file
	setupFileLogging()
}

func setupFileLogging() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Logger.Warn("Home directory bulunamadı, sadece console logging kullanılacak")
		return
	}
	
	logDir := filepath.Join(homeDir, ".spotomusic", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		Logger.Warn("Log directory oluşturulamadı: %v", err)
		return
	}
	
	logFile := filepath.Join(logDir, "spotomusic.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.Warn("Log file açılamadı: %v", err)
		return
	}
	
	// Write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	Logger.SetOutput(multiWriter)
}

// SetLevel sets the log level
func SetLevel(level logrus.Level) {
	Logger.SetLevel(level)
}

// SetVerbose enables verbose logging
func SetVerbose(verbose bool) {
	if verbose {
		Logger.SetLevel(logrus.DebugLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}

// Info logs an info message
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}
