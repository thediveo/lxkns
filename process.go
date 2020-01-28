// Rather simple Linux process tree discovery and representation; that is,
// (go)psutil for the really "boor". It just has to suffice our needs, so
// there's no multi-platform support.

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

// +build linux

package lxkns

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// PIDType expresses things more clearly. And no, that's not a "PidType" since
// "PID" is an acronym (https://en.wikipedia.org/wiki/Acronym), but neither an
// abbreviation, nor an ordinary word (yet/still) in itself.
type PIDType int32

// Process represents our very limited view and even more limited interest in
// a specific Linux process.
type Process struct {
	PID        PIDType       // this process' identifier.
	PPID       PIDType       // parent's process identifier.
	Parent     *Process      // our parent's process description.
	Name       string        // synthesized name of process.
	Namespaces NamespacesSet // the 7 namespaces joined by this process.
	Starttime  uint64        // Time of process start, since the Kernel boot epoch.
}

// ProcessTable maps PIDs to their Process descriptions, allowing for quick
// lookups.
type ProcessTable map[PIDType]*Process

// NewProcess returns a Process object describing certain properties of the
// Linux process with the specified PID. In particular, the parent PID and the
// name of the process.
func NewProcess(PID PIDType) (proc *Process) {
	line, err := ioutil.ReadFile("/proc/" + strconv.Itoa(int(PID)) + "/stat")
	if err != nil {
		return nil
	}
	return newProcess(string(line))
}

// newProcess parses a process status line (as read from /proc/[PID]/status)
// into a Process object. Factoring out the parsing functionality allows unit
// testing it separately from the live process tree.
func newProcess(procstat string) (proc *Process) {
	proc = &Process{}
	// Gather the PID from the (1) pid field. Please note that the bracketed
	// numbers and field names are following man proc(5),
	// http://man7.org/linux/man-pages/man5/proc.5.html. Fields are separated
	// by spaces.
	pidmore := strings.SplitN(procstat, " ", 2)
	if len(pidmore) < 2 {
		return nil
	}
	pid, err := strconv.Atoi(pidmore[0])
	if err != nil || pid < 0 {
		return nil
	}
	proc.PID = PIDType(pid)
	// Extract the process name from the process status line. Please note that
	// the process name (2) is in parentheses. Now, process names may contain
	// parentheses themselves, so we have to look for the last ")" to
	// terminate the process name. And, of course, process names may also
	// container spaces.
	remainder := pidmore[1]
	namestart := strings.Index(remainder, "(")
	if namestart < 0 {
		return nil
	}
	nameend := strings.LastIndex(remainder, ")")
	if nameend < 0 {
		return nil
	}
	proc.Name = remainder[namestart+1 : nameend]
	// Now split the remainder of the process status line into fields
	// separated by simple spaces. As of Linux 3.5 there are 52 fields in
	// total (according to "man proc"), but we've alread chopped off the first
	// two ones. However, as we're only interested in the fields of up to
	// (22), we're getting sloppy and don't care about what happens after
	// field (22).
	if nameend+2 > len(remainder) {
		return nil
	}
	fields := strings.Split(remainder[nameend+2:], " ")
	if len(fields) < 22-2 {
		return nil
	}
	// Extract the Parent PID. Please note that we've chopped off two fields,
	// and array indices start at 0: so the index is 3 less than the field
	// number.
	ppid, err := strconv.Atoi(fields[4-3])
	if err != nil || ppid < 0 {
		return nil
	}
	proc.PPID = PIDType(ppid)
	// The (22) starttime filed is the start time of this process since the
	// Kernel boot epoch.
	st, err := strconv.ParseInt(fields[22-3], 10, 64)
	if err != nil || st < 0 {
		return nil
	}
	proc.Starttime = uint64(st)
	return
}

// Valid checks for the same process to still be present in the OS process
// table and then returns true, otherwise false. The validity check bases on
// the start time of the process, so stale PIDs can be detected even if they
// get reused after some time.
func (p *Process) Valid() bool {
	digitaltwin := NewProcess(p.PID)
	return digitaltwin != nil && p.Starttime == digitaltwin.Starttime
}

// String praises a Process object with a text hymn.
func (p *Process) String() string {
	return fmt.Sprintf("process PID %d %q, PPID %d",
		p.PID, p.Name, p.PPID)
}

// NewProcessTable takes returns the currently available processes (as usual,
// without tasks=threads). The process table is in fact a map, indexed by
// PIDs.
func NewProcessTable() (pt ProcessTable) {
	procentries, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil
	}
	// Phase I: discover all processes, together with some of their
	// properties, such as name and PPID.
	pt = map[PIDType]*Process{}
	for _, procentry := range procentries {
		pid, err := strconv.Atoi(procentry.Name())
		if err == nil && pid > 0 {
			proc := NewProcess(PIDType(pid))
			if proc != nil {
				pt[proc.PID] = proc
			}
		}
	}
	// Phase II: form a process object tree to speed up repeated traversals,
	// which we'll need to run during namespace discovery. This is a simple
	// optimization, just cutting map lookups at the expense of typed
	// pointers. We're even so lazy as to not check for the PPID being
	// present, as we'll get back a zero value anyway.
	for _, proc := range pt {
		proc.Parent = pt[proc.PPID]
	}
	// Phew: done.
	return
}
