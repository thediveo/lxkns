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
	"reflect"
	"sort"

	"github.com/thediveo/lxkns"
)

type PIDVisitor struct {
	Details      bool
	PIDMap       *lxkns.PIDMap
	InitialPIDNS lxkns.Namespace
}

// Roots returns the given topmost hierarchical process namespaces sorted; it
// will be called on the list of topmost PID namespace(s). It won't be called
// ever afterwards.
func (v *PIDVisitor) Roots(roots reflect.Value) (children []reflect.Value) {
	pidroots := lxkns.SortNamespaces(roots.Interface().([]lxkns.Namespace))
	count := len(pidroots)
	children = make([]reflect.Value, count)
	for idx := 0; idx < count; idx++ {
		children[idx] = reflect.ValueOf(pidroots[idx])
	}
	return
}

// Label returns the text label for a namespace node. Everything else will have
// no label.
func (v *PIDVisitor) Label(node reflect.Value) (label string) {
	if proc, ok := node.Interface().(*lxkns.Process); ok {
		if procpidns := proc.Namespaces[lxkns.PIDNS]; procpidns != nil {
			localpid := v.PIDMap.Translate(proc.PID, v.InitialPIDNS, procpidns)
			if localpid != proc.PID {
				return fmt.Sprintf("PID %d=%d %q", proc.PID, localpid, proc.Name)
			}
		}
		return fmt.Sprintf("PID %d %q", proc.PID, proc.Name)
	}
	pidns := node.Interface().(lxkns.Namespace)
	label = fmt.Sprintf("%s, owned by UID %d",
		pidns.(lxkns.NamespaceStringer).TypeIDString(),
		pidns.Owner().(lxkns.Ownership).UID())
	return
}

// Get is called on nodes which can be either (1) PID namespaces or (2)
// processes. For (1), the visitor returns information about the PID
// namespace, but then specifies processes as children. For (2), the visitor
// returns process children, unless these children are in a different PID
// namespace: then, their PID namespaces are returned instead. Polymorphism
// galore!
func (v *PIDVisitor) Get(node reflect.Value) (
	label string, properties []string, children reflect.Value) {
	// Label for this (1) PID namespace or (2) process.
	label = v.Label(node)
	// Children of this (1) PID namespace are always processes, but for (2)
	// processes the children can be any combination of (a) child processes
	// still in the same namespace and (b) child PID namespaces.
	clist := []interface{}{}
	if proc, ok := node.Interface().(*lxkns.Process); ok {
		// TODO:
		pidns := proc.Namespaces[lxkns.PIDNS]
		childprocesses := lxkns.ProcessListByPID(proc.Children)
		sort.Sort(childprocesses)
		for _, childproc := range childprocesses {
			if childproc.Namespaces[lxkns.PIDNS] == pidns {
				clist = append(clist, childproc)
			} else {
				clist = append(clist, childproc.Namespaces[lxkns.PIDNS])
			}
		}
	} else {
		// The child nodes of a PID namespace tree node will be the "leader"
		// (or "topmost") processes inside the PID namespace.
		leaders := lxkns.ProcessListByPID(node.Interface().(lxkns.Namespace).Leaders())
		sort.Sort(leaders)
		for _, proc := range leaders {
			clist = append(clist, proc)
		}
	}
	children = reflect.ValueOf(clist)
	return
}
