// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux
// +build linux

package discover

import (
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
)

const procPrefix = "/proc/"

// PIDfromPath returns the PID embedded in a /proc/$PID/... path, or 0 if the
// path doesn't contain a /proc/$PID.
func PIDfromPath(path string) model.PIDType {
	if !strings.HasPrefix(path, procPrefix) {
		return 0
	}
	path = path[len(procPrefix):]
	pidfield := path
	if idx := strings.Index(path, "/"); idx >= 0 {
		pidfield = path[:idx]
	}
	pid, err := strconv.ParseUint(pidfield, 10, 32)
	if err != nil {
		return 0
	}
	return model.PIDType(pid)
}

// NewlyProcfsPathIsBetter returns true if the passed reference path for a
// "newly" found namespace reference is "better" than an already known one for
// that particular namespace. "Better" here is defined as the newly reference
// belonging to a process process that is older than the process with the
// already known namespace reference AND all (newly and known) reference
// elements except for the first one match each other. In all other cases,
// NewlyProcfsPathIsBetter returns false; in particular, when one of the
// namespace references isn't from a process.
func NewlyProcfsPathIsBetter(newly, known model.NamespaceRef, processes model.ProcessTable) bool {
	// We never consider replacing an already known reference when the newly one
	// doesn't even is of the same length.
	if l := len(newly); l < 2 || l != len(known) {
		return false
	}
	// And we'll only consider references any further if both have their first
	// reference elements to be for processes.
	newpid := PIDfromPath(newly[0])
	knownpid := PIDfromPath(known[0])
	if newpid == 0 || knownpid == 0 {
		return false
	}
	// The remaining reference elements must match, otherwise we won't pondering
	// to replace an already known reference...
	for idx := 1; idx < len(newly); idx++ {
		if newly[idx] != known[idx] {
			return false
		}
	}
	log.Warnf("newly: %s, known: %s", newly.String(), known.String())

	newproc := processes[newpid]
	knownproc := processes[knownpid]
	if newproc == nil || knownproc == nil {
		return false
	}
	// prefer the more elder process over the younger one, hoping that it was
	// the one that originally created the bind-mounted reference in its own
	// mount namespace and the younger processes then "inherited" these
	// bind-mounts through mount point replication.
	switch delta := int64(newproc.Starttime - knownproc.Starttime); {
	case delta > 0:
		return false
	case delta == 0:
		return knownproc.PID < newproc.PID // not exactly correct after PID roll-over...
	default:
		return true
	}
}
