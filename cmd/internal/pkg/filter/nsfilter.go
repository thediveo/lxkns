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
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"github.com/thediveo/go-plugger/v2"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// Filter returns true if the given Linux-kernel namespace passes the namespace
// type filter.
func Filter(ns model.Namespace) bool {
	if filterMask == 0 {
		for _, f := range namespaceFilters {
			filterMask |= f
		}
	}
	return ns.Type()&filterMask != 0
}

// filterMask is a set of OR'ed namespace CLONE_NEWxx constants indicating the
// type of namespaces allowed to pass the filter.
var filterMask species.NamespaceType

// The user-controlled namespace filters; they default to showing all types of
// Linux-kernel namespaces.
var namespaceFilters = []species.NamespaceType{
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
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Register(
		plugger.WithName("filter"),
		plugger.WithGroup(cliplugin.Group),
		plugger.WithNamedSymbol("SetupCLI", FilterSetupCLI))
}

// FilterSetupCLI adds the "--filter" flag to the specified command. The filter
// flag accepts a set of namespace type names, either separated by commas, or
// specified using multiple "--filter" flags.
func FilterSetupCLI(cmd *cobra.Command) {
	filterMask = species.NamespaceType(0) // ensure clean initial state for testing
	cmd.PersistentFlags().VarP(
		enumflag.NewSlice(&namespaceFilters, "filter", nsFilterIds, enumflag.EnumCaseSensitive),
		"filter", "f",
		"shows only selected namespace types; can be 'cgroup'/'c', 'ipc'/'i', 'mnt'/'m',\n"+
			"'net'/'n', 'pid'/'p', 'user'/'U', 'uts'/'u'")
}
