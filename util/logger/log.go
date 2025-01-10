package logger

import (
	"io"
	"log"
	"os"
)

var logger *log.Logger

// Initialize sets up the logger
func Initialize(logFilePath string) error {
	// Open or create the log file
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	// Create a logger instance
	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)

	// Create a MultiWriter for both file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(multiWriter)
	return nil
}

// Infof logs an informational message
func Infof(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, v...)
	}
}

// Errorf logs an error message
func Errorf(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, v...)
	}
}

// Error logs an error message without format
func Error(msg string) {
	if logger != nil {
		logger.Fatal("[ERROR] " + msg)
	}
}

// Fatalf logs a fatal error message and exits the application
func Fatalf(format string, v ...interface{}) {
	if logger != nil {
		logger.Fatalf("[FATAL] "+format, v...)
	}
}

// Fatal logs a fatal error message and exits the application
func Fatal(msg string) {
	if logger != nil {
		logger.Fatal("[FATAL] " + msg)
	}
}
