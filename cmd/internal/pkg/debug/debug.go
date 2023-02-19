// Copyright 2022 Harald Albrecht.
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

package debug

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/log"

	_ "github.com/thediveo/lxkns/log/logrus"
)

// Names of the CLI flags defined and used in this package.
const (
	DebugFlagName = "debug"
	LogFlagName   = "log"
)

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("debug"))
	plugger.Group[cliplugin.BeforeCommand]().Register(
		BeforeCommand, plugger.WithPlugin("debug"))
}

// SetupCLI adds the "--debug" and "--log" flags to the specified command that
// changes the logging level to debug or enable logging at the info level.
func SetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(DebugFlagName, false, "enables debug logging output")
	cmd.PersistentFlags().Bool(LogFlagName, false, "enables logging output (but no debug logging)")
}

// Ensure to enable debug logging before any command finally is executed.
func BeforeCommand(cmd *cobra.Command) error {
	if debug, _ := cmd.PersistentFlags().GetBool(DebugFlagName); debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("debug logging enabled")
	} else if logging, _ := cmd.PersistentFlags().GetBool(LogFlagName); logging {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.FatalLevel)
	}
	return nil
}
