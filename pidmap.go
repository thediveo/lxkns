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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// NamespacedPID is PID in the context of a specific PID namespace.
type NamespacedPID struct {
	PIDNS model.Namespace // PID namespace ID for PID.
	PID   model.PIDType   // PID within PID namespace (of ID).
}

// NamespacedPIDs is a list of PIDs for the same process, but in different PID
// namespaces. The order of the list is undefined.
type NamespacedPIDs []NamespacedPID

// PIDs just returns the different PIDs assigned to a single process in
// different PID namespaces, without the namespaces. This is a convenience
// function for those lazy cases where just the PID list is wanted, but no PID
// namespace details.
func (ns NamespacedPIDs) PIDs() []model.PIDType {
	pids := make([]model.PIDType, len(ns))
	for idx, el := range ns {
		pids[idx] = el.PID
	}
	return pids
}

// PIDMap maps a single namespaced PID to the list of PIDs for this process in
// different PID namespaces. Further PIDMap methods then allow simple
// translation of PIDs between different PID namespaces.
type PIDMap struct {
	m map[NamespacedPID]NamespacedPIDs
}

// Translate translates a PID "pid" in PID namespace "from" to the
// corresponding PID in PID namespace "to". Returns 0, if PID "pid" either
// does not exist in namespace "from", or PID namespace "to" isn't either a
// parent or child of PID namespace "from".
func (pm *PIDMap) Translate(pid model.PIDType, from model.Namespace, to model.Namespace) model.PIDType {
	if namespacedpids, ok := pm.m[NamespacedPID{PID: pid, PIDNS: from}]; ok {
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
func (pm *PIDMap) NamespacedPIDs(pid model.PIDType, from model.Namespace) NamespacedPIDs {
	if namespacedpids, ok := pm.m[NamespacedPID{PID: pid, PIDNS: from}]; ok {
		size := len(namespacedpids)
		nspids := make([]NamespacedPID, size)
		for idx, el := range namespacedpids {
			nspids[size-1-idx] = el
		}
		return nspids
	}
	return nil
}

// NewPIDMap returns a new PID map based on the specified discovery results
// and further information gathered from the /proc filesystem.
func NewPIDMap(res *DiscoveryResult) *PIDMap {
	pm := &PIDMap{
		m: map[NamespacedPID]NamespacedPIDs{},
	}
	for _, proc := range res.Processes {
		// For each process, first get its list of namespaced PIDs, which
		// lists the PIDs starting from the PID namespace we're currently in
		// and continues into nested child PID namespaces.
		pids := NSpid(proc)
		pidns := proc.Namespaces[model.PIDNS].(model.Hierarchy)
		// The namespaced PIDs are top-down, while we have to go bottom-up
		// from the process' current PID namespace, in order to assemble the
		// list of NamespacedPIDs correctly.
		pidslen := len(pids)
		if pidslen == 0 {
			continue
		}
		namespacedpids := make(NamespacedPIDs, pidslen)
		idx := 0
		for pidns != nil {
			namespacedpids[idx] = NamespacedPID{
				PIDNS: pidns.(model.Namespace),
				PID:   pids[pidslen-idx-1],
			}
			pidns = pidns.Parent()
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
			pm.m[namespacedpid] = namespacedpids
		}
	}
	return pm
}

// NSpid returns the list of namespaced PIDs for the process proc, based on
// information from the /proc filesystem (the "NSpid:" field in particular).
// NSpid only returns the list of PIDs, but not the corresponding PID
// namespaces; this is because the Linux kernel doesn't give us the namespace
// information as part of the process status. Instead, a caller (such as
// NewPIDMap) needs to combine a namespaced PIDs list with the hierarchy own
// PID namespaces to calculate the correct namespacing.
func NSpid(proc *model.Process) (pids []model.PIDType) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", proc.PID))
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
				pid, err := strconv.Atoi(pidtxt)
				if err != nil {
					return []model.PIDType{}
				}
				pids[idx] = model.PIDType(pid)
			}
			return
		}
	}
	panic("/proc filesystem broken: no NSpid element in status.")
}
