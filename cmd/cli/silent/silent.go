// Copyright 2025 Harald Albrecht.
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

package silent

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/go-plugger/v3"
)

// Names of the CLI flags provided in this package.
const (
	SilentFlagName = "silent"
)

func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("lxkns/silent"),
		plugger.WithPlacement(">clippy/debug"))
	plugger.Group[cliplugin.BeforeCommand]().Register(
		BeforeCommand, plugger.WithPlugin("lxkns/silent"),
		plugger.WithPlacement("<clippy/debug"))
}

// SetupCLI runs after(!) the debug flag's SetupCLI so that we can add our
// "--silent" flag and make it mutually exclusive to the "--debug" flag.
func SetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(SilentFlagName, false, "enables logging output")
	cmd.MarkFlagsMutuallyExclusive(SilentFlagName, debug.DebugFlagName)
}

// BeforeCommand runs before(!) the debug flag's BeforeCommand and raises the
// logging bar when the "--silent" flag has been specified with the command. It
// does so by attaching the forced level to the context of the command.
func BeforeCommand(cmd *cobra.Command) error {
	silence, _ := cmd.PersistentFlags().GetBool(SilentFlagName)
	if !cmd.PersistentFlags().Lookup(SilentFlagName).Changed {
		if ets, ok := cmd.Context().Value(ctxEnjoyTheSilence).(bool); ok {
			silence = ets
		}
	}
	if silence {
		debug.SetLevel(cmd, slog.LevelError)
	}
	return nil
}

type ctxKey int

const (
	ctxEnjoyTheSilence ctxKey = iota
)

// PreferSilence configures the command to default to --silence instead of
// --silence=false; CLI users can still override this by specifying
// “--silence=false”. PreferSilence needs to be called before the BeforeCommand
// plugins chain runs, that is, before clippy.BeforeCommand.
func PreferSilence(cmd *cobra.Command) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, ctxEnjoyTheSilence, true))
}
