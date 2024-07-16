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

package model

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/plural"
	"golang.org/x/sys/unix"
)

// PIDType expresses things more clearly.
//
//   - No, that's not a "PidType" since “PID” is an [acronym],
//     but neither an abbreviation, nor an ordinary word (yet/still) in itself.
//
// [acronym]: https://en.wikipedia.org/wiki/Acronym
type PIDType int32

// ProTaskCommon defines the fields we're interested in that are common to both
// [Process] and [Task] objects.
type ProTaskCommon struct {
	Name       string        `json:"name"`      // (limited) name, from the comm status field.
	Namespaces NamespacesSet `json:"-"`         // the 8 namespaces joined by this process.
	Starttime  uint64        `json:"starttime"` // time of process start, since the Kernel boot epoch.
	CpuCgroup  string        `json:"cpucgroup"` // (relative) path of CPU control group for this process.
	// (relative) path of freezer control group for this process. Please note
	// that for a cgroup v2 unified and non-hybrid hierarchy this path will
	// always be the same as for CpuCgroup.
	FridgeCgroup string `json:"fridgecgroup"`
	FridgeFrozen bool   `json:"fridgefrozen"` // effective freezer state.
	// CPU ranges affinity list, need explicit request via
	// ProTaskCommon.GetAffinity.
	Affinity CPUList `json:"affinity,omitempty"`
	Policy   int     `json:"policy,omitempty"`
	// priority value is considered by the following schedulers:
	//   - SCHED_FIFO: prio 1..99.
	//   - SCHED_RR: prio 1..99.
	//   - SCHED_NORMAL (=SCHED_OTHER): not used/prio is 0.
	//   - SCHED_IDLE: not used/prio is 0.
	//   - SCHED_BATCH: not used/prio is 0.
	//   - SCHED_DEADLINE: doesn't use prio.
	Priority int `json:"priority,omitempty"`
	// nice value in the range +19..-20 (very nice ... less nice) is considered
	// by the following schedulers:
	//   - SCHED_NORMAL (=SCHED_OTHER): nice is taken into account.
	//   - SCHED_BATCH: nice is taken into account.
	//   - SCHED_IDLE: nice is ignored (basically below a nic of +19).
	Nice int `json:"nice,omitempty"`
}

// Task represents our very, very limited view and interest in a particular
// Linux task (including the main task that represents the whole process).
type Task struct {
	ProTaskCommon
	TID     PIDType  `json:"tid"` // our task's identifier.
	Process *Process `json:"-"`   // our main task ~process.
}

// Process represents our very limited view and even more limited interest in
// a specific Linux process. Well, the limitation comes from what we need for
// namespace discovery to be useful.
type Process struct {
	ProTaskCommon
	PID       PIDType    `json:"pid"`             // this process' identifier.
	PPID      PIDType    `json:"ppid"`            // parent's process identifier.
	Parent    *Process   `json:"-"`               // our parent's process description.
	Children  []*Process `json:"-"`               // child processes.
	Cmdline   []string   `json:"cmdline"`         // command line of process.
	Tasks     []*Task    `json:"tasks,omitempty"` // tasks of this process, including the main task.
	Container *Container `json:"-"`               // associated container; only for the leader.
}

// ProcessTable maps PIDs to their [model.Process] descriptions, allowing for
// quick lookups.
type ProcessTable map[PIDType]*Process

// Process/task status field indices for a split /proc/$PID/stat line, with
// zero-based indices. See also
// https://man7.org/linux/man-pages/man5/proc.5.html.
const (
	statlineFieldPID       = 1 - 1 //nolint:staticcheck // SA4000 stupid, stupid lill' linter.
	statlineFieldComm      = 2 - 1
	statlineFieldPPID      = 4 - 1
	statlineFieldStarttime = 22 - 1
)

// Basename returns the process executable name with the directory stripped off,
// similar to what [basename(1)] does when applied to the “$0” argument.
// However, in case the basename would be empty, then the process name is
// returned instead as fallback.
//
// [basename(1)]: https://man7.org/linux/man-pages/man1/basename.1.html
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

// NewProcess returns a [model.Process] object describing certain properties of
// the Linux process with the specified PID. In particular, the parent PID and
// the name of the process, as well as the command line. If withtasks is true,
// it will additionally discover all tasks of the process.
func NewProcess(PID PIDType, withtasks bool) (proc *Process) {
	return NewProcessInProcfs(PID, withtasks, "/proc")
}

