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
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/cmd/internal/tool"
	"github.com/thediveo/lxkns/model"
)

// orderProcessByPID is a less function that returns true if the PID of a first
// process is lower than that of a second process.
func orderProcessByPID(e1, e2 *model.Process) int {
	return int(e1.PID) - int(e2.PID)
}

// NamespaceReferenceLabel returns a string describing a reference to the
// specified namespace, either in form of a (leader) process name and PID, or if
// there is no such process then in form of a filesystem reference.
func NamespaceReferenceLabel(ns model.Namespace) string {
	if ancient := ns.Ealdorman(); ancient != nil {
		return namespaceProcessLabel(ns, ancient)
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

// namespaceProcessLabel returns a rendered label for the namespace with the
// attached containers and processes information.
func namespaceProcessLabel(ns model.Namespace, ealdorman *model.Process) string {
	// The earldorman always comes first ... age before beauty.
	procs := []*model.Process{ealdorman}
	if allLeaders {
		// Sort the leader processes by their PID and then add the other
		// leaders afterwards, so without doubt we can term these leaders
		// "trailers" or "followers"...
		for _, proc := range tool.Sort(ns.Leaders(), orderProcessByPID) {
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
				style.ControlGroupStyle.V(ControlgroupDisplayName(proc.CpuCgroup)))
		}
	}
	return s
}

// namespaceLooseThreadLabel returns a rendered label for the namespace with the
// attached loose threads (tasks). In this case there is no process information
// available and thus not rendered.
func namespaceLooseThreadLabel(looseThreads []*model.Task) string {
	looseThreads = tool.Sort(looseThreads, func(t1, t2 *model.Task) int {
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

// ControlgroupDisplayName takes a control group name (path) and, depending on
// the display flags set, returns a name better suited for display. In
// particular, it optionally shortens 64 hex digit IDs as used by Docker for
// identifying containers to the Docker-typical 12 hex digit "digest".
func ControlgroupDisplayName(s string) string {
	if controlGroupNames == CgroupComplete {
		return s
	}
	labels := strings.Split(s, "/")
	for idx, label := range labels {
		if len(label) == 64 && ishex(label) {
			labels[idx] = label[:12] + "…"
		}
	}
	return strings.Join(labels, "/")
}

// ishex checks if the given string solely consists of ASCII hex digits, and
// nothing else, then return true.
func ishex(hex string) bool {
	for _, char := range hex {
		if !unicode.In(char, unicode.ASCII_Hex_Digit) {
			return false
		}
	}
	return true
}

// allLeaders switches on/off displaying all leader processes in a given
// namespace, or only the most senior "ealdorman" process (to reduce noise).
var allLeaders bool

// controlGroupNames switches between control group name shorting and full glory.
var controlGroupNames ControlGroupNames

// ControlGroupNames defines the enumeration flag type for controlling
// optimizing control group names for display (or not).
type ControlGroupNames enumflag.Flag

const (
	// CgroupShortened enables optimizing the display of Docker container IDs.
	CgroupShortened ControlGroupNames = iota
	// CgroupComplete switches off any display optimization of control group
	// names.
	CgroupComplete
)

// ControlGroupNameModes specifies the mapping between the user-facing CLI flag
// values and the program-internal flag values.
var ControlGroupNameModes = map[ControlGroupNames][]string{
	CgroupShortened: {"short"},
	CgroupComplete:  {"full", "complete"},
}

// Register our plugin functions for delayed registration of CLI flags we bring
// into the game and the things to check or carry out before the selected
// command is finally run.
func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		LabelSetupCLI, plugger.WithPlugin("controlgroup"))
}

// LabelSetupCLI adds the flags for controlling control group name display.
func LabelSetupCLI(cmd *cobra.Command) {
	controlGroupNames = CgroupShortened // ensure clean initial state for testing
	cmd.PersistentFlags().Var(
		enumflag.New(&controlGroupNames, "cgformat", ControlGroupNameModes, enumflag.EnumCaseInsensitive),
		"cgroup",
		"control group name display; can be 'full' or 'short'")
	//cmd.PersistentFlags().Lookup("cgroup").NoOptDefVal = "short"
	allLeaders = false
	cmd.PersistentFlags().BoolVar(&allLeaders, "all-leaders", false,
		"show all leader processes instead of only the most senior one")
}
