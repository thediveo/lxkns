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

package cricontainerd

import (
	"github.com/thediveo/go-plugger/v2"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher/containerd"
)

// containerKind specifies the kind of container at the engine level, in order
// to differentiate between user containers and infrastructure "sandbox"
// containers that haven't been specified by users (deployments).
const containerKindLabel = "io.cri-containerd.kind"

// Register this Decorator plugin.
func init() {
	plugger.Register(
		plugger.WithName("cri-containerd"),
		plugger.WithGroup(decorator.PluginGroup),
		plugger.WithSymbol(decorator.Decorate(Decorate)))
}

// Decorate decorates the discovered Docker containers with pod groups, where
// applicable.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	total := 0
	for _, engine := range engines {
		// If it "ain't no" containerd, skip it, as we're looking specifically
		// for containerd engines and their particular Kubernetes pod labelling.
		if engine.Type != containerd.Type {
			continue
		}
		// Pods cannot span container engines ;)
		podgroups := map[string]*model.Group{}
		for _, container := range engine.Containers {
			podNamespace := container.Labels[kuhbernetes.PodNamespaceLabel]
			podName := container.Labels[kuhbernetes.PodNameLabel]
			if podName == "" || podNamespace == "" {
				continue
			}
			// Create a new pod group, if it doesn't exist yet. Add the
			// container to its pod group.
			namespacedpodname := podNamespace + "/" + podName
			podgroup, ok := podgroups[namespacedpodname]
			if !ok {
				podgroup = &model.Group{
					Name:   namespacedpodname,
					Type:   kuhbernetes.PodGroupType,
					Flavor: kuhbernetes.PodGroupType,
				}
				podgroups[namespacedpodname] = podgroup
				total++
			}
			podgroup.AddContainer(container)
			// Sandbox? Then tag (label) the container.
			if container.Labels[containerKindLabel] == "sandbox" {
				container.Labels[kuhbernetes.PodSandboxLabel] = ""
			}
		}
	}
	if total > 0 {
		log.Infof("discovered %d containerd pods", total)
	}
}
