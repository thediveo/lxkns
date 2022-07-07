// Copyright 2021 Harald Albrecht.
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

package composer

import (
	"github.com/thediveo/go-plugger/v2"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/plural"
)

// ComposerProjectLabel specifies the label (name) that identifies the project
// name for a container being part of a (Docker, nerdctl, ...) composer project.
const ComposerProjectLabel = "com.docker.compose.project"

// ComposerGroupType identifies container groups representing composer projects.
const ComposerGroupType = ComposerProjectLabel

// Register this Decorator plugin.
func init() {
	plugger.Register(
		plugger.WithName("composer"),
		plugger.WithGroup(decorator.PluginGroup),
		plugger.WithSymbol(decorator.Decorate(Decorate)))
}

// Decorate decorates the discovered Docker (and nerdctl) containers with
// composer groups, where applicable.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	total := 0
	for _, engine := range engines {
		// Projects do not span multiple container engines inside the same host,
		// so we avoid mixing them up by accident. This assumes that we do not
		// need to support multiple containerized Docker Swarm instances on a
		// single host, so no SinD (Swarm in Docker, with a nod to
		// Kubernetes-in-Docker).
		projects := map[string]*model.Group{}
		for _, container := range engine.Containers {
			projectname, ok := container.Labels[ComposerProjectLabel]
			if !ok {
				continue
			}
			// If not yet known, create a new group for each Composer project we
			// find. Add this composer container then to the per-project group.
			project, ok := projects[projectname]
			if !ok {
				project = &model.Group{
					Name:   projectname,
					Type:   ComposerGroupType,
					Flavor: ComposerGroupType,
				}
				projects[projectname] = project
				total++
			}
			project.AddContainer(container)
		}
	}
	if total > 0 {
		log.Infof("discovered %s", plural.Elements(total, "composer projects"))
	}
}
