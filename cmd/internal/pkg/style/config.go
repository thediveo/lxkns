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
)

// AddFlags adds global CLI command flags related to colorization and styling.
func AddFlags(rootCmd *cobra.Command) {
	pflagCreators.Run(rootCmd)
}

// BeforeCommand needs to be called just before the selected command is run,
// ideally as a "PersistentPreRun" of a Cobra root command. It configures the
// various output rendering styles based on CLI command flags and/or
// configuration files.
func BeforeCommand() error {
	runhooks.Run()
	return nil
}

// PrepareForTest needs to be called during test setup in order to correctly
// initialize styling to a well-known minimal state suitable for testing.
func PrepareForTest() {
	rootCmd := &cobra.Command{
		Run: func(_ *cobra.Command, _ []string) { BeforeCommand() },
	}
	AddFlags(rootCmd)
	rootCmd.SetArgs([]string{
		"--treestyle=line",
		"--color=never",
	})
	_ = rootCmd.Execute()
}

// pflagCreators lists the CLI flag constructor functions to be called in
// order to register these flags. This trick here helps us in keeping things
// modular in this package. According to
// https://golang.org/doc/effective_go.html#initialization,
// variables are initialized before the init functions in a package.
var pflagCreators = flagCreators{}

type flagCreators []func(*cobra.Command)

// Register a creator which then registers root command flags, et cetera.
func (fc *flagCreators) Register(creator func(rootCmd *cobra.Command)) {
	*fc = append(*fc, creator)
}

// Run runs the pflag rootCmd registering functions.
func (fc flagCreators) Run(cmd *cobra.Command) {
	for _, creatorf := range fc {
		creatorf(cmd)
	}
}

// runhooks lists hooks to be run after CLI args/flags have been parsed and
// just before the selected command is executed.
var runhooks = hooks{}

type hooks []func()

// Register a hook to be run immediately before executing the selected command.
func (h *hooks) Register(hook func()) {
	*h = append(*h, hook)
}

// Run all registered hooks.
func (h hooks) Run() {
	for _, hook := range h {
		hook()
	}
}
