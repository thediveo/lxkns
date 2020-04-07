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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Style/colorization-related CLI command flags.
var (
	colorize  ColorMode // colorization mode: "always", "auto", or "never"
	theme     Theme     // dark or light color theme
	dumptheme bool      // print the selected color theme to stdout
)

// AddStyleFlags adds global CLI command flags related to colorization and
// styling.
func AddStyleFlags(rootCmd *cobra.Command) {
	pf := rootCmd.PersistentFlags()
	pf.VarP(&colorize, "color", "c",
		"colorize the output; can be 'always' (default if omitted), 'auto',\n"+
			"or 'never'")
	pf.Lookup("color").NoOptDefVal = "always"
	pf.Var(&theme, "theme", "colorization theme 'dark' or 'light'")
	pf.BoolVar(&dumptheme, "dump", false,
		"dump colorization theme to stdout (for saving to ~/.lxknsrc.yaml)")
}

// HandleStyles configures the various output rendering styles based on CLI
// command flags and configuration files. It needs to be called before
// rendering any (styled) output, ideally as a "PersistentPreRun" of a Cobra
// root command.
func HandleStyles() error {
	// Colorization mode...
	switch colorize {
	case CmAlways:
		colorProfile = termenv.ANSI256
	case CmAuto:
		colorProfile = termenv.ColorProfile()
	case CmNever:
		colorProfile = termenv.Ascii
	}
	// First look for a user-defined theme in the user's home directory.
	var th string
	if home, err := os.UserHomeDir(); err == nil {
		if styling, err := ioutil.ReadFile(filepath.Join(home, ".lxknsrc.yaml")); err == nil {
			th = string(styling)
		}
	}
	if th == "" || dumptheme {
		// Theme selection (or dumping): dark or light...
		switch theme {
		case ThDark:
			th = darkTheme
		case ThLight:
			th = lightTheme
		}
	}
	if dumptheme {
		fmt.Fprint(os.Stdout, th)
		os.Exit(0)
	}
	readStyles(th)
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