// NewProcessInProcfs implements [model.NewProcess] and additionally allows for
// testing on fake /proc "filesystems".
func NewProcessInProcfs(PID PIDType, withtasks bool, procroot string) (proc *Process) {
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
	if err == nil {                                    // be forgiving
		cmdparts := bytes.Split(bytes.TrimRight(cmdline, "\x00"), []byte{0x00})
		proc.Cmdline = make([]string, len(cmdparts))
		for idx, part := range cmdparts {
			proc.Cmdline[idx] = string(part)
		}
	}
	if !withtasks {
		return proc
	}
	proc.discoverTasks(procbase)
	return proc
}

// discoverTasks discovers the tasks of this particular process in the process
// filesystem pointed to by procbase.
func (p *Process) discoverTasks(procbase string) {
	procbase += "/task"
	taskentries, err := os.ReadDir(procbase)
	if err != nil {
		return
	}
	for _, taskentry := range taskentries {
		// Get the task TID as a number and then read its
		// /proc/[PID]/task/[TASK]/stat procfs entry in order to get some
		// details about the task.
		tid, err := strconv.ParseInt(taskentry.Name(), 10, 32)
		if err != nil || tid <= 0 {
			continue
		}
		line, err := os.ReadFile(procbase + "/" + taskentry.Name() + "/stat") // #nosec G304
		if err != nil {
			continue
		}
		task := newTaskFromStatline(string(line), p)
		if task == nil {
			continue
		}
		p.Tasks = append(p.Tasks, task)
	}
}

// newProcessStat parses a process status line (as read from /proc/[PID]/status)
// into a Process object. Factoring out the parsing functionality allows unit
// testing it separately from the live process tree.
func newProcessFromStatline(procstat string) (proc *Process) {
	pid, starttime, statFields := commonFromStatline(procstat)
	if statFields == nil {
		return nil
	}
	proc = &Process{
		PID: pid,
		ProTaskCommon: ProTaskCommon{
			Name:      statFields[statlineFieldComm],
			Starttime: starttime,
		},
	}
	// PIDs are unsigned, but passed as int32...
	ppid, err := strconv.ParseUint(statFields[statlineFieldPPID], 10, 31)
	if err != nil {
		return nil
	}
	proc.PPID = PIDType(ppid)
	return
}

// commonFromStatline returns the PID, starttime and finally all stat line
// fields, or nil for the fields if the stat line turns out to be invalid.
func commonFromStatline(statline string) (PIDType, uint64, []string) {
	statFields := splitStatline(statline)
	if statFields == nil {
		return 0, 0, nil
	}
	pid, err := strconv.ParseInt(statFields[statlineFieldPID], 10, 32)
	if err != nil || pid <= 0 {
		return 0, 0, nil
	}
	starttime, err := strconv.ParseInt(statFields[statlineFieldStarttime], 10, 64)
	if err != nil || starttime < 0 {
		return 0, 0, nil
	}
	return PIDType(pid), uint64(starttime), statFields
}

// splitStatline returns the individual fields of a /proc/$PID/stat line,
// properly handling the name field with its special round bracket escaping.
// splitStatline returns nil if the supplied stat line is malformed, such as not
// offering fields up to the start time field. It does not check for correct
// individual field values though.
func splitStatline(statline string) []string {
	// As the second field may contain spaces, it is always bracket in itself.
	// So we first split of the first field, the PID (or TID) and then deal with
	// the second field separately, before all the remaining fields again are
	// simply separated by spaces.
	firstAndMore := strings.SplitN(statline, " ", 2)
	if len(firstAndMore) < 2 {
		return nil
	}
	// Now isolate the process name from the process status line. Please note
	// that the process name (2) is in parentheses. Now, process names may
	// contain parentheses themselves, so we have to look for the last ")" to
	// terminate the process name. And, of course, process names may also
	// container spaces.
	nameAndMore := firstAndMore[1]
	nameStart := strings.Index(nameAndMore, "(")
	if nameStart < 0 {
		return nil
	}
	nameEnd := strings.LastIndex(nameAndMore, ")")
	if nameEnd < 0 {
		return nil
	}
	name := nameAndMore[nameStart+1 : nameEnd]
	if nameEnd+2 > len(nameAndMore) {
		return nil
	}
	fields := strings.Split(nameAndMore[nameEnd+2:], " ")
	if len(fields) < 22-2 {
		return nil
	}
	statfields := []string{
		firstAndMore[0],
		name,
	}
	return append(statfields, fields...)
}

