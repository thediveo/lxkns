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

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
)

// ProcessLabel returns the text label for a Process, rendering such
// information such as not only the PID and process name, but also translating
// the PID into the process' "own" PID namespace, if it differs from the
// initial/root PID namespace.
func ProcessLabel(proc *lxkns.Process, pidmap *lxkns.PIDMap, rootpidns lxkns.Namespace) string {
	// Do we have namespace information for it? If yes, then we can translate
	// between the process-local PID namespace and the "initial" PID
	// namespace.
	if procpidns := proc.Namespaces[lxkns.PIDNS]; procpidns != nil {
		localpid := pidmap.Translate(proc.PID, rootpidns, procpidns)
		if localpid != proc.PID {
			return fmt.Sprintf("%q (%d=%d)",
				style.ProcessStyle.V(proc.Name),
				proc.PID, localpid)
		}
		return fmt.Sprintf("%q (%d)", style.ProcessStyle.V(proc.Name), proc.PID)
	}
	// PID namespace information is NOT known, so this is a process out of
	// our reach. We thus print it in a way to signal that we don't know
	// about this process' PID namespace
	return fmt.Sprintf("%s %q (%d=%s)",
		style.PIDStyle.S("pid:[", style.UnknownStyle.V("???"), "]"),
		style.ProcessStyle.V(proc.Name),
		proc.PID,
		style.UnknownStyle.V("???"))
}

// PIDNamespaceLabel returns the text label for a PID namespace, giving not
// only the details about type (always PID) and ID, but additionally the
// owner's UID and user name.
func PIDNamespaceLabel(pidns lxkns.Namespace) (label string) {
	label = style.PIDStyle.S(pidns.(lxkns.NamespaceStringer).TypeIDString())
	if pidns.Owner() != nil {
		uid := pidns.Owner().(lxkns.Ownership).UID()
		var userstr string
		if u, err := user.LookupId(fmt.Sprintf("%d", uid)); err == nil {
			userstr = fmt.Sprintf(" (%q)", style.OwnerStyle.V(u.Username))
		}
		label += fmt.Sprintf(", owned by UID %d%s",
			style.OwnerStyle.V(uid), userstr)
	}
	return
}
