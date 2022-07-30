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

//go:build linux
// +build linux

package model

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/plural"
)

// PIDType expresses things more clearly. And no, that's not a "PidType" since
// "PID" is an acronym (https://en.wikipedia.org/wiki/Acronym), but neither an
// abbreviation, nor an ordinary word (yet/still) in itself.
type PIDType int32

// Process represents our very limited view and even more limited interest in
// a specific Linux process. Well, the limitation comes from what we need for
// namespace discovery to be useful.
type Process struct {
	PID        PIDType       `json:"pid"`       // this process' identifier.
	PPID       PIDType       `json:"ppid"`      // parent's process identifier.
	Parent     *Process      `json:"-"`         // our parent's process description.
	Children   []*Process    `json:"-"`         // child processes.
	Name       string        `json:"name"`      // synthesized name of process.
	Cmdline    []string      `json:"cmdline"`   // command line of process.
	Namespaces NamespacesSet `json:"-"`         // the 7 namespaces joined by this process.
	Starttime  uint64        `json:"starttime"` // time of process start, since the Kernel boot epoch.
	CpuCgroup  string        `json:"cpucgroup"` // (relative) path of CPU control group for this process.
	// (relative) path of freezer control group for this process. Please note
	// that for a cgroup v2 unified and non-hybrid hierarchy this path will
	// always be the same as for CpuCgroup.
	FridgeCgroup string `json:"fridgecgroup"`
	FridgeFrozen bool   `json:"fridgefrozen"` // effective freezer state.

	Container *Container `json:"-"` // associated container.
}

// ProcessTable maps PIDs to their Process descriptions, allowing for quick
// lookups.
type ProcessTable map[PIDType]*Process

// Basename returns the process executable name with the directory stripped
// off, similar to what basename(1) does when applied to the "$0" argument. In
// case the basename would be empty, then the process name is returned instead
// as fallback.
func (p *Process) Basename() (basename string) {
	if len(p.Cmdline) > 0 {
		if idx := strings.LastIndex(p.Cmdline[0], "/"); idx >= 0 {
			basename = p.Cmdline[0][idx+1:]
		} else {
			basename = p.Cmdline[0]
		}
	}
	// Fall back to the process name if the command line did play tricks and
	// didn't gave us a useable name.
	if basename == "" {
		basename = p.Name
	}
	// Really fall back if even trying the process name plays tricks on us. We
	// then use a synthesized name in the form of "process (PID)".
	if basename == "" {
		basename = "process (" + strconv.FormatUint(uint64(p.PID), 10) + ")"
	}
	return
}

// NewProcess returns a Process object describing certain properties of the
// Linux process with the specified PID. In particular, the parent PID and the
// name of the process, as well as the command line.
func NewProcess(PID PIDType) (proc *Process) {
	return NewProcessInProcfs(PID, "/proc")
}

// NewProcessInProcfs implements NewProcess and additionally allows for testing on
// fake /proc "filesystems".
func NewProcessInProcfs(PID PIDType, procroot string) (proc *Process) {
	procbase := procroot + "/" + strconv.Itoa(int(PID))
	line, err := os.ReadFile(procbase + "/stat") // #nosec G304
	if err != nil {
		return nil
	}
	proc = newProcessFromStatline(string(line))
	if proc == nil {
		return
	}
	// Also get the process command line, so later tools can decide to
	// either go for the process name or the executable basename, et
	// cetera.
	cmdline, err := os.ReadFile(procbase + "/cmdline") // #nosec G304
	if err == nil {
		cmdparts := bytes.Split(bytes.TrimRight(cmdline, "\x00"), []byte{0x00})
		proc.Cmdline = make([]string, len(cmdparts))
		for idx, part := range cmdparts {
			proc.Cmdline[idx] = string(part)
		}
	}
	return proc
}

