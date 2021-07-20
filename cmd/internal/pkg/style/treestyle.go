// Implements controlling the tree rendering style via the "--treestyle" CLI
// flag, which can be either "line", or "ascii" ()="plain").

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
	"github.com/thediveo/enumflag"
	asciitree "github.com/thediveo/go-asciitree"
	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
)

// NamespaceStyler styles namespace hierarchies (trees) using the selected
// tree style. This object can directly be used by other packages consuming
// our cmd/internal/style package. This styler object is correctly set when
// the particular (cobra) command runs.
var NamespaceStyler *asciitree.TreeStyler

// The CLI flag treestyle selects the specific style for rendering trees in
// the terminal.
var treestyle TreeStyle

// TreeStyle is an enumeration for selecting a specific tree style.
type TreeStyle enumflag.Flag

// Enumeration of allowed Theme values.
const (
	TreeStyleLine  TreeStyle = iota // default tree line style
	TreeStyleAscii                  // simple ASCII tree style
)

// Defines the textual representations for the TreeStyle values.
var treeStyleIds = map[TreeStyle][]string{
	TreeStyleLine:  {"line"},
	TreeStyleAscii: {"ascii", "plain"},
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:  "treestyle",
		Group: cliplugin.Group,
		Symbols: []plugger.Symbol{
			plugger.NamedSymbol{Name: "SetupCLI", Symbol: TreeStyleSetupCLI},
			plugger.NamedSymbol{Name: "BeforeRun", Symbol: TreeStyleBeforeRun},
		},
	})
}

// TreeStyleSetupCLI is a plugin function that registers the CLI "treestyle"
// flag.
func TreeStyleSetupCLI(rootCmd *cobra.Command) {
	treestyle = TreeStyleLine
	rootCmd.PersistentFlags().Var(
		enumflag.New(&treestyle, "treestyle", treeStyleIds, enumflag.EnumCaseSensitive),
		"treestyle",
		"select the tree render style; can be 'line' or 'ascii'")
}

// TreeStyleBeforeRun is a plugin function that handles selection, reading, or
// dumping of styling profiles, just before the selected command runs. In case
// of dumping, it also exits this process, so the itself command won't ever
// start.
func TreeStyleBeforeRun() error {
	switch treestyle {
	case TreeStyleLine:
		NamespaceStyler = asciitree.NewTreeStyler(asciitree.TreeStyle{
			Fork:     "├", // Don't print this on an FX-80/100 ;)
			Nodeconn: "─",
			Nofork:   "│",
			Lastnode: "└",
			Property: "⋄─",
		})
	case TreeStyleAscii:
		NamespaceStyler = asciitree.NewTreeStyler(asciitree.TreeStyle{
			Fork:     `\`,
			Nodeconn: "_",
			Nofork:   "|",
			Lastnode: `\`,
			Property: "o-",
		})
	}
	return nil
}
