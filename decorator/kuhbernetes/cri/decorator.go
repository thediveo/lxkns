// Copyright 2023 Harald Albrecht.
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

package cri

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher/cri"
)

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("cri"))
}

// Decorate decorates the discovered k8s (CRI) containers with pod groups, where
// applicable.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	total := 0
	for _, engine := range engines {
		// If it "ain't no" CRI, skip it, as we're looking specifically for
		// CRI-supporting engines. Please note that we handle the deprecated
		// Docker shim in a separate decorator. Finally, the containerd
		// decorator won't touch the k8s.io and moby namespaces, so that we're
		// really responsible here.
		if engine.Type != cri.Type {
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
		}
	}
	if total > 0 {
		log.Infof("discovered %d CRI pods", total)
	}
}
