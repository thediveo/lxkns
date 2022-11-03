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

package cli

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
)

// AddFlags runs all registered SetupCLI plugin functions (in group "cli") in
// order to register CLI flags for the specified root command.
func AddFlags(rootCmd *cobra.Command) {
	for _, setupCLI := range plugger.Group[cliplugin.SetupCLI]().Symbols() {
		setupCLI(rootCmd)
	}
}

// BeforeCommand runs all registered BeforeRun plugin functions (in group "cli")
// just before the selected command runs; it terminates as soon as the first
// plugin function returns a non-nil error.
func BeforeCommand() error {
	for _, beforeCmd := range plugger.Group[cliplugin.BeforeCommand]().Symbols() {
		if err := beforeCmd(); err != nil {
			return err
		}
	}
	return nil
}
