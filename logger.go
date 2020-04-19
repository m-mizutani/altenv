package main

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

type loggerHook struct{}

func (x *loggerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (x *loggerHook) Fire(entry *logrus.Entry) error {
	entry.Message = fmt.Sprintf("[envctl] %s", entry.Message)
	return nil
}

func setupLogger() {
	logger.AddHook(&loggerHook{})
}

func setLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		return fmt.Errorf("Invalid log level '%s'", level)
	}

	return nil
}
