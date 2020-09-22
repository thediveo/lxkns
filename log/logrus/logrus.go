package logrus

import (
	"github.com/sirupsen/logrus"
	logger "github.com/thediveo/lxkns/log"
)

func logToLogrus(level logger.Level, msg string) {
	switch level {
	case logger.PanicLevel:
		logrus.Panic(msg)
	case logger.FatalLevel:
		logrus.Fatal(msg)
	case logger.ErrorLevel:
		logrus.Error(msg)
	case logger.WarnLevel:
		logrus.Warn(msg)
	case logger.InfoLevel:
		logrus.Info(msg)
	case logger.DebugLevel:
		logrus.Debug(msg)
	case logger.TraceLevel:
		logrus.Trace(msg)
	}
}

func init() {
	logger.Loggerfn = logToLogrus
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
}
