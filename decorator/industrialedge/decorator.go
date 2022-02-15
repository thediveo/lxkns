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

package industrialedge

import (
	"strings"

	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/model"
)

// IndustrialEdgeAppFlavor identifies composer projects which are Industrial
// Edge apps.
const IndustrialEdgeAppFlavor = "com.siemens.industrialedge.app"

// IndustrialEdgeRuntimeFlavor identifiers the container housing the Industrial
// Edge runtime.
const IndustrialEdgeRuntimeFlavor = "com.siemens.industrialedge.runtime"

// edgeRuntimeContainerName is the name of the container housing the Siemens
// Industrial Edge runtime. We use this container name to detect the runtime in
// order to decorate its container with the IED runtime flavor.
const edgeRuntimeContainerName = "edge-iot-core"

// edgeAppConfigLabelPrefix defines the label name prefix used by the Industrial
// Edge device runtime to attach certain configuration information to containers
// of Industrial Edge apps. We use the presence of these labels to decorate
// composer-project containers as Industrial Edge app containers.
const edgeAppConfigLabelPrefix = "com_mwp_conf_"

// Register this Decorator plugin.
func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:      "industrialedge",
		Group:     decorator.PluginGroup,
		Placement: ">composer",
		Symbols: []plugger.Symbol{
			decorator.Decorate(Decorate),
		},
	})
}

// Decorate decorates the discovered Docker containers with Industrial Edge app
// project flavor, where applicable.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	for _, engine := range engines {
		// If there's an IE runtime container, then decorate it with its special
		// flavor, so UI tools might choose to display a dedicated icon for easy
		// visual identification.
		isEdge := false
		for _, container := range engine.Containers {
			if container.Name == edgeRuntimeContainerName {
				isEdge = true
				container.Flavor = IndustrialEdgeRuntimeFlavor
				break
			}
		}
		if !isEdge {
			continue
		}
		// Decorate the IE app flavor for composer projects as well as their
		// containers, where applicable. As there is no explicit indication of
		// IE apps, we use the presence of labels with certain name prefixes to
		// deduce that such containers are part of an Industrial Edge app.
		for _, container := range engine.Containers {
			isEdgeApp := false
			for labelkey := range container.Labels {
				if strings.HasPrefix(labelkey, edgeAppConfigLabelPrefix) {
					isEdgeApp = true
					break
				}
			}
			if !isEdgeApp {
				continue
			}
			if project := container.Group(composer.ComposerGroupType); project != nil {
				project.Flavor = IndustrialEdgeAppFlavor
				container.Flavor = IndustrialEdgeAppFlavor
			}
		}
	}
}
