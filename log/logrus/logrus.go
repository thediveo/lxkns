// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package logrus

import (
	"github.com/sirupsen/logrus"
	"github.com/thediveo/lxkns/log"
)

// logrusser implements the Logger interface xlkns expects.
type logrusser struct{}

// adapter can be left zero, as it just serves to satisfy lxkns' external logger
// interface.
var adapter logrusser

// SetLevel propagates the lxkns module-specific logging level to the logrus
// standard logger.
func (*logrusser) SetLevel(level log.Level) {
	logrus.SetLevel(logrus.Level(level))
}

// Log logs a single message of a particular logging level to the logrus
// standard logger, keeping the logging level intact.
func (*logrusser) Log(level log.Level, msg string) {
	switch level {
	case log.PanicLevel:
		logrus.Panic(msg)
	case log.FatalLevel:
		logrus.Fatal(msg)
	case log.ErrorLevel:
		logrus.Error(msg)
	case log.WarnLevel:
		logrus.Warn(msg)
	case log.InfoLevel:
		logrus.Info(msg)
	case log.DebugLevel:
		logrus.Debug(msg)
	case log.TraceLevel:
		logrus.Trace(msg)
	}
}

// Wires up lxkns module-internal logging with the standard logrus logger and
// sets the logging format to use full timestamps and coloring.
func init() {
	log.SetLogger(&adapter)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}
