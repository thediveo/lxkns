// Copyright 2024 Harald Albrecht.
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

//go:build go1.23

package model

import "iter"

// Tasks returns an iterator over all tasks of the process identified by PID. In
// case of non-existing/invalid PIDs, Tasks returns an empty iterator. If the
// process table does not contain information about the tasks of the process,
// the iterator won't emit any tasks, not even the task group leader
// representing the process.
func (t ProcessTable) Tasks(pid PIDType) iter.Seq[*Task] {
	return Tasks(t, pid)
}

// Tasks returns an iterator over all tasks of the process identified by PID. In
// case of non-existing/invalid PIDs, Tasks returns an empty iterator. If the
// process table does not contain information about the tasks of the process,
// the iterator won't emit any tasks, not even the task group leader
// representing the process.
func Tasks(t ProcessTable, pid PIDType) iter.Seq[*Task] {
	proc := t[pid]
	if proc == nil {
		return func(yield func(*Task) bool) {}
	}
	return func(yield func(*Task) bool) {
		tasksOfProcess(proc, yield)
	}
}

// TasksRecursive returns an iterator over all tasks of the process identified
// by PID, as well as all tasks of children, grandchildren, et cetera. If the
// process table does not contain information about the tasks of the process,
// the iterator won't emit any tasks, not even the task group leader
// representing the process.
func (t ProcessTable) TasksRecursive(pid PIDType) iter.Seq[*Task] {
	return TasksRecursive(t, pid)
}

// TasksRecursive returns an iterator over all tasks of the process identified
// by PID, as well as all tasks of children, grandchildren, et cetera. If the
// process table does not contain information about the tasks of the process,
// the iterator won't emit any tasks, not even the task group leader
// representing the process.
func TasksRecursive(t ProcessTable, pid PIDType) iter.Seq[*Task] {
	proc := t[pid]
	if proc == nil {
		return func(yield func(*Task) bool) {}
	}
	return func(yield func(*Task) bool) {
		tasksRecursive(proc, yield)
	}
}

func tasksRecursive(proc *Process, yield func(*Task) bool) bool {
	if !tasksOfProcess(proc, yield) {
		return false
	}
	for _, child := range proc.Children {
		if !tasksRecursive(child, yield) {
			return false
		}
	}
	return true
}

func tasksOfProcess(proc *Process, yield func(*Task) bool) bool {
	for _, task := range proc.Tasks {
		if !yield(task) {
			return false
		}
	}
	return true
}
