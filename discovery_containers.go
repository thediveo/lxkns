// Copyright 2021 Harald Albrecht.
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

// +build linux

package lxkns

import (
	"context"

	"github.com/thediveo/lxkns/model"
)

// discoverContainers discovers alive containers using the optionally specified
// Containerizer and then resolves the relationships between containers and
// processes (and thus also namespaces).
func discoverContainers(result *DiscoveryResult) {
	if result.Options.Containerizer() == nil {
		return
	}
	containers := result.Options.Containerizer().Containers(context.Background(), result.Processes, nil) // TODO:
	// Update the discovery information with the container found and establish
	// the links between container and process information model objects.
	result.Containers = containers
	for idx := range containers {
		// TODO: translate PID to initial PID namespace, if necessary.
		if proc, ok := result.Processes[containers[idx].PID()]; ok {
			proc.Container = containers[idx]
			containers[idx].(model.ContainerFixer).SetProcess(proc)
		}
	}

}
