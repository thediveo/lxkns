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

package filter

import (
	"slices"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/enumflag/v2"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// Names of the CLI flags provided in this package.
const (
	FilterFlagName      = "filter"
	FilterFlagShorthand = "f"
)

// New returns a filter function configured from the filter flag of the
// passed command, where the returned function returns true if the given
// Linux-kernel namespace passes the namespace type filter.
func New(cmd *cobra.Command) func(model.Namespace) bool {
	filters := cmd.PersistentFlags().Lookup(FilterFlagName).
		Value.(*enumflag.EnumFlagValue[species.NamespaceType]).GetSliceValue()
	filtermask := species.NamespaceType(0)
	for _, f := range filters {
		filtermask |= f
	}
	return func(ns model.Namespace) bool {
		return ns.Type()&filtermask != 0
	}
}

// The default user-controlled namespace filters: showing all types of
// Linux-kernel namespaces.
var defaultNamespaceFilters = []species.NamespaceType{
	species.CLONE_NEWNS,
	species.CLONE_NEWCGROUP,
	species.CLONE_NEWUTS,
	species.CLONE_NEWIPC,
	species.CLONE_NEWUSER,
	species.CLONE_NEWPID,
	species.CLONE_NEWNET,
	species.CLONE_NEWTIME,
}

// Maps namespace type names to their corresponding filter/type constants.
var nsFilterIds = map[species.NamespaceType][]string{
	species.CLONE_NEWNS:     {"mnt", "m"},
	species.CLONE_NEWCGROUP: {"cgroup", "c"},
	species.CLONE_NEWUTS:    {"uts", "u"},
	species.CLONE_NEWIPC:    {"ipc", "i"},
	species.CLONE_NEWUSER:   {"user", "U"},
	species.CLONE_NEWPID:    {"pid", "p"},
	species.CLONE_NEWNET:    {"net", "n"},
	species.CLONE_NEWTIME:   {"time", "t"},
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		SetupCLI, plugger.WithPlugin("lxkns/filter"))
}

// SetupCLI adds the "--filter" flag to the specified command. The filter
// flag accepts a set of namespace type names, either separated by commas, or
// specified using multiple "--filter" flags.
func SetupCLI(cmd *cobra.Command) {
	filtersValue := slices.Clone(defaultNamespaceFilters)
	cmd.PersistentFlags().VarP(
		enumflag.NewSlice(&filtersValue,
			FilterFlagName,
			nsFilterIds,
			enumflag.EnumCaseSensitive),
		FilterFlagName, FilterFlagShorthand,
		"shows only selected namespace types; can be 'cgroup'/'c', 'ipc'/'i', 'mnt'/'m',\n"+
			"'net'/'n', 'pid'/'p', 'time/t', 'user'/'U', 'uts'/'u'")
}
