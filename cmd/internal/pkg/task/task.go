// Copyright 2023 Harald Albrecht.
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

package task

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
)

// Names of the CLI flags defined and used in this package.
const (
	TaskFlagName = "task"
)

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("task"))
}

// SetupCLI adds the "task" flag to enable/disable task discovery.
func SetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(TaskFlagName, true, "discover also tasks")
}

// Enabled returns true if tasks should be discovered, otherwise false.
func Enabled(cmd *cobra.Command) bool {
	enabled, _ := cmd.PersistentFlags().GetBool(TaskFlagName)
	return enabled
}
