package logger

import "log"

type (
	loggerType int

	Logger struct {
		enableInfoLog  bool
		enableErrorLog bool
		infoLogger     *log.Logger
		errorLogger    *log.Logger
		logPath        string
	}

	ILogger interface {
		SetInfoEnabled(bool) ILogger
		SetFilePath(string) ILogger
		Initialize() error
	}
)

const (
	TYPE_INFO loggerType = iota
	TYPE_ERROR
)
