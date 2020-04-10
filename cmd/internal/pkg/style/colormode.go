// Implements a --colormode pflag value enumeration type, which can only be
// either "always", "auto", or "never".

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

package style

import (
	"errors"
	"strings"
)

// ColorMode is an enumeration for colorizing output always, auto(matic), and
// never.
type ColorMode int

// Enumeration of allowed ColorMode values.
const (
	CmAlways ColorMode = iota // always colorize
	CmAuto                    // colorize if output goes to a terminal
	CmNever                   // never colorize
)

// colorModusName maps color mode enum values to their textual
// representations.
var colorModeNames = map[ColorMode]string{
	CmAlways: "always",
	CmAuto:   "auto",
	CmNever:  "never",
}

// String returns the text representation of a color mode value.
func (cm *ColorMode) String() string {
	return colorModeNames[*cm]
}

// Set parses the given color mode string and converts it into the
// corresponding (enumeration) value. We're actually more liberal than what
// "ls" accepts.
func (cm *ColorMode) Set(s string) error {
	switch strings.ToLower(s) {
	case "always", "on":
		*cm = CmAlways
	case "auto":
		*cm = CmAuto
	case "never", "off":
		*cm = CmNever
	default:
		return errors.New("must be 'always'/'on', 'never'/'off', or 'auto'")
	}
	return nil
}

// Type returns the pflag name for color mode values.
func (cm *ColorMode) Type() string {
	return "colormode"
}
