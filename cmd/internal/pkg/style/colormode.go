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
	"github.com/thediveo/enumflag/v2"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
)

// The termenv color profile to be used when styling, such as plain colorless
// ASCII, 256 colors, et cetera.
var colorProfile termenv.Profile

// The CLI flag colorize controls output colorization.
var colorize = ColorAuto

// ColorMode is an enumeration for colorizing output always, auto(matic), and
// never.
type ColorMode enumflag.Flag

// Enumeration of allowed ColorMode values.
const (
	ColorAlways ColorMode = iota // always colorize
	ColorAuto                    // colorize if output goes to a terminal
	ColorNever                   // never colorize
)

// Defines the textual representations for the ColorMode values.
var colorModeIds = map[ColorMode][]string{
	ColorAlways: {"always", "on"},
	ColorAuto:   {"auto"},
	ColorNever:  {"never", "none", "off"},
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		ColorModeSetupCLI, plugger.WithPlugin("colormode"))
	plugger.Group[cliplugin.BeforeCommand]().Register(
		ColorModeBeforeCommand, plugger.WithPlugin("colormode"))
}

// ColorModeSetupCLI is a plugin function that registers the CLI "color" flag.
func ColorModeSetupCLI(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().VarP(
		enumflag.New(&colorize, "color", colorModeIds, enumflag.EnumCaseSensitive),
		"color", "c",
		"colorize the output; can be 'always' (default if omitted), 'auto',\n"+
			"or 'never'")
	rootCmd.PersistentFlags().Lookup("color").NoOptDefVal = "always"
}

// ColorModeBeforeCommand is a plugin function that delays color profile
// selection based on our CLI flag and terminal profile detection until the last
// minute, just before the selected command runs.
func ColorModeBeforeCommand() error {
	// Colorization mode...
	switch colorize {
	case ColorAlways:
		colorProfile = termenv.ANSI256
	case ColorAuto:
		colorProfile = termenv.ColorProfile()
	case ColorNever:
		colorProfile = termenv.Ascii
	}
	return nil
}
