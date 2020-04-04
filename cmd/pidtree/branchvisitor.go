// A visitor implementing the view on the process tree and PID namespaces.

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
	"reflect"

	"github.com/thediveo/lxkns"
)

// BranchVisitor is an asciitree.Visitor which works on a branch from an
// initial/root PID namespace down to a specific process (sic!). It expects a
// sequence of PID namespaces and/or processes, which it will follow until it
// ends.
type BranchVisitor struct {
	Details   bool
	PIDMap    *lxkns.PIDMap
	RootPIDNS lxkns.Namespace
}

// Roots simply returns the specified branch as it, as the Get visitor method
// will take care of all details.
func (v *BranchVisitor) Roots(roots reflect.Value) (children []reflect.Value) {
	return []reflect.Value{roots}
}

// Label returns a node label text, which varies depending on whether the node
// is a Process or a (PID) Namespace. In case of a PID Namespace, the label
// will show the namespace type and its ID, as well as the owner name and UID
// (via the owning user Namespace). If it's a Process instead, then the text
// label contains the name and "global" PID, but also the translated "local"
// PID (which is the PID as seen from inside the PID namespace of the
// Process).
func (v *BranchVisitor) Label(branch reflect.Value) (label string) {
	nodeif := branch.Index(0).Interface()
	if proc, ok := nodeif.(*lxkns.Process); ok {
		// It's a Process; do we have namespace information for it? If yes,
		// then we can translate between the process-local PID namespace and
		// the "initial" PID namespace.
		if procpidns := proc.Namespaces[lxkns.PIDNS]; procpidns != nil {
			localpid := v.PIDMap.Translate(proc.PID, v.RootPIDNS, procpidns)
			if localpid != proc.PID {
				return fmt.Sprintf("%q (%d=%d)", proc.Name, proc.PID, localpid)
			}
			return fmt.Sprintf("%q (%d)", proc.Name, proc.PID)
		}
		// PID namespace information is NOT known, so this is a process out of
		// our reach. We thus print it in a way to signal that we don't know
		// about this process' PID namespace
		return fmt.Sprintf("pid:[???] %q (%d=???)", proc.Name, proc.PID)
	}
	// It's a PID namespace, so we give details about the ID and the owner's
	// UID and name. And if it's not ... PANIC!!!
	pidns := nodeif.(lxkns.Namespace)
	label = pidns.(lxkns.NamespaceStringer).TypeIDString()
	if pidns.Owner() != nil {
		uid := pidns.Owner().(lxkns.Ownership).UID()
		var userstr string
		if u, err := user.LookupId(fmt.Sprintf("%d", uid)); err == nil {
			userstr = fmt.Sprintf(" (%q)", u.Username)
		}
		label += fmt.Sprintf(", owned by UID %d%s", uid, userstr)
	}
	return
}

// Get is called on nodes which can be either (1) PID namespaces or (2)
// processes. For (1), the visitor returns information about the PID
// namespace, but then specifies processes as children. For (2), the visitor
// returns process children, unless these children are in a different PID
// namespace: then, their PID namespaces are returned instead. Polymorphism
// galore!
func (v *BranchVisitor) Get(branch reflect.Value) (
	label string, properties []string, children reflect.Value) {
	// Label for this (1) PID namespace or (2) process.
	label = v.Label(branch)
	// The only child can be either a PID namespace or a process, as we'll
	// find out later ... but there will only be exactly one "child" in any
	// case.
	if b := branch.Interface().([]interface{})[1:]; len(b) > 0 {
		children = reflect.ValueOf([]interface{}{b})
	} else {
		children = reflect.ValueOf([]interface{}{})
	}
	return
}
