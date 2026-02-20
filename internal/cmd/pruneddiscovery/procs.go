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
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
)

// matchAny returns a compiled regular expression that matches any of the
// specified patterns.
func matchAny(patterns []string) *regexp.Regexp {
	var bigBeautifulPattern strings.Builder
	bigBeautifulPattern.WriteString("^(")
	for idx, pattern := range patterns {
		if idx > 0 {
			bigBeautifulPattern.WriteRune('|')
		}
		bigBeautifulPattern.WriteString(pattern)
	}
	bigBeautifulPattern.WriteString(")$")
	return regexp.MustCompile(bigBeautifulPattern.String())
}

// prunes all processes (together with their tasks) from the process
// table/hierarchy that doesn't match the passed regexp matcher. The removed
// processes and tasks are also detached from their namespaces. Processes in
// between the root of the process hierarchy and a process to be kept are kept
// too.
func pruneProcesses(d *discover.Result, re *regexp.Regexp) {
	keep := map[model.PIDType]struct{}{}
	// First scan through all processes, deciding which processes to keep;
	// either directly or indirectly because they're needed to keep a grandchild
	// process.
	for pid := range d.Processes {
		proc, ok := d.Processes[pid]
		if !ok {
			continue
		}
		if proc.Container == nil && !re.MatchString(proc.Name) {
			continue
		}
		purgeTasks(proc)
		// mark this process as well as all its grandparent processes as to be
		// kept.
		keep[pid] = struct{}{}
		proc = proc.Parent
		for proc != nil {
			if _, kept := keep[proc.PID]; kept {
				break
			}
			purgeTasks(proc)
			keep[proc.PID] = struct{}{}
			proc = proc.Parent
		}
	}
	// Now go through all processes a second time, purging all that are not on
	// our positive list.
	for _, pid := range slices.Collect(maps.Keys(d.Processes)) {
		proc, ok := d.Processes[pid]
		if !ok {
			continue
		}
		if _, keepit := keep[pid]; keepit {
			continue
		}
		purgeProcess(proc)
		delete(d.Processes, pid)
	}
}

func sanitizeProcesses(d *discover.Result) {
	// For all the processes to keep now remove any CLI arguments
	for _, proc := range d.Processes {
		proc.Cmdline = []string{proc.Name}
	}
}

// purgeTasks of a process, except for the task group leader itself.
func purgeTasks(proc *model.Process) {
	proc.Tasks = slices.DeleteFunc(proc.Tasks,
		func(task *model.Task) bool {
			del := task.TID != proc.PID
			if del {
				detachTaskNamespaces(task)
			}
			return del
		})
}

// detachTaskNamespaces detaches the passed task from its namespaces, where
// necessary.
func detachTaskNamespaces(task *model.Task) {
	for namespace := range allUnzeros(task.Namespaces[:]) {
		if namespace == nil {
			continue
		}
		asPlain(namespace).RemoveLooseThread(task)
	}
}

// purgeProcess recursively unlinks the specified process from the process
// hierarchy. It additionally detaches the process and its tasks from their
// namespaces. Purging is done depth-first.
func purgeProcess(proc *model.Process) {
	// remove from hierarchy, first recursively all children, then ourselves. As
	// this modifies the list of children we better work on a list copy ;)
	for _, child := range slices.Clone(proc.Children) {
		purgeProcess(child)
	}
	if proc.Parent != nil {
		proc.Parent.Children = slices.DeleteFunc(proc.Parent.Children, func(child *model.Process) bool {
			return child.PID == proc.PID
		})
		proc.Parent = nil // be kind to the GC
	}
	// detach process and tasks from namespaces, where necessary.
	for namespace := range allUnzeros(proc.Namespaces[:]) {
		asPlain(namespace).RemoveLeader(proc)
		for _, task := range proc.Tasks {
			detachTaskNamespaces(task)
		}
	}
}
