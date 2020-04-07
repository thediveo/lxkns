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

// Theme is an enumeration for selecting either a light or dark theme.
type Theme int

// Enumeration of allowed Theme values.
const (
	ThDark  Theme = iota // default dark (background) theme
	ThLight              // light (background) theme
)

// themeNames maps them enum values to their textual representations.
var themeNames = map[Theme]string{
	ThDark:  "dark",
	ThLight: "light",
}

// String returns the text representation of a theme value.
func (t *Theme) String() string {
	return themeNames[*t]
}

// Set parses the given theme name string and converts it into the
// corresponding (enumeration) value.
func (t *Theme) Set(s string) error {
	switch strings.ToLower(s) {
	case "dark":
		*t = ThDark
	case "light":
		*t = ThLight
	default:
		return errors.New("must be 'dark' or 'light'")
	}
	return nil
}

// Type returns the pflag name for color mode values.
func (t *Theme) Type() string {
	return "theme"
}
