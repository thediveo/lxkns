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

package reflabel

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/cli/cgrp"
	"github.com/thediveo/lxkns/cmd/cli/style"
	"github.com/thediveo/lxkns/internal/xslices"
	"github.com/thediveo/lxkns/model"
)

// Names of the CLI flags defined and used in this package.
const (
	AllLeadersFlagName = "all-leaders"
)

// NamespaceReferenceLabel returns a string describing a reference to the
// specified namespace, either in form of a (leader) process name and PID, or if
// there is no such process then in form of a filesystem reference.
func NamespaceReferenceLabel(cmd *cobra.Command) func(model.Namespace) string {
	allLeaders, _ := cmd.PersistentFlags().GetBool(AllLeadersFlagName)
	cgroupDisplayName := cgrp.CgroupDisplayName(cmd)
	return func(ns model.Namespace) string {
		if ancient := ns.Ealdorman(); ancient != nil {
			return namespaceProcessLabel(ns, ancient, allLeaders, cgroupDisplayName)
		}
		// No leaders, so maybe one or more loose threads, perchance?
		if looseThreads := ns.LooseThreads(); len(looseThreads) > 0 {
			return namespaceLooseThreadLabel(looseThreads)
		}
		// There were no leaders and no loose threads, so all we're left with now is
		// a "bind-mounted" reference.
		if ref := ns.Ref(); len(ref) != 0 {
			if len(ref) > 0 && strings.HasPrefix(ref[0], "/proc/") {
				return fmt.Sprintf("referenced from process/task %q",
					style.PathStyle.V(ref.String()))
			}
			return fmt.Sprintf("bind-mounted at %q",
				style.PathStyle.V(ref.String()))
		}
		// Hmpf; we actually don't know much about this namespace.
		return ""
	}
}

// namespaceProcessLabel returns a rendered label for the namespace with the
// attached containers and processes information.
func namespaceProcessLabel(ns model.Namespace, ealdorman *model.Process, allLeaders bool, cgroupDisplayName func(string) string) string {
	// The earldorman always comes first ... age before beauty.
	procs := []*model.Process{ealdorman}
	if allLeaders {
		// Sort the leader processes by their PID and then add the other
		// leaders afterwards, so without doubt we can term these leaders
		// "trailers" or "followers"...
		for _, proc := range xslices.SortedCopy(ns.Leaders(), orderProcessByPID) {
			if proc != ealdorman {
				procs = append(procs, proc)
			}
		}
	}

	var s string
	for idx, proc := range procs {
		if idx > 0 {
			s += ", "
		}
		showCgroupController := true
		// If there's a container associated with this particular process
		// then additionally show the container's name.
		if proc.Container != nil {
			s = fmt.Sprintf("container %q ", style.ContainerStyle.V(proc.Container.Name))
			showCgroupController = false
		}
		// Now render the process information.
		s += fmt.Sprintf("process %q (%d)",
			style.ProcessStyle.V(style.ProcessName(proc)),
			proc.PID)
		if showCgroupController && proc.CpuCgroup != "" {
			s += fmt.Sprintf(
				" controlled by %q",
				style.ControlGroupStyle.V(cgroupDisplayName(proc.CpuCgroup)))
		}
	}
	return s
}

// orderProcessByPID is a less function that returns true if the PID of a first
// process is lower than that of a second process.
func orderProcessByPID(e1, e2 *model.Process) int {
	return int(e1.PID) - int(e2.PID)
}

// namespaceLooseThreadLabel returns a rendered label for the namespace with the
// attached loose threads (tasks). In this case there is no process information
// available and thus not rendered.
func namespaceLooseThreadLabel(looseThreads []*model.Task) string {
	looseThreads = xslices.SortedCopy(looseThreads, func(t1, t2 *model.Task) int {
		return int(t1.TID) - int(t2.TID)
	})
	s := ""
	for idx, task := range looseThreads {
		if idx > 0 {
			s += ", "
		}
		if task.Process != nil && task.Process.Container != nil {
			s = fmt.Sprintf("container %q ", style.ContainerStyle.V(task.Process.Container.Name))
		}
		s += fmt.Sprintf("task %q [%d]",
			style.TaskStyle.V(task.Name), task.TID)
		if task.Process != nil {
			s += fmt.Sprintf(" of %q (%d)",
				style.ProcessStyle.V(style.ProcessName(task.Process)), task.Process.PID)
		}
	}
	return s
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		LabelSetupCLI, plugger.WithPlugin("lxkns/ref-label"))
}

// LabelSetupCLI adds the flags for controlling control group name display.
func LabelSetupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(AllLeadersFlagName, false,
		"show all leader processes instead of only the most senior one")
}
