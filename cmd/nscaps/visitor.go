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
	"fmt"
	"reflect"
	"strings"

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/output"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/species"
)

// NodeVisitor is an asciitree.Visitor which works on a node tree produced by
// combine().
type NodeVisitor struct {
}

// Roots simply returns the specified topmost node.
func (v *NodeVisitor) Roots(branch reflect.Value) (children []reflect.Value) {
	return []reflect.Value{branch.Index(0)}
}

// Label returns a node label text, which varies depending on whether the node
// is a process node or a namespace node. For a process node we return its
// process name and PID. For namespace nodes we return the usual textual
// representation.
func (v *NodeVisitor) Label(n reflect.Value) (label string) {
	if n.Kind() == reflect.Ptr {
		n = n.Elem()
	}
	if pn, ok := n.Interface().(processnode); ok {
		return fmt.Sprintf("process %q (%d)",
			style.ProcessStyle.V(style.ProcessName(pn.proc)),
			pn.proc.PID)
	}
	// We're looking at a namespace node.
	nsn := n.Interface().(nsnode)
	ns := nsn.ns
	prefix := ""
	// The style of the namespace label, or rather: the namespace type and ID
	// style, depends for user namespaces on the position of that user namespace
	// with relation to the process.
	var sty *style.Style
	if ns.Type() == species.CLONE_NEWUSER {
		sty = nscapstyles[nsn.targetcaps]
		prefix += nscapmarks[nsn.targetcaps] + " "
	} else {
		sty = style.Styles[ns.Type().Name()]
	}
	if nsn.istarget {
		prefix += "target "
	}
	// Finally return the rendered namespace label.
	return fmt.Sprintf("%s%s%s %s",
		prefix,
		output.NamespaceIcon(ns),
		sty.V(ns.(lxkns.NamespaceStringer).TypeIDString()),
		output.NamespaceReferenceLabel(ns))
}

var nscapstyles = map[targetcaps]*style.Style{
	incapable: &style.UserNoCapsStyle,
	effcaps:   &style.UserEffCapsStyle,
	allcaps:   &style.UserFullCapsStyle,
}

var nscapmarks = map[targetcaps]string{
	incapable: "⛔",
	effcaps:   "⛛",
	allcaps:   "✓",
}

// Get is called on nodes which can be either (1) namespaces or (2)
// processes. TODO: complete
func (v *NodeVisitor) Get(n reflect.Value) (
	label string, properties []string, children reflect.Value) {
	if n.Kind() == reflect.Ptr {
		n = n.Elem()
	}
	// Label for this (1) namespace or (2) process.
	label = v.Label(n)
	// Properties
	if tns, ok := n.Interface().(nsnode); ok {
		if tns.istarget {
			switch tns.targetcaps {
			case incapable:
				properties = []string{"(no capabilities)"}
			case effcaps:
				if briefCaps {
					properties = []string{"(process effective capabilities)"}
				} else {
					properties = propcaps(ProcessCapabilities(procPID))
				}
			case allcaps:
				if briefCaps {
					properties = []string{"(ALL capabilities)"}
				} else {
					properties = propcaps(ProcessCapabilities(1))
				}
			}
		}
	} else {
		if pn, ok := n.Interface().(processnode); ok && showProcCaps {
			if len(pn.caps) > 0 {
				properties = propcaps(pn.caps)
			} else {
				properties = []string{"(no capabilities)"}
			}
		}
	}
	// Children
	clist := []interface{}{}
	for _, childn := range n.Interface().(node).Children() {
		clist = append(clist, childn)
	}
	children = reflect.ValueOf(clist)
	return
}

const capsperline = 4

func propcaps(caps []string) (props []string) {
	max := 0
	for _, c := range caps {
		if len(c) > max {
			max = len(c)
		}
	}
	for len(caps) > 0 {
		end := len(caps)
		if end > capsperline {
			end = capsperline
		}
		fields := []string{}
		for _, c := range caps[:end] {
			fields = append(fields, fmt.Sprintf("%-[1]*s", max, c))
		}
		props = append(props, strings.Join(fields, " "))
		caps = caps[end:]
	}
	return
}
