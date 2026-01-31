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
	"log/slog"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/model"

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
			} else if container.Engine.PPIDHint != 0 {
				// This is a newly socket-activated engine that isn't yet
				// included in the process tree – that process tree that
				// ironically lead to the detection of the socket activator and
				// then activation of that container engine. As we cannot change
				// the past discovery some kind soul – a turtle, perchance? –
				// might have passed us a hint about the engine's parent process
				// PID. This parent process's PID namespace should be the same
				// as the container engine, so it should be good for container
				// PID translation.
				//
				// This deserves a badge: [COMMENTOR] ... rhymes with
				// "tormentor" *snicker*
				if parentProc, ok := result.Processes[container.Engine.PPIDHint]; ok {
					enginePIDns = parentProc.Namespaces[model.PIDNS]
				}
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
	slog.Info("discovered containers",
		slog.Int("count", len(containers)), slog.Int("engine_count", len(engines)))
	// Run registered Decorators on discovered containers.
	for _, decorator := range plugger.Group[decorator.Decorate]().Symbols() {
		decorator(engines, result.Options.Labels)
	}
}
