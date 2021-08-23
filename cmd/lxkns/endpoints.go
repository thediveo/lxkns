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

package main

import (
	"encoding/json"
	"net/http"

	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/species"
)

// GetNamespacesHandler takes a containerizer and then returns a handler
// function that returns the results of a namespace discovery, as JSON.
// Additionally, we opt in to mount path+point discovery.
func GetNamespacesHandler(cizer containerizer.Containerizer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		allns := discover.Namespaces(
			discover.WithFullDiscovery(),
			discover.WithContainerizer(cizer),
			discover.WithPIDMapper(), // recommended when using WithContainerizer.
		)
		// Note bene: set header before writing the header with the status code;
		// actually makes sense, innit?
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(
			types.NewDiscoveryResult(types.WithResult(allns))) // ...brackets galore!!!
		if err != nil {
			log.Errorf("namespaces discovery error: %s", err.Error())
		}
	}
}

// GetProcessesHandler returns the process table with namespace references, as
// JSON.
func GetProcessesHandler(w http.ResponseWriter, req *http.Request) {
	disco := discover.Namespaces(discover.FromProcs())

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(
		types.NewProcessTable(types.WithProcessTable(disco.Processes)))
	if err != nil {
		log.Errorf("processes discovery error: %s", err.Error())
	}
}

// GetPIDMapHandler returns data for translating PIDs between hierarchical PID
// namespaces, as JSON.
func GetPIDMapHandler(w http.ResponseWriter, req *http.Request) {
	pidmap := discover.NewPIDMap(discover.Namespaces(discover.WithStandardDiscovery(), discover.WithNamespaceTypes(species.CLONE_NEWPID)))

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(
		types.NewPIDMap(types.WithPIDMap(pidmap)))
	if err != nil {
		log.Errorf("pid translation map discovery error: %s", err.Error())
	}
}