// newProcessStat parses a process status line (as read from /proc/[PID]/status)
// into a Process object. Factoring out the parsing functionality allows unit
// testing it separately from the live process tree.
func newProcessFromStatline(procstat string) (proc *Process) {
	proc = &Process{}
	// Gather the PID from the (1) pid field. Please note that the bracketed
	// numbers and field names are following man proc(5),
	// http://man7.org/linux/man-pages/man5/proc.5.html. Fields are separated
	// by spaces.
	pidmore := strings.SplitN(procstat, " ", 2)
	if len(pidmore) < 2 {
		return nil
	}
	pid, err := strconv.ParseInt(pidmore[0], 10, 32)
	if err != nil || pid <= 0 {
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
	// Extract the Parent PID (field 4). Please note that we've chopped off
	// two fields, and array indices start at 0: so the index is 3 less than
	// the field number.
	ppid, err := strconv.ParseUint(fields[4-3], 10, 32)
	if err != nil {
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
	if p == nil {
		return "Process <nil>"
	}
	return fmt.Sprintf("process PID %d %q, PPID %d", p.PID, p.Name, p.PPID)
}

// NewProcessTable takes returns the currently available processes (as usual,
// without tasks/threads). The process table is in fact a map, indexed by PIDs.
// When the freezer parameter is true then additionally the cgroup freezer
// states will also be discovered; as this might require switching into the
// initial mount namespace and this is possible in Go only when re-executing as
// a child, the caller must explicitly request this additional discovery.
func NewProcessTable(freezer bool) (pt ProcessTable) {
	pt = NewProcessTableFromProcfs(freezer, "/proc")
	log.Infof("discovered %s", plural.Elements(len(pt), "processes"))
	return
}

// NewProcessTableFromProcfs implements NewProcessTable and allows for testing on fake
// /proc "filesystems".
func NewProcessTableFromProcfs(freezer bool, procroot string) (pt ProcessTable) {
	procentries, err := os.ReadDir(procroot)
	if err != nil {
		return nil
	}
	// Phase I: discover all processes, together with some of their
	// properties, such as name and PPID.
	pt = map[PIDType]*Process{}
	for _, procentry := range procentries {
		// Get the process PID as a number and then read its /proc/[PID]/stat
		// procfs entry in order to get some details about the process.
		pid, err := strconv.ParseInt(procentry.Name(), 10, 32)
		if err != nil || pid <= 0 {
			continue
		}
		proc := NewProcessInProcfs(PIDType(pid), procroot)
		if proc == nil {
			continue
		}
		pt[proc.PID] = proc
	}
	// Phase II: form a process object tree to speed up repeated traversals,
	// which we'll need to run during namespace discovery. This is a simple
	// optimization, just cutting map lookups at the expense of typed
	// pointers. We're even so lazy as to not check for the PPID being
	// present, as we'll get back a zero value anyway.
	for _, proc := range pt {
		if parent, ok := pt[proc.PPID]; ok {
			proc.Parent = parent
			parent.Children = append(parent.Children, proc)
		}
	}
	// Scan for the control groups of the processes in the table. We always
	// scan, as we can do so from our current mount and cgroup namespaces.
	// However, in order to correctly discover the cgroup paths for all
	// processes we need to run this while inside the initial cgroup namespace.
	pt.scanCgroups()
	// If requested, additionally scan for the freezer states; this is a more
	// expensive operation in case we need to switch into the initial mount
	// namespace, as otherwise we might not see the full cgroups freezer
	// hierarchy.
	if freezer {
		pt.scanFridges()
	}
	// Phew: done.
	return
}

// ByName returns all processes with the specified name.
func (t ProcessTable) ByName(name string) (procs []*Process) {
	for _, proc := range t {
		if proc.Name == name {
			procs = append(procs, proc)
		}
	}
	return
}

// ProcessesByPIDs returns the Process objects corresponding to the specified
// PIDs. It skips PIDs for which no Process object is known and only returns
// Process objects for known PIDs. If you need error handling, then you'll
// better roll your own function.
func (t ProcessTable) ProcessesByPIDs(pid ...PIDType) []*Process {
	procs := make([]*Process, 0, len(pid))
	for _, p := range pid {
		proc, ok := t[p]
		if !ok {
			continue
		}
		procs = append(procs, proc)
	}
	return procs
}

// ProcessListByPID is a type alias for sorting slices of *Process by their
// PIDs in numerically ascending order.
type ProcessListByPID []*Process

func (l ProcessListByPID) Len() int      { return len(l) }
func (l ProcessListByPID) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ProcessListByPID) Less(i, j int) bool {
	return l[i].PID < l[j].PID
}