// Valid checks for the same process to still be present in the OS process
// table and then returns true, otherwise false. The validity check bases on
// the start time of the process, so stale PIDs can be detected even if they
// get reused after some time.
func (p *Process) Valid() bool {
	digitaltwin := NewProcess(p.PID, false)
	return digitaltwin != nil && p.Starttime == digitaltwin.Starttime
}

// String praises a Process object with a text hymn.
func (p *Process) String() string {
	if p == nil {
		return "Process <nil>"
	}
	return fmt.Sprintf("process PID %d %q, PPID %d", p.PID, p.Name, p.PPID)
}

// NewProcessTable returns the currently available processes (as usual, without
// tasks/threads). The process table is in fact a map, indexed by PIDs. When the
// freezer parameter is true then additionally the cgroup freezer states will
// also be discovered; as this might require switching into the initial mount
// namespace and this is possible in Go only when re-executing as a child, the
// caller must explicitly request this additional discovery.
func NewProcessTable(freezer bool) (pt ProcessTable) {
	pt = NewProcessTableFromProcfs(freezer, false, "/proc")
	log.Infof("discovered %s", plural.Elements(len(pt), "processes"))
	return
}

// NewProcessTableWithTasks returns not only the currently available tasks, but
// optionally also the tasks/threads.
func NewProcessTableWithTasks(freezer bool) (pt ProcessTable) {
	pt = NewProcessTableFromProcfs(freezer, true, "/proc")
	log.Infof("discovered %s", plural.Elements(len(pt), "processes"))
	return
}

// NewProcessTableFromProcfs implements [model.NewProcessTable] and allows for
// testing on fake /proc "filesystems".
func NewProcessTableFromProcfs(freezer bool, withtasks bool, procroot string) (pt ProcessTable) {
	procentries, err := os.ReadDir(procroot)
	if err != nil {
		return nil
	}
	// Phase I: discover all processes, together with some of their
	// properties, such as name and PPID.
	pt = map[PIDType]*Process{}
	for _, procentry := range procentries {
		// Get the process PID as a number and then read its /proc/[PID]/stat
		// procfs entry in order to get some details about the process. Skip
		// entries that do not represent PIDs.
		pid, err := strconv.ParseInt(procentry.Name(), 10, 32)
		if err != nil || pid <= 0 {
			continue
		}
		proc := NewProcessInProcfs(PIDType(pid), withtasks, procroot)
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

// ProcessesByPIDs returns the [model.Process] objects corresponding to the
// specified PIDs. It skips PIDs for which no Process object is known and only
// returns Process objects for known PIDs. If you need error handling, then
// you'll better roll your own function.
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

// newTaskFromStatline parses a task (process) status line (as read from
// /proc/[PID]/task/[TID]/status) into a Task object.
func newTaskFromStatline(procstat string, proc *Process) (task *Task) {
	tid, starttime, statFields := commonFromStatline(procstat)
	if statFields == nil {
		return nil
	}
	task = &Task{
		TID:     tid,
		Process: proc,
		ProTaskCommon: ProTaskCommon{
			Name:      statFields[statlineFieldComm],
			Starttime: starttime,
		},
	}
	return
}

// MainTask returns true if the given Task is the process main task.
func (t *Task) MainTask() bool {
	return t.TID == t.Process.PID
}

func (c *ProTaskCommon) retrieveAffinityScheduling(pid PIDType) error {
	var err error
	c.Affinity, err = GetCPUList(pid)
	if err != nil {
		return err
	}
	schedattr, err := unix.SchedGetAttr(int(pid), 0)
	if err != nil {
		return err
	}
	c.Policy = int(schedattr.Policy)
	c.Nice = int(schedattr.Nice)
	c.Priority = int(schedattr.Priority)
	return nil
}

// RetrieveAffinity updates this Process object's Affinity CPU range list and
// scheduling information (policy, priority, ...), returning nil when
// successful. Otherweise, it returns an error.
func (p *Process) RetrieveAffinityScheduling() error {
	return p.retrieveAffinityScheduling(p.PID)
}

// RetrieveAffinity updates this Task object's Affinity CPU range list and
// scheduling information (policy, priority, ...), returning nil when
// successful. Otherweise, it returns an error.
func (t *Task) RetrieveAffinityScheduling() error {
	return t.retrieveAffinityScheduling(t.TID)
}
