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

package model

// PIDMapper translates PIDs from one PID namespace to another.
type PIDMapper interface {
	// Translate translates a PID "pid" in PID namespace "from" to the
	// corresponding PID in PID namespace "to". Returns 0, if PID "pid" either
	// does not exist in namespace "from", or PID namespace "to" isn't either a
	// parent or child of PID namespace "from".
	Translate(pid PIDType, from Namespace, to Namespace) PIDType
	// NamespacedPIDs returns for a specific namespaced PID the list of all PIDs
	// the corresponding process has been given in different PID namespaces.
	// Returns nil if the PID doesn't exist in the specified PID namespace. The
	// list is ordered from the topmost PID namespace down to the leaf PID
	// namespace to which a process actually is joined to.
	NamespacedPIDs(pid PIDType, from Namespace) NamespacedPIDs
}

// NamespacedPID is a PID valid only in the context of its PID namespace.
type NamespacedPID struct {
	PIDNS Namespace // PID namespace ID for PID.
	PID   PIDType   // PID within PID namespace (of ID).
}

// NamespacedPIDs is a list of PIDs for the same single process, but in
// different PID namespaces. The order of the list is undefined.
type NamespacedPIDs []NamespacedPID

// PIDs just returns the different PIDs assigned to a single process in
// different PID namespaces, without the namespaces. This is a convenience
// function for those simple use cases where just the PID list is wanted, but no
// further PID namespace details.
func (ns NamespacedPIDs) PIDs() []PIDType {
	pids := make([]PIDType, len(ns))
	for idx, el := range ns {
		pids[idx] = el.PID
	}
	return pids
}
