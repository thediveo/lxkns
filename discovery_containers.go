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
	"fmt"

	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/plural"

	_ "github.com/thediveo/lxkns/decorator/all" // register all decorator plugins
)

// discoverContainers discovers alive containers using the optionally specified
// Containerizer and then resolves the relationships between containers and
// processes (and thus also namespaces).
func discoverContainers(result *DiscoveryResult) {
	if result.Options.Containerizer == nil {
		return
	}
	containers := result.Options.Containerizer.Containers(context.Background(), result.Processes, nil) // TODO:
	// Update the discovery information with the container found and establish
	// the links between container and process information model objects.
	result.Containers = containers
	enginesmap := map[*model.ContainerEngine]struct{}{}
	for _, container := range containers {
		if container.Engine == nil {
			panic(fmt.Sprintf("containerizer returned container without engine: %+v", container))
		}
		enginesmap[container.Engine] = struct{}{}
		// TODO: translate PID to initial PID namespace, if necessary.
		if proc, ok := result.Processes[container.PID]; ok {
			proc.Container = container
			container.Process = proc
		}
	}
	engines := make([]*model.ContainerEngine, 0, len(enginesmap))
	for engine := range enginesmap {
		engines = append(engines, engine)
	}
	log.Infof("discovered %s from %s",
		plural.Elements(len(containers), "containers"),
		plural.Elements(len(engines), "container engines"))
	// Run registered Decorators on discovered containers.
	decorators := plugger.New(decorator.PluginGroup)
	for _, decorateur := range decorators.Func("Decorate") {
		decorateur.(decorator.Decorate)(engines)
	}
}
