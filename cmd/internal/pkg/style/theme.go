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
	"fmt"
	"os"
	"path/filepath"

	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"github.com/thediveo/go-plugger/v2"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
)

// The CLI flag controlling which theme to use when colorization is used.
var theme Theme // dark or light color theme

// The CLI flag instructing us to dump a default theme, either the dark or
// light one, as specified in the theme variable.
var dumptheme bool // print the selected color theme to stdout

// Theme is an enumeration for selecting either a light or dark theme.
type Theme enumflag.Flag

// Enumeration of allowed Theme values.
const (
	ThemeDark  Theme = iota // default dark (background) theme
	ThemeLight              // light (background) theme
)

// Defines the textual representations for the Theme values.
var themeIds = map[Theme][]string{
	ThemeDark:  {"dark"},
	ThemeLight: {"light"},
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Register(
		plugger.WithName("theme"),
		plugger.WithGroup(cliplugin.Group),
		plugger.WithNamedSymbol("SetupCLI", ThemeSetupCLI),
		plugger.WithNamedSymbol("BeforeRun", ThemeBeforeRun))
}

// ThemeSetupCLI is a plugin function that registers the CLI flags related to
// theming.
func ThemeSetupCLI(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().Var(
		enumflag.New(&theme, "theme", themeIds, enumflag.EnumCaseSensitive),
		"theme", "colorization theme 'dark' or 'light'")
	rootCmd.PersistentFlags().BoolVar(&dumptheme, "dump", false,
		"dump colorization theme to stdout (for saving to ~/.lxknsrc.yaml)")
}

// ThemeBeforeRun is a plugin function that handles selection, reading, or
// dumping of styling profiles, just before the selected command runs. In case
// of dumping, it also exits this process, so the itself command won't ever
// start.
func ThemeBeforeRun() error {
	// If the user wants to dump a theme using "--dump" then the selected
	// default theme, light or dark, takes precedence and any user
	// definitions get ignored in this special case. This allows users to
	// recreate a clean user-defined theme.
	if dumptheme {
		fmt.Fprint(os.Stdout, defaultThemes[theme])
		os.Exit(0)
	}
	// If there is a user-defined theme in the user's home directory, then
	// this takes precedence over any --theme selection. Unless the file
	// is empty, then we fall back onto the default themes.
	var th string
	if home, err := os.UserHomeDir(); err == nil {
		// #nosec G304
		if styling, err := os.ReadFile(filepath.Join(home, ".lxknsrc.yaml")); err == nil {
			th = string(styling)
		}
	}
	if th == "" {
		th = defaultThemes[theme]
	}
	// If the colorProfile is set to Ascii, then we actually skip all
	// styling, not just coloring, such as "ls" does.
	if colorProfile != termenv.Ascii {
		parseStyles(th)
	}
	return nil
}

// Maps the Theme enumeration to the corresponding theme descriptions.
var defaultThemes = map[Theme]string{
	ThemeDark:  defaultDarkTheme,
	ThemeLight: defaultLightTheme,
}
