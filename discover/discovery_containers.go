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

//go:build linux

package discover

import (
	"context"
	"fmt"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/plural"

	_ "github.com/thediveo/lxkns/decorator/all" // register all decorator plugins
)

// discoverContainers discovers alive containers using the optionally specified
// Containerizer (as part of Result.Options) and then resolves the relationships
// between containers and processes (and thus also namespaces). Also translates
// container PIDs for containers in containers when their container engine PIDs
// are known so that PID translation is possible.
func discoverContainers(result *Result) {
	if result.Options.Containerizer == nil {
		return
	}
	// Get the initial PID namespace so we can translate container-in-container
	// PIDs if it should become necessary.
	var initialPIDns model.Namespace
	if len(result.PIDNSRoots) == 1 {
		initialPIDns = result.PIDNSRoots[0]
	}
	containers := result.Options.Containerizer.Containers(context.Background(), result.Processes, result.PIDMap)
	// Update the discovery information with the container found and establish
	// the links between container and process information model objects. Also
	// translate container PIDs where necessary, such as in case of
	// containerized container engines (sic!).
	result.Containers = containers
	enginesPIDns := map[*model.ContainerEngine]model.Namespace{} // cache engines' PID namespaces
	pidmap := result.PIDMap                                      // might be nil
	for _, container := range containers {
		if container.Engine == nil {
			panic(fmt.Sprintf("containerizer returned container without engine: %+v", container))
		}
		enginePIDns, ok := enginesPIDns[container.Engine]
		if !ok {
			if engineProc, ok := result.Processes[container.Engine.PID]; ok {
				enginePIDns = engineProc.Namespaces[model.PIDNS]
			}
			// Cache even unsuckcessful engine PID namespace lookups.
			enginesPIDns[container.Engine] = enginePIDns
		}
		// Translate container PID from its managing container engine PID
		// namespace to initial PID namespace, if necessary.
		if pidmap != nil && enginePIDns != nil && enginePIDns != initialPIDns {
			if pid := pidmap.Translate(container.PID, enginePIDns, initialPIDns); pid != 0 {
				container.PID = pid
			}
		}
		// Relate this container with its initial process and vice versa.
		if containerProc, ok := result.Processes[container.PID]; ok {
			containerProc.Container = container
			container.Process = containerProc
		}
	}
	engines := make([]*model.ContainerEngine, 0, len(enginesPIDns))
	for engine := range enginesPIDns {
		engines = append(engines, engine)
	}
	log.Infof("discovered %s managed by %s",
		plural.Elements(len(containers), "containers"),
		plural.Elements(len(engines), "container engines"))
	// Run registered Decorators on discovered containers.
	for _, decorator := range plugger.Group[decorator.Decorate]().Symbols() {
		decorator(engines, result.Options.Labels)
	}
}
