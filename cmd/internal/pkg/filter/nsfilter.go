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
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/nstypes"
)

// AddFilterFlag adds the "--filter" flag to the specified command. The filter
// flag accepts a set of namespace type names, either separated by commas, or
// specified using multiple "--filter" flags.
func AddFilterFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().VarP(
		enumflag.NewSlice(&namespaceFilters, "filter", nsFilterIds, enumflag.EnumCaseSensitive),
		"filter", "f",
		"shows only selected namespace types; can be 'cgroup'/'c', 'ipc'/'i', 'mnt'/'m',\n"+
			"'net'/'n', 'pid'/'p', 'user'/'U', 'uts'/'u'")
}

// Filter returns true if the given Linux-kernel namespace passes the namespace
// type filter.
func Filter(ns lxkns.Namespace) bool {
	if filterMask == 0 {
		for _, f := range namespaceFilters {
			filterMask |= f
		}
	}
	return ns.Type()&filterMask != 0
}

var filterMask nstypes.NamespaceType = 0

// The user-controlled namespace filters; they default to showing all types of
// Linux-kernel namespaces.
var namespaceFilters = []nstypes.NamespaceType{
	nstypes.CLONE_NEWNS,
	nstypes.CLONE_NEWCGROUP,
	nstypes.CLONE_NEWUTS,
	nstypes.CLONE_NEWIPC,
	nstypes.CLONE_NEWUSER,
	nstypes.CLONE_NEWPID,
	nstypes.CLONE_NEWNET,
}

// Maps namespace type names to their corresponding filter/type constants.
var nsFilterIds = map[nstypes.NamespaceType][]string{
	nstypes.CLONE_NEWNS:     {"mnt", "m"},
	nstypes.CLONE_NEWCGROUP: {"cgroup", "c"},
	nstypes.CLONE_NEWUTS:    {"uts", "u"},
	nstypes.CLONE_NEWIPC:    {"ipc", "i"},
	nstypes.CLONE_NEWUSER:   {"user", "U"},
	nstypes.CLONE_NEWPID:    {"pid", "p"},
	nstypes.CLONE_NEWNET:    {"net", "n"},
}
