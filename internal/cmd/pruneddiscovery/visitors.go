// Copyright 2026 Harald Albrecht.
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
	"github.com/thediveo/lxkns/internal/xslices"
	"github.com/thediveo/lxkns/model"
)

// processVisitor helps asciitree.Render render the hierarchy of processes with
// their tasks. This is mainly for debugging purposes.
//
// Each node consists of the process or task name, followed by its PID or TID.
// For processes, child nodes list the tasks first (sorted by TID), then child
// processes (sorted by name then PID).
type processVisitor struct{}

var _ asciitree.Visitor = (*processVisitor)(nil)

func (v *processVisitor) Roots(roots any) (children []any) {
	return xslices.Any(roots.([]*model.Process))
}

func (v *processVisitor) Label(branch any) (label string) {
	switch br := branch.(type) {
	case *model.Process:
		return fmt.Sprintf("%s (%d)", br.Name, br.PID)
	case *model.Task:
		return fmt.Sprintf("%s [%d]", br.Name, br.TID)
	}
	return ""
}

func (v *processVisitor) Get(branch any) (label string, properties []string, children []any) {
	label = v.Label(branch)
	switch br := branch.(type) {
	case *model.Process:
		// Only visit tasks of this process if there is more than one task
		// (which is probably the task group leader representing the process).
		if len(br.Tasks) > 1 {
			children = xslices.Any(xslices.SortedCopy(br.Tasks,
				func(a, b *model.Task) int { return int(a.TID) - int(b.TID) }))
		}
		// Always visit the children.
		children = append(children,
			xslices.Any(xslices.SortedCopy(br.Children,
				func(a, b *model.Process) int {
					if d := strings.Compare(a.Name, b.Name); d != 0 {
						return d
					}
					return int(a.PID) - int(b.PID)
				}))...,
		)
	}
	return
}

type hierarchyVisitor struct{}

var _ asciitree.Visitor = (*hierarchyVisitor)(nil)

func (v *hierarchyVisitor) Roots(roots any) (children []any) {
	return xslices.Any(roots.([]model.Hierarchy))
}

func (v *hierarchyVisitor) Label(branch any) (label string) {
	return branch.(model.NamespaceStringer).TypeIDString()
}

func (v *hierarchyVisitor) Get(branch any) (label string, properties []string, children []any) {
	label = v.Label(branch)
	children = xslices.Any(branch.(model.Hierarchy).Children())
	return
}
