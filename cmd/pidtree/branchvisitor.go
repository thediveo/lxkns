// A visitor implementing the single-branch view on the process tree and PID
// namespaces.

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
	"reflect"

	"github.com/thediveo/lxkns/model"
)

// BranchVisitor is an asciitree.Visitor which works on a single branch from
// an initial/root PID namespace going down to a specific process (sic!). It
// expects a sequence of PID namespaces and/or processes, which it will follow
// until the branch ends.
type BranchVisitor struct {
	Details   bool
	PIDMap    model.PIDMapper
	RootPIDNS model.Namespace
}

// Roots simply returns the specified branch as the only root, as the Get
// visitor method will take care of all details.
func (v *BranchVisitor) Roots(branch reflect.Value) (children []reflect.Value) {
	return []reflect.Value{branch.Index(0)}
}

// Label returns a node label text, which varies depending on whether the node
// is a Process or a (PID) Namespace. In case of a PID Namespace, the label
// will show the namespace type and its ID, as well as the owner name and UID
// (via the owning user Namespace). If it's a Process instead, then the text
// label contains the name and "global" PID, but also the translated "local"
// PID (which is the PID as seen from inside the PID namespace of the
// Process).
func (v *BranchVisitor) Label(branch reflect.Value) (label string) {
	nodeif := branch.Interface().(SingleBranch).Branch[0]
	if proc, ok := nodeif.(*model.Process); ok {
		return ProcessLabel(proc, v.PIDMap, v.RootPIDNS)
	}
	return PIDNamespaceLabel(nodeif.(model.Namespace))
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
	clist := []interface{}{}
	if b := branch.Interface().(SingleBranch).Branch[1:]; len(b) > 0 {
		subbranch := SingleBranch{Branch: b}
		clist = append(clist, subbranch)
	}
	children = reflect.ValueOf(clist)
	return
}
