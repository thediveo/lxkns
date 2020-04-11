// Implements controlling the colorization mode via the "--colormode" CLI
// flag, which can be either "always" (="on"), "auto", or "never" (="off").

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
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
)

// The termenv color profile to be used when styling, such as plain colorless
// ASCII, 256 colors, et cetera.
var colorProfile termenv.Profile

// The CLI flag colorize controls output colorization.
var colorize ColorMode

// ColorMode is an enumeration for colorizing output always, auto(matic), and
// never.
type ColorMode enumflag.Flag

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
		CmNever:  {"never", "none", "off"},
	}, enumflag.EnumCaseSensitive
}

func init() {
	// Delayed registration our CLI flag.
	pflagCreators.Register(func(rootCmd *cobra.Command) {
		rootCmd.PersistentFlags().VarP(&colorize, "color", "c",
			"colorize the output; can be 'always' (default if omitted), 'auto',\n"+
				"or 'never'")
		rootCmd.PersistentFlags().Lookup("color").NoOptDefVal = "always"
	})
	// Delayed color profile selection based on our CLI flag and terminal
	// profile detection, just before the selected command runs.
	runhooks.Register(func() {
		// Colorization mode...
		switch colorize {
		case CmAlways:
			colorProfile = termenv.ANSI256
		case CmAuto:
			colorProfile = termenv.ColorProfile()
		case CmNever:
			colorProfile = termenv.Ascii
		}
	})
}
