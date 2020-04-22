// Convenience functions for rendering various namespace properties as text.

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

package output

import (
	"fmt"

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
)

// NamespaceReferenceLabel returns a string describing a reference to the
// specified namespace, either in form of a (leader) process name and PID, or if
// there is no such process then in form of a filesystem reference.
//
// TODO: allow styling with simplified versus full representation (ealdorman
// versus all leader processes)
func NamespaceReferenceLabel(ns lxkns.Namespace) string {
	if ancient := ns.Ealdorman(); ancient != nil {
		return fmt.Sprintf("process %q (%d)",
			style.ProcessStyle.V(style.ProcessName(ancient)),
			ancient.PID)
	}
	if ref := ns.Ref(); ref != "" {
		// TODO: deal with references in other mount namespaces :)
		return fmt.Sprintf("bind-mounted at %q",
			ref)
	}
	return ""
}

/*
	if leaders := ns.Leaders(); len(leaders) > 0 {
			sorted := make([]*lxkns.Process, len(leaders))
			copy(sorted, leaders)
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].PID < sorted[j].PID
			})
			s := []string{}
			for _, leader := range sorted {
				s = append(s, fmt.Sprintf("%q (%d)", leader.Name, leader.PID))
			}
			procs = strings.Join(s, ", ")
			if len(leaders) > 1 {
				procs = "processes " + procs
			} else {
				procs = "process " + procs
			}
	}
*/
