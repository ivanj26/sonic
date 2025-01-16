package logger

import (
	"errors"
	"fmt"
	"io"
	"ivanj26/sonic/constant"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type loggerType int

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

const (
	TYPE_INFO loggerType = iota
	TYPE_ERROR
)

// Initialize sets up the logger
func Initialize(logFilePath string, enableInfoLog bool) error {
	fileName, dir := filepath.Base(logFilePath), filepath.Dir(logFilePath)
	if len(fileName) == 0 {
		return errors.New("Invalid filename, unable to initialize log!")
	}
	fileName = strings.Split(fileName, ".")[0]

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	infoFileName := dir + "/" + fmt.Sprintf(constant.INFO_LOG_FILENAME_FMT, fileName)
	err := initLogger(TYPE_INFO, infoFileName, enableInfoLog)
	if err != nil {
		return err
	}

	errFileName := dir + "/" + fmt.Sprintf(constant.ERROR_LOG_FILENAME_FMT, fileName)
	err = initLogger(TYPE_ERROR, errFileName, true)
	if err != nil {
		return err
	}

	return nil
}

func initLogger(lType loggerType, fileName string, enableLog bool) error {
	if enableLog {
		// Open or create the info log file
		logFile, err := os.OpenFile(
			fileName,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			return err
		}
		multiWriter := io.MultiWriter(os.Stdout, logFile)

		if lType == TYPE_INFO {
			infoLogger = log.New(logFile, "[INFO] ", log.LstdFlags|log.Lshortfile)
			infoLogger.SetOutput(multiWriter)
		} else {
			errorLogger = log.New(logFile, "[ERROR] ", log.LstdFlags|log.Lshortfile)
			errorLogger.SetOutput(multiWriter)
		}
	}

	return nil
}

// Infof logs an informational message
func Infof(format string, v ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf(format, v...)
	}
}

// Debugf logs a debug message
func Debugf(format string, v ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf("[DEBUG] "+format, v...)
	}
}

// Errorf logs an error message
func Errorf(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, v...)
	}
}

// Error logs an error message without format
func Error(msg string) {
	if errorLogger != nil {
		errorLogger.Fatal(msg)
	}
}

// Fatalf logs a fatal error message and exits the application
func Fatalf(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Fatalf("[FATAL] "+format, v...)
	}
}

// Fatal logs a fatal error message and exits the application
func Fatal(msg string) {
	if errorLogger != nil {
		errorLogger.Fatal("[FATAL] " + msg)
	}
}
