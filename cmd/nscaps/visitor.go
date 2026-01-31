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
	"strings"

	"github.com/thediveo/go-asciitree/v2"
	"github.com/thediveo/lxkns/cmd/cli/style"
	incaps "github.com/thediveo/lxkns/cmd/internal/caps"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// NodeVisitor is an asciitree.Visitor which works on a node tree produced by
// combine().
type NodeVisitor struct {
	// render function for namespace icons, where its exact behavior depends on
	// CLI flags.
	NamespaceIcon func(model.Namespace) string
	// render function for namespace references in form of either process names
	// (as well as additional process properties) or file system references.
	NamespaceReferenceLabel func(model.Namespace) string
}

var _ asciitree.Visitor = (*NodeVisitor)(nil)

// Roots simply returns the specified topmost node.
func (v *NodeVisitor) Roots(branch any) (children []any) {
	return []any{branch.([]node)[0]}
}

// Label returns a node label text, which varies depending on whether the node
// is a process node or a namespace node. For a process node we return its
// process name and PID. For namespace nodes we return the usual textual
// representation.
func (v *NodeVisitor) Label(n any) (label string) {
	if pn, ok := n.(*processNode); ok {
		return fmt.Sprintf("process %q (%d)",
			style.ProcessStyle.V(style.ProcessName(pn.proc)),
			pn.proc.PID)
	}
	// We're looking at a namespace node.
	nsn := n.(*namespaceNode)
	ns := nsn.ns
	prefix := ""
	// The style of the namespace label, or rather: the namespace type and ID
	// style, depends for user namespaces on the position of that user namespace
	// with relation to the process.
	var sty *style.Style
	if ns.Type() == species.CLONE_NEWUSER {
		sty = nscapstyles[nsn.targetCapsSummary]
		prefix += nscapmarks[nsn.targetCapsSummary] + " "
	} else {
		sty = style.Styles[ns.Type().Name()]
	}
	if nsn.isTarget {
		prefix += "target "
	}
	// Finally return the rendered namespace label.
	return fmt.Sprintf("%s%s%s %s",
		prefix,
		v.NamespaceIcon(ns),
		sty.V(ns.(model.NamespaceStringer).TypeIDString()),
		v.NamespaceReferenceLabel(ns))
}

var nscapstyles = map[targetCapsSummary]*style.Style{
	incapable: &style.UserNoCapsStyle,
	effcaps:   &style.UserEffCapsStyle,
	allcaps:   &style.UserFullCapsStyle,
}

var nscapmarks = map[targetCapsSummary]string{
	incapable: "⛔",
	effcaps:   "⛛",
	allcaps:   "✓",
}

// Get is called on nodes which can be either (1) namespaces or (2)
// processes. TODO: complete
func (v *NodeVisitor) Get(n any) (label string, properties []string, children []any) {
	// Label for this (1) namespace or (2) process.
	label = v.Label(n)
	// Properties
	if tns, ok := n.(*namespaceNode); ok {
		// It's a namespace node, but it is also the target? Only then add
		// properties: we misuse the tree node properties to show the
		// capabilities in the target namespace.
		if tns.isTarget {
			switch tns.targetCapsSummary {
			case incapable:
				properties = []string{"(no capabilities)"}
			case effcaps:
				if briefCaps {
					properties = []string{"(process effective capabilities)"}
				} else {
					properties = propcaps(incaps.ProcessCapabilities(procPID))
					if len(properties) == 0 {
						properties = []string{"(no effective capabilities)"}
					}
				}
			case allcaps:
				if briefCaps {
					properties = []string{"(ALL capabilities)"}
				} else {
					properties = propcaps(incaps.ProcessCapabilities(1))
				}
			}
		}
	} else {
		if pn, ok := n.(*processNode); ok && showProcCaps {
			// It's the process node, so we want to show the effective
			// capabilities of the process (mis)using the tree node properties.
			if len(pn.caps) > 0 {
				properties = propcaps(pn.caps)
			} else {
				properties = []string{"(no effective capabilities)"}
			}
		}
	}
	// Children
	clist := []any{}
	for _, childn := range n.(node).Children() {
		clist = append(clist, childn)
	}
	children = clist
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
		end := min(len(caps), capsperline)
		fields := []string{}
		for _, c := range caps[:end] {
			fields = append(fields, fmt.Sprintf("%-[1]*s", max, c))
		}
		props = append(props, strings.Join(fields, " "))
		caps = caps[end:]
	}
	return
}
