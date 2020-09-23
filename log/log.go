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

package log

import "fmt"

// Level defines the representation of a single logging level for the purposes
// of logging in the lxkns module.
type Level uint32

// The available logging levels; please note that lxkns only uses a subset of
// them. By coincidence, these levels match the logging levels of logus ;)
const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

// Logger defines the logging interface that lxkns expects of adapters to
// external logging modules.
type Logger interface {
	Log(level Level, msg string)
	SetLevel(level Level)
}

// SetLogger sets the logging adapter to be used. Please note that this is not a
// multiple Go-routine safe operation, so it should be carried out before
// spawning any logging Go routines.
func SetLogger(logger Logger) {
	loggeradapter = logger
}

// loggeradapter points to a logging adapter, if any, for interfacing with an
// lxkns-external logging module.
var loggeradapter Logger

// SetLevel sets the logging level for lxkns-internal module logging. Depending
// on the external logger adapter, the level setting might also be propagated to
// this external logger.
func SetLevel(level Level) {
	loglevel = level
	if loggeradapter != nil {
		loggeradapter.SetLevel(level)
	}
}

// The current logging level for lxkns-internal module logging.
var loglevel = InfoLevel

// Errorf logs a formatted error.
func Errorf(format string, args ...interface{}) {
	forward(ErrorLevel, format, args...)
}

// Warnf logs a formatted warning.
func Warnf(format string, args ...interface{}) {
	forward(WarnLevel, format, args...)
}

// Infof logs formatted information.
func Infof(format string, args ...interface{}) {
	forward(InfoLevel, format, args...)
}

// Infofn logs information from the specified fn producer function.
func Infofn(fn func() string) {
	callforward(InfoLevel, fn)
}

// Debugf logs formatted debug information.
func Debugf(format string, args ...interface{}) {
	forward(DebugLevel, format, args...)
}

// Debugfn logs debugging information from the specified fn producer function.
func Debugfn(fn func() string) {
	callforward(DebugLevel, fn)
}

// forward sends a logging message to an external logging module (if set), if
// the set logging level is reached or surpassed. Otherwise, it silently drops
// the message.
func forward(level Level, format string, args ...interface{}) {
	if level <= loglevel && loggeradapter != nil {
		loggeradapter.Log(level, fmt.Sprintf(format, args...))
	}
}

// callforward collects a logging message from a producer function, but only if
// an external logging module has been set and the necessary logging level
// reached or surpassed.
func callforward(level Level, fn func() string) {
	if level <= loglevel && loggeradapter != nil && fn != nil {
		loggeradapter.Log(level, fn())
	}
}
