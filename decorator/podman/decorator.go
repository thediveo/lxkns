//go:build podman

// Copyright 2022 Harald Albrecht.
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

package podman

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/plural"
	"github.com/thediveo/sealwatcher/podman"
)

// PodGroupType identifies container groups representing Podman pods.
const PodGroupType = "io.podman"

// Register this decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("podman-pods"))
}

// Decorate decorates the discovered Podman containers with Podman pod groups,
// where applicable.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	total := 0
	for _, engine := range engines {
		if engine.Type != podman.Type {
			continue
		}
		// Podman pods don't span multiple Podman engines, so we can check each
		// container engine individually.
		pods := map[string]*model.Group{}
		for _, container := range engine.Containers {
			podname, ok := container.Labels[podman.PodLabelName]
			if !ok {
				continue
			}
			// If not yet known, create a new group for each Podman pod we
			// find. Add this pod container then to the per-pod group.
			pod, ok := pods[podname]
			if !ok {
				pod = &model.Group{
					Name:   podname,
					Type:   PodGroupType,
					Flavor: PodGroupType,
				}
				pods[podname] = pod
				total++
			}
			pod.AddContainer(container)
		}
	}
	if total > 0 {
		log.Infof("discovered %s", plural.Elements(total, "Podman pods"))
	}
}
