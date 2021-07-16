// Renders the label text of a node which can be either a PID Namespace or a
// Process.

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

package main

import (
	"fmt"
	"os/user"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/cmd/internal/pkg/output"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/model"
)

// ProcessLabel returns the text label for a Process, rendering such
// information such as not only the PID and process name, but also translating
// the PID into the process' "own" PID namespace, if it differs from the
// initial/root PID namespace.
func ProcessLabel(proc *model.Process, pidmap model.PIDMapper, rootpidns model.Namespace) string {
	// Do we have namespace information for it? If yes, then we can translate
	// between the process-local PID namespace and the "initial" PID
	// namespace. For convenience, we show all PIDs in all PID namespaces,
	// from the initial PID namespace down to the PID namespace this process
	// is joined to.
	if procpidns := proc.Namespaces[model.PIDNS]; procpidns != nil {
		var s string
		showCgroupController := true

		// If there's a container associated with the process, then show the
		// container's name.
		if proc.Container != nil {
			s = fmt.Sprintf("container %q ", style.ContainerStyle.V(proc.Container.Name))
			showCgroupController = false
		}

		pids := []string{}
		for _, el := range pidmap.NamespacedPIDs(proc.PID, rootpidns) {
			pids = append(pids, strconv.FormatUint(uint64(el.PID), 10))
		}
		if len(pids) > 1 {
			s += fmt.Sprintf("%q (%s)",
				style.ProcessStyle.V(style.ProcessName(proc)),
				strings.Join(pids, "/"))
		} else {
			s += fmt.Sprintf("%q (%d)",
				style.ProcessStyle.V(style.ProcessName(proc)), proc.PID)
		}
		if showCgroupController && proc.CpuCgroup != "" {
			s += fmt.Sprintf(" controlled by %q", style.ControlGroupStyle.V(output.ControlgroupDisplayName(proc.CpuCgroup)))
		}
		return s
	}
	// PID namespace information is NOT known, so this is a process out of
	// our reach. We thus print it in a way to signal that we don't know
	// about this process' PID namespace
	return fmt.Sprintf("%s %q (%d/%s)",
		style.PIDStyle.S("pid:[", style.UnknownStyle.V("???"), "]"),
		style.ProcessStyle.V(style.ProcessName(proc)),
		proc.PID,
		style.UnknownStyle.V("???"))
}

// PIDNamespaceLabel returns the text label for a PID namespace, giving not
// only the details about type (always PID) and ID, but additionally the
// owner's UID and user name.
func PIDNamespaceLabel(pidns model.Namespace) (label string) {
	label = output.NamespaceIcon(pidns) +
		style.PIDStyle.S(pidns.(model.NamespaceStringer).TypeIDString())
	if pidns.Owner() != nil {
		uid := pidns.Owner().(model.Ownership).UID()
		var userstr string
		if u, err := user.LookupId(fmt.Sprintf("%d", uid)); err == nil {
			userstr = fmt.Sprintf(" (%q)", style.OwnerStyle.V(u.Username))
		}
		label += fmt.Sprintf(", owned by UID %d%s",
			style.OwnerStyle.V(uid), userstr)
	}
	return
}
