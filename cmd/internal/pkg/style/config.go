// This implements central part of the style configuration mechanism: packages
// consuming this package must call AddFlags() after they've created their
// root command, as well as BeforeCommand() before their command executes,
// preferably as PersistentPreRunE of their root command.
//
// Other parts of this package then hook themselves automatically into the two
// phases of flag creation and just-before-command-execution. This helps
// keeping this package more modular.

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
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
)

// PrepareForTest needs to be called during test setup in order to correctly
// initialize styling to a well-known minimal state suitable for testing.
func PrepareForTest() {
	rootCmd := &cobra.Command{
		Run: func(_ *cobra.Command, _ []string) {
			_ = cli.BeforeCommand()
		},
	}
	cli.AddFlags(rootCmd)
	rootCmd.SetArgs([]string{
		"--treestyle=line",
		"--color=never",
	})
	_ = rootCmd.Execute()
}
