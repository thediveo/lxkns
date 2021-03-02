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

package model

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/lxkns/log"
)

// PIDType expresses things more clearly. And no, that's not a "PidType" since
// "PID" is an acronym (https://en.wikipedia.org/wiki/Acronym), but neither an
// abbreviation, nor an ordinary word (yet/still) in itself.
type PIDType int32

// ProcessFridgeStatus represents the the freezer state of a process. For
// cgroups v1 this represents the freezer cgroup state. For cgroups v2 in a
// unified hierarchy this represents the unified cgroup freezer status. See
// also:
// https://www.kernel.org/doc/Documentation/admin-guide/cgroup-v1/freezer-subsystem.rst
// and
// https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#core-interface-files.
type ProcessFridgeStatus uint8

const (
	ProcessThawed ProcessFridgeStatus = iota
	ProcessFreezing
	ProcessFrozen
)

// String returns the textual representation of the "fridge" status of a
// process.
func (s ProcessFridgeStatus) String() string {
	switch s {
	case ProcessThawed:
		return "thawed"
	case ProcessFreezing:
		return "freezing"
	case ProcessFrozen:
		return "frozen"
	default:
		return fmt.Sprintf("ProcessFridgeStatus(%d)", s)
	}
}

// MarshalJSON marshals a process fridge status as a JSON string with the fixed
// enum values "thawed", "freezing", and "frozen".
func (s ProcessFridgeStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals either a JSON string or a number into a process
// fridge status enum value.
func (s *ProcessFridgeStatus) UnmarshalJSON(b []byte) error {
	var fridgestatus interface{}
	if err := json.Unmarshal(b, &fridgestatus); err == nil {
		switch fridgestatus := fridgestatus.(type) {
		case float64:
			*s = ProcessFridgeStatus(fridgestatus)
			return nil
		case string:
			switch fridgestatus {
			case "thawed":
				*s = ProcessThawed
			case "freezing":
				*s = ProcessFreezing
			case "frozen":
				*s = ProcessFrozen
			default:
				return fmt.Errorf("invalid ProcessFridgeStatus %q", fridgestatus)
			}
			return nil
		}
	}
	return fmt.Errorf("cannot convert %q to ProcessFridgeStatus", b)
}

// Process represents our very limited view and even more limited interest in
// a specific Linux process. Well, the limitation comes from what we need for
// namespace discovery to be useful.
type Process struct {
	PID          PIDType       `json:"pid"`       // this process' identifier.
	PPID         PIDType       `json:"ppid"`      // parent's process identifier.
	Parent       *Process      `json:"-"`         // our parent's process description.
	Children     []*Process    `json:"-"`         // child processes.
	Name         string        `json:"name"`      // synthesized name of process.
	Cmdline      []string      `json:"cmdline"`   // command line of process.
	Namespaces   NamespacesSet `json:"-"`         // the 7 namespaces joined by this process.
	Starttime    uint64        `json:"starttime"` // time of process start, since the Kernel boot epoch.
	Controlgroup string        `json:"cgroup"`    // (relative) path of CPU control group for this process.
	// (relative) path of freezer control group for this process. Please note
	// that for a cgroup v2 unified and non-hybrid hierarchy this path will
	// always be the same as for Controlgroup.
	FridgeCgroup string              `json:"fridgecgroup"`
	Fridge       ProcessFridgeStatus `json:"fridge"`       // effective freezer state.
	Selffridge   ProcessFridgeStatus `json:"selffridge"`   // own freezer state as last set; only ProcessThawned or ProcessFrozen.
	Parentfridge ProcessFridgeStatus `json:"parentfridge"` // parent freezer state as last set; only ProcessThawned or ProcessFrozen.
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
	return newProcess(PID, "/proc")
}

// newProcess implements NewProcess and additionally allows for testing on
// fake /proc "filesystems".
func newProcess(PID PIDType, procroot string) (proc *Process) {
	procbase := procroot + "/" + strconv.Itoa(int(PID))
	line, err := ioutil.ReadFile(procbase + "/stat")
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
	cmdline, err := ioutil.ReadFile(procbase + "/cmdline")
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
	// Extract the Parent PID (field 4). Please note that we've chopped off
	// two fields, and array indices start at 0: so the index is 3 less than
	// the field number.
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
	if p == nil {
		return "Process <nil>"
	}
	return fmt.Sprintf("process PID %d %q, PPID %d", p.PID, p.Name, p.PPID)
}

// NewProcessTable takes returns the currently available processes (as usual,
// without tasks=threads). The process table is in fact a map, indexed by
// PIDs.
func NewProcessTable() (pt ProcessTable) {
	pt = newProcessTable("/proc")
	log.Infof("discovered %d processes", len(pt))
	return
}

// newProcessTable implements NewProcessTable and allows for testing on fake
// /proc "filesystems".
func newProcessTable(procroot string) (pt ProcessTable) {
	procentries, err := ioutil.ReadDir(procroot)
	if err != nil {
		return nil
	}
	// Phase I: discover all processes, together with some of their
	// properties, such as name and PPID.
	pt = map[PIDType]*Process{}
	for _, procentry := range procentries {
		// Get the process PID as a number and then read its /proc/[PID]/stat
		// procfs entry in order to get some details about the process.
		pid, err := strconv.Atoi(procentry.Name())
		if err != nil || pid == 0 {
			continue
		}
		proc := newProcess(PIDType(pid), procroot)
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
	// Scan for the control groups of the processes in the table.
	_ = pt.scanCgroups()
	// Phew: done.
	return
}

// ProcessListByPID is a type alias for sorting slices of *Process by their
// PIDs in numerically ascending order.
type ProcessListByPID []*Process

func (l ProcessListByPID) Len() int      { return len(l) }
func (l ProcessListByPID) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ProcessListByPID) Less(i, j int) bool {
	return l[i].PID < l[j].PID
}

// scanCgroups scans all processes for their control groups; it scans only on a
// specific type of controller, the "cpu" v1 controller on (1) the assumption
// that this controller is widely used and (2) we're interested in the fridge
// (well, "freezer") state. On a side note, the "memory" controller
// unfortunately has been disabled on some architectures (ARM) for some time.
func (p ProcessTable) scanCgroups() error {
	// Try to find the freezer cgroup v1 hierarchy root, if available...
	fridgeroot := ""
	fridgev1 := true
Fridge:
	for _, mountinfo := range mntinfo.MountsOfType(-1, "cgroup") {
		for _, sopt := range strings.Split(mountinfo.SuperOptions, ",") {
			if sopt == "freezer" {
				fridgeroot = mountinfo.MountPoint
				break Fridge
			}
		}
	}
	// ...otherwise, there must be a cgroups v2 unified hierarchy.
	if fridgeroot == "" {
		mountinfo := mntinfo.MountsOfType(-1, "cgroup2")
		if len(mountinfo) > 0 {
			fridgev1 = false
			fridgeroot = mountinfo[0].MountPoint
		}
	}
	// Finally scan the processes for their cgroups...
	for pid, proc := range p {
		controllers := processCgroup(cgrouptypes, pid)
		proc.Controlgroup = controllers[0]
		proc.FridgeCgroup = controllers[1]
		if fridgev1 {
			freezerV1(proc, filepath.Join(fridgeroot, controllers[1]))
		} else {
			freezerV2(proc, filepath.Join(fridgeroot, controllers[1]))
		}
	}
	return nil
}

var cgrouptypes = []string{"cpu", "freezer"}

// freezerV2 reads and stores the cgroups v2 freezer status information for the
// specified process.
func freezerV2(proc *Process, fridgepath string) {
	// Please note that the v2 root cgroup doesn't have the "cgroup.freeze"
	// interface file. Other than that, see
	// https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#core-interface-files
	// for details.
	//
	// Now, the v2 freezer uses "cgroup.freeze" to indicate the desired self
	// state, thus replacing v1's "freezer.self_freezing".
	if state, err := ioutil.ReadFile(
		filepath.Join(fridgepath, "cgroup.freeze")); err == nil {
		switch strings.TrimSuffix(string(state), "\n") {
		case "1":
			proc.Selffridge = ProcessFrozen
		default:
			proc.Selffridge = ProcessThawed
		}
	}
	// But, where's v1's "freezer.state", that is, the process' effective state?
	// It can now be found as one of possibly many event entries in
	// "cgroup.events". Yuk.
	freezerState := ProcessThawed
	if events, err := ioutil.ReadFile(
		filepath.Join(fridgepath, "cgroup.events")); err == nil {
		for _, event := range strings.Split(string(events), "\n") {
			if strings.HasPrefix(event, "frozen ") {
				if event[7] == '1' {
					freezerState = ProcessFrozen
				}
				break
			}
		}
	}
	// Emulate the effective "freezer.state" from v1: take the effective state
	// from the events interface file, but assume the "freezing" state iff this
	// process should be frozen, yet the events yet don't indicate so.
	if proc.Selffridge == ProcessFrozen && freezerState != ProcessFrozen {
		proc.Fridge = ProcessFreezing
	} else {
		proc.Fridge = freezerState
	}
	// TODO: parent_freezing
}

// freezerV1 reads and stores the cgroups v1 freezer status information for the
// specified process.
func freezerV1(proc *Process, fridgepath string) {
	// Please note: "the root cgroup is non-freezable and the above
	// interface files don't exist."
	// (https://www.kernel.org/doc/Documentation/admin-guide/cgroup-v1/freezer-subsystem.rst)
	if state, err := ioutil.ReadFile(
		filepath.Join(fridgepath, "freezer.state")); err == nil {
		switch strings.TrimSuffix(string(state), "\n") {
		case "FREEZING":
			proc.Fridge = ProcessFreezing
		case "FROZEN":
			proc.Fridge = ProcessFrozen
		case "THAWED":
			fallthrough
		default:
			proc.Fridge = ProcessThawed
		}
	} else {
		proc.Fridge = ProcessThawed
	}
	if selfstate, err := ioutil.ReadFile(
		filepath.Join(fridgepath, "freezer.self_freezing")); err == nil {
		switch strings.TrimSuffix(string(selfstate), "\n") {
		case "1":
			proc.Selffridge = ProcessFrozen
		default:
			proc.Selffridge = ProcessThawed
		}
	} else {
		proc.Selffridge = ProcessThawed
	}
	if parentstate, err := ioutil.ReadFile(
		filepath.Join(fridgepath, "freezer.parent_freezing")); err == nil {
		switch strings.TrimSuffix(string(parentstate), "\n") {
		case "1":
			proc.Parentfridge = ProcessFrozen
		default:
			proc.Parentfridge = ProcessThawed
		}
	} else {
		proc.Parentfridge = ProcessThawed
	}
}

// processCgroup returns the name (hierarchy path) of some of the cgroup
// controllers a specific process is in (as specified in the controllertypes
// parameter).
//
// We first try to find the specified cgroup v1 controllers if available and
// only then fall back to the unified cgroups v2 hierarchy.
//
// Note: the cgroup path(s) returned is (are) relative to their respective
// process cgroup roots (as can be found by inspecting mountinfo), even as they
// start with "/" (at least when discovered inside the initial pid+cgroup
// namespaces).
func processCgroup(controllertypes []string, pid PIDType) (paths []string) {
	paths = make([]string, len(controllertypes))
	cgroup, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return
	}
	defer cgroup.Close()
	scanner := bufio.NewScanner(cgroup)
	unifiedroot := "" // (if detected) the cgroups v2 unified hierarchy root
	for scanner.Scan() {
		if err == nil {
			// See https://man7.org/linux/man-pages/man7/cgroups.7.html, section
			// "NOTES", subsection "/proc files". For cgroups v1 controllers,
			// the second field specifies the comma-separated list of the
			// controllers bound to the hierarchy: here, we look for, say, the
			// "cpu" controller. The third field specifies the path in the
			// cgroups hierarchy; it is relative to the mount point of the
			// hierarchy -- which in turn depends on the mount namespace of this
			// process :)
			//
			// For the unified cgroups v2 hierarchy the second field will be
			// empty, which otherwise would specify the particular cgroup v1
			// hierarchy/-ies.
			if fields := strings.Split(scanner.Text(), ":"); len(fields) == 3 {
				if fields[1] != "" {
					// cgroups v1 hierarchies
					controllers := strings.Split(fields[1], ",")
					for _, ctrl := range controllers {
						for idx, controllertype := range controllertypes {
							if ctrl == controllertype {
								paths[idx] = fields[2]
							}
						}
					}
				} else {
					// when we come across a single unified cgroups v2 hierarchy
					// root, remember it so we can later fix any missing
					// controller paths.
					unifiedroot = fields[2]
				}
			}
		}
	}
	// Now fix the missing cgroups v2 controller paths we couldn't satisfy from
	// v1 (if present).
	for idx, path := range paths {
		if path == "" {
			paths[idx] = unifiedroot
		}
	}
	// Hopefully, we've gathered all controller paths by now.
	return
}
