package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

// const
const (
	PanicLevel = "panic"
	FatalLevel = "fatal"
	ErrorLevel = "error"
	WarnLevel  = "warn"
	InfoLevel  = "info"
	DebugLevel = "debug"
)

type emptyWriter struct{}

func (ew emptyWriter) Write(p []byte) (int, error) {
	return 0, nil
}

var clog *logrus.Logger

// Log return console logger
func Log() *logrus.Logger {
	if clog == nil {
		Init("/tmp", "log", "info", 0)
	}
	return clog
}

func convertLevel(level string) logrus.Level {
	switch level {
	case PanicLevel:
		return logrus.PanicLevel
	case FatalLevel:
		return logrus.FatalLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case WarnLevel:
		return logrus.WarnLevel
	case InfoLevel:
		return logrus.InfoLevel
	case DebugLevel:
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}

// Init loggers
func Init(path string, appName string, level string, age uint64) {
	fileHooker := NewFileRotateHooker(path, appName, age)

	clog = logrus.New()
	LoadFunctionHooker(clog)
	clog.Hooks.Add(fileHooker)
	clog.Out = os.Stdout
	clog.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	clog.Level = convertLevel(level)
}
