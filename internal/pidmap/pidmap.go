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

package pidmap

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// PIDMap implements PIDMapper for translating PIDs between PID namespaces.
type PIDMap map[model.NamespacedPID]model.NamespacedPIDs

// Ensure that PIDMap implements PIDMapper.
var _ (model.PIDMapper) = (*PIDMap)(nil)

// Translate translates a PID "pid" in PID namespace "from" to the
// corresponding PID in PID namespace "to". Returns 0, if PID "pid" either
// does not exist in namespace "from", or PID namespace "to" isn't either a
// parent or child of PID namespace "from".
func (pidmap PIDMap) Translate(pid model.PIDType, from model.Namespace, to model.Namespace) model.PIDType {
	if namespacedpids, ok := pidmap[model.NamespacedPID{PID: pid, PIDNS: from}]; ok {
		for _, namespacedpid := range namespacedpids {
			if namespacedpid.PIDNS == to {
				return namespacedpid.PID
			}
		}
	}
	return 0
}

// NamespacedPIDs returns for a specific namespaced PID the list of all PIDs
// the corresponding process has been given in different PID namespaces.
// Returns nil if the PID doesn't exist in the specified PID namespace. The
// list is ordered from the topmost PID namespace down to the leaf PID
// namespace to which a process actually is joined to.
func (pidmap PIDMap) NamespacedPIDs(pid model.PIDType, from model.Namespace) model.NamespacedPIDs {
	if namespacedpids, ok := pidmap[model.NamespacedPID{PID: pid, PIDNS: from}]; ok {
		size := len(namespacedpids)
		nspids := make([]model.NamespacedPID, size)
		for idx, el := range namespacedpids {
			nspids[size-1-idx] = el
		}
		return nspids
	}
	return nil
}

// NewPIDMap returns a new PID map based on the specified PID table
// and further information gathered from the /proc filesystem.
func NewPIDMap(processes model.ProcessTable) PIDMap {
	pidmap := PIDMap{}
	for _, proc := range processes {
		// For each process, first get its list of namespaced PIDs, which
		// lists the PIDs starting from the PID namespace we're currently in
		// and continues into nested child PID namespaces.
		pids := NSpid(proc)
		pidns := proc.Namespaces[model.PIDNS]
		if pidns == nil {
			continue
		}
		pidhns := pidns.(model.Hierarchy)
		// The namespaced PIDs are top-down, while we have to go bottom-up
		// from the process' current PID namespace, in order to assemble the
		// list of NamespacedPIDs correctly.
		pidslen := len(pids)
		if pidslen == 0 {
			continue
		}
		namespacedpids := make(model.NamespacedPIDs, pidslen)
		idx := 0
		for pidhns != nil {
			namespacedpids[idx] = model.NamespacedPID{
				PIDNS: pidhns.(model.Namespace),
				PID:   pids[pidslen-idx-1],
			}
			pidhns = pidhns.Parent()
			idx++
		}
		if idx != pidslen {
			// Did someone forgot to also discover the hierarchy???
			continue
		}
		// Now index these NamespacedPIDs by the list NamespacedPID elements,
		// so we can later quickly look up the list of namespaced PIDs for
		// this process using (PID, PID-namespace).
		for _, namespacedpid := range namespacedpids {
			pidmap[namespacedpid] = namespacedpids
		}
	}
	return pidmap
}

// NSpid returns the list of namespaced PIDs for the process proc, based on
// information from the /proc filesystem (the "NSpid:" field in particular).
// NSpid only returns the list of PIDs, but not the corresponding PID
// namespaces; this is because the Linux kernel doesn't give us the namespace
// information as part of the process status. Instead, a caller (such as
// NewPIDMap) needs to combine a namespaced PIDs list with the hierarchy of
// PID namespaces to calculate the correct namespacing.
func NSpid(proc *model.Process) (pids []model.PIDType) {
	return nspid(proc, "/proc")
}

// nspid implements NSpid, allowing for testing on fake /proc "filesystems".
func nspid(proc *model.Process, procroot string) (pids []model.PIDType) {
	f, err := os.Open(fmt.Sprintf("%s/%d/status", procroot, proc.PID))
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// Scan through the process status information until we arrive at the
	// sought-after "NSpid:" field. That's the only field interesting to us.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "NSpid:\t") {
			pidstxts := strings.Split(line[7:], "\t")
			pids = make([]model.PIDType, len(pidstxts))
			for idx, pidtxt := range pidstxts {
				pid, err := strconv.ParseInt(pidtxt, 10, 32)
				if err != nil {
					return []model.PIDType{}
				}
				pids[idx] = model.PIDType(pid)
			}
			return
		}
	}
	panic(procroot + " filesystem broken: no NSpid element in status.")
}
