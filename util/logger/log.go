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

var logInstance *Logger

func New() ILogger {
	if logInstance != nil {
		return logInstance
	}

	logInstance = &Logger{
		enableInfoLog:  true,
		enableErrorLog: true,
		infoLogger:     nil,
		errorLogger:    nil,
		logPath:        "",
	}
	return logInstance
}

// Enable/disable info log
func (l *Logger) SetInfoEnabled(isEnable bool) ILogger {
	if isEnable {
		l.enableInfoLog = true
	} else {
		l.enableInfoLog = false
	}
	return l
}

// Set the log file path
func (l *Logger) SetFilePath(filePath string) ILogger {
	l.logPath = filePath
	return l
}

// Initialize sets up the logger
func (l *Logger) Initialize() error {
	fileName, dir := filepath.Base(l.logPath), filepath.Dir(l.logPath)
	if len(fileName) == 0 {
		return errors.New("Invalid filename, unable to initialize log!")
	}
	fileName = strings.Split(fileName, ".")[0]

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	infoFileName := dir + "/" + fmt.Sprintf(constant.INFO_LOG_FILENAME_FMT, fileName)
	err := initLogger(TYPE_INFO, infoFileName, l.enableInfoLog)
	if err != nil {
		return err
	}

	errFileName := dir + "/" + fmt.Sprintf(constant.ERROR_LOG_FILENAME_FMT, fileName)
	err = initLogger(TYPE_ERROR, errFileName, l.enableErrorLog)
	if err != nil {
		return err
	}

	return nil
}

// Infof logs an informational message
func Infof(format string, v ...interface{}) {
	if logInstance != nil && logInstance.infoLogger != nil {
		logInstance.infoLogger.Printf(format, v...)
	}
}

// Debugf logs a debug message
func Debugf(format string, v ...interface{}) {
	if logInstance != nil && logInstance.infoLogger != nil {
		logInstance.infoLogger.Printf("[DEBUG] "+format, v...)
	}
}

// Errorf logs an error message
func Errorf(format string, v ...interface{}) {
	if logInstance != nil && logInstance.errorLogger != nil {
		logInstance.errorLogger.Printf(format, v...)
	}
}

// Error logs an error message without format
func Error(msg string) {
	if logInstance != nil && logInstance.errorLogger != nil {
		logInstance.errorLogger.Fatal(msg)
	}
}

// Fatalf logs a fatal error message and exits the application
func Fatalf(format string, v ...interface{}) {
	if logInstance != nil && logInstance.errorLogger != nil {
		logInstance.errorLogger.Fatalf("[FATAL] "+format, v...)
	}
}

// Fatal logs a fatal error message and exits the application
func Fatal(msg string) {
	if logInstance != nil && logInstance.errorLogger != nil {
		logInstance.errorLogger.Fatal("[FATAL] " + msg)
	}
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
			logInstance.infoLogger = log.New(logFile, "[INFO] ", log.LstdFlags|log.Lshortfile)
			logInstance.infoLogger.SetOutput(multiWriter)
		} else {
			logInstance.errorLogger = log.New(logFile, "[ERROR] ", log.LstdFlags|log.Lshortfile)
			logInstance.errorLogger.SetOutput(multiWriter)
		}
	}

	return nil
}
