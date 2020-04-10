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
	"github.com/thediveo/enumflag"
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

// Implements the methods required by spf13/cobra in order to use the enum as
// a flag.
func (cm *ColorMode) String() string     { return enumflag.String(cm) }
func (cm *ColorMode) Set(s string) error { return enumflag.Set(cm, s) }
func (cm *ColorMode) Type() string       { return "colormode" }

// Implements the method required by enumflag to map enum values to their
// textual identifiers.
func (cm *ColorMode) Enums() (interface{}, enumflag.EnumCaseSensitivity) {
	return map[ColorMode][]string{
		CmAlways: {"always", "on"},
		CmAuto:   {"auto"},
		CmNever:  {"never", "off"},
	}, enumflag.EnumCaseSensitive
}
