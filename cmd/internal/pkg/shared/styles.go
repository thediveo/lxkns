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

package shared

import (
	"fmt"

	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	COLOR_ALWAYS = "always"
	COLOR_AUTO   = "auto"
	COLOR_NEVER  = "never"
)

// Style-related CLI command flags.
var (
	ColorMode string // colorization mode: "always", "auto", or "never"
)

// AddStyleFlags adds global CLI command flags related to colorization and
// styling.
func AddStyleFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVarP(&ColorMode, "color", "c", "auto",
		"colorize the output; can be 'always' (default if omitted), 'auto',\n"+
			"or 'never'")
	rootCmd.PersistentFlags().Lookup("color").NoOptDefVal = "always"
}

// ConfigureStyles configures the various output rendering styles based on CLI
// command flags and configuration files. It needs to be called before
// rendering any (styled) output, ideally as a "PersistentPreRun" of a Cobra
// root command.
func ConfigureStyles() error {
	switch ColorMode {
	case COLOR_ALWAYS:
		colorProfile = termenv.ANSI256
	case COLOR_AUTO:
		colorProfile = termenv.ColorProfile()
	case COLOR_NEVER:
		colorProfile = termenv.Ascii
	default:
		return fmt.Errorf(
			"invalid --color mode %q, must be 'always', 'never', or 'auto'",
			ColorMode)
	}
	readStyles(defaultstyles)
	return nil
}

// The termenv color profile to be used when styling, such as plain colorless
// ASCII, 256 colors, et cetera.
var colorProfile termenv.Profile

// The set of styles for styling types of Linux-kernel namespaces differently,
// as well as some more elements, such as process names, user names, et
// cetera.
var (
	MntStyle    Style // styles mnt: namespaces
	CgroupStyle Style // styles cgroup: namespaces
	UTSStyle    Style // styles uts: namespaces
	IPCStyle    Style // styles ipc: namespaces
	UserStyle   Style // styles utc: namespaces
	PIDStyle    Style // styles pid: namespaces
	NetStyle    Style // styles net: namespaces

	OwnerStyle   Style // styles owner username and UID
	ProcessStyle Style // styles process names
	UnknownStyle Style // styles undetermined elements, such as unknown PIDs.
)

// Maps configuration top-level element names to their corresponding Style
// objects for storing and using specific style information.
var Styles = map[string]*Style{
	"mnt":    &MntStyle,
	"cgroup": &CgroupStyle,
	"uts":    &UTSStyle,
	"ipc":    &IPCStyle,
	"user":   &UserStyle,
	"pid":    &PIDStyle,
	"net":    &NetStyle,

	"owner":   &OwnerStyle,
	"process": &ProcessStyle,
	"unknown": &UnknownStyle,
}

// Style represents a set of style settings to apply to text when rendering to
// a terminal supporting ANSI coloring and styling.
type Style struct {
	style termenv.Style
}

// S returns the specified text s styled according to this Style's
// configuration. If multiple strings are specified, then the styling is
// applied to each string anew, thus allowing interleaving differently styled
// strings with this styling. The individual strings are put immediately
// adjacent to each without any intermediate spaces.
func (st *Style) S(s ...string) string {
	r := ""
	for _, str := range s {
		r += st.style.Styled(str)
	}
	return r
}

// Q returns the specified text s properly quoted and styled according to this
// Style's configuration.
func (st *Style) Q(s string) string {
	return st.style.Styled(fmt.Sprintf("%q", s))
}

func (st *Style) D(d int64) string {
	return ""
}

func readStyles(configyaml string) {
	y := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(configyaml), &y)
	if err != nil {
		panic(err) // FIXME: better error reporting
	}
	for key, settings := range y {
		sty, ok := Styles[key]
		if !ok {
			continue
		}
		for _, setting := range settings.([]interface{}) {
			if s, ok := setting.(string); ok {
				switch s {
				case "bold":
					sty.style = sty.style.Bold()
				}
			} else {
				if kvs, ok := setting.(map[interface{}]interface{}); ok {
					for k, v := range kvs {
						if s, ok := k.(string); ok {
							switch s {
							case "foreground":
								sty.style = sty.style.Foreground(colorProfile.Color(v.(string)))
							case "background":
								sty.style = sty.style.Background(colorProfile.Color(v.(string)))
							}
						}
					}
				}
			}
		}
	}
}

const defaultstyles = `
user:
- bold
- background: '#dadada' # Schwitters would be delighted!
- foreground: '#000000'
pid:
- bold
- background: '#e6ffff'
- foreground: '#000000'
cgroup:
- background: '#ffe6e6'
ipc:
- background: '#ffffcc'
mnt:
- background: '#e6e6ff'
net:
- background: '#ccffdd'
uts:
- background: '#d9b3ff'

process:
- foreground: '#004000'
owner:
- foreground: '#808000'
unknown:
- foreground: '#800000'
`
